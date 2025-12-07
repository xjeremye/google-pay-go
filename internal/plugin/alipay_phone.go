package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang-pay-core/internal/alipay"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
)

// AlipayPhonePlugin 支付宝手机网站支付插件
// 参考 Python 的 AlipayPhonePluginResponder，继承自 AlipayWapPluginResponder
type AlipayPhonePlugin struct {
	*BasePlugin // 嵌入 BasePlugin，可以直接使用基类的方法
}

// NewAlipayPhonePlugin 创建支付宝手机网站支付插件
func NewAlipayPhonePlugin(pluginID int64) *AlipayPhonePlugin {
	return &AlipayPhonePlugin{
		BasePlugin: NewBasePlugin(pluginID),
	}
}

// CreateOrder 创建订单
// 参考 Python: AlipayPhonePluginResponder.create_order
// Python 方法签名: create_order(self, raw_order_no: str, order_no: str, out_order_no: str, money: int,
//
//	product_id: int, plugin_id: int, order_id: int, **kwargs) -> dict:
func (p *AlipayPhonePlugin) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// 从请求中获取 product_id（如果已提供）
	productID := req.ProductID

	// 获取订单详情（用于后续使用）
	var orderDetail models.OrderDetail
	if req.DetailID > 0 {
		// 使用 DetailID 直接查询
		if err := database.DB.Where("id = ?", req.DetailID).First(&orderDetail).Error; err != nil {
			return NewErrorResponse(7320, "订单详情不存在"), nil
		}
	} else if req.OrderID != "" {
		// 使用 OrderID（订单主键）查询订单详情
		if err := database.DB.Where("order_id = ?", req.OrderID).First(&orderDetail).Error; err != nil {
			return NewErrorResponse(7320, "订单详情不存在"), nil
		}
	} else {
		return NewErrorResponse(7320, "缺少订单ID或详情ID"), nil
	}

	// 如果没有提供 product_id，从订单详情中获取
	if productID == "" {
		productID = orderDetail.ProductID
	}

	// 如果没有 product_id，返回错误
	if productID == "" {
		return NewErrorResponse(7320, "产品ID不能为空"), nil
	}

	// 获取订单信息
	var order models.Order
	if req.OrderID != "" {
		if err := database.DB.Where("id = ?", req.OrderID).First(&order).Error; err != nil {
			return NewErrorResponse(7320, "订单不存在"), nil
		}
	} else if req.OrderNo != "" {
		if err := database.DB.Where("order_no = ?", req.OrderNo).First(&order).Error; err != nil {
			return NewErrorResponse(7320, "订单不存在"), nil
		}
	} else {
		return NewErrorResponse(7320, "缺少订单ID或订单号"), nil
	}

	// 获取域名信息
	// 参考 Python: domain_id 通过 kwargs.get("domain_id") 获取
	// 在 Go 中，domain_id 应该在订单创建时已经设置好了
	// 如果没有设置，尝试从订单详情中获取
	if req.DomainID == nil {
		// 从订单详情中获取 domain_id
		if orderDetail.DomainID != nil {
			req.DomainID = orderDetail.DomainID
		} else {
			return NewErrorResponse(7320, "域名ID不能为空"), nil
		}
	}

	// 生成支付URL
	payURL, err := p.generatePayURL(ctx, req, productID, &order)
	if err != nil {
		return NewErrorResponse(7320, fmt.Sprintf("生成支付URL失败: %v", err)), nil
	}

	return NewSuccessResponse(payURL), nil
}

// generatePayURL 生成支付URL
// 参考 Python: AlipayPhonePluginResponder.create_order 和 AlipayWapPluginResponder.get_pay_url
func (p *AlipayPhonePlugin) generatePayURL(ctx context.Context, req *CreateOrderRequest, productID string, order *models.Order) (string, error) {
	// 获取域名信息（从 Channel 或 Domain 中获取）
	domainURL := req.DomainURL
	if domainURL == "" {
		return "", fmt.Errorf("域名URL不能为空")
	}

	// 获取通知域名（从系统配置获取，如果没有则使用域名URL）
	// Python: host = cache.get("system_config.alipay.inline_notify_domain")
	notifyDomain := getSystemConfigByPath(ctx, "alipay.inline_notify_domain")
	if notifyDomain == "" {
		// 如果系统配置中没有，使用域名URL
		notifyDomain = domainURL
	}

	// 构建通知URL
	// Python: notify_url=f"{host}/api/pay/order/notify/{self._key}/{product_id}/"
	notifyURL := fmt.Sprintf("%s/api/pay/order/notify/alipay_phone/%s/", notifyDomain, productID)

	// 生成订单主题
	subject := p.generateSubject(req)

	// 构建支付URL
	// Python: 调用 self.get_pay_url 方法，传入 redirects=True
	// 这里调用支付宝SDK生成支付URL
	payURL, err := p.buildAlipayWapPayURL(req, domainURL, notifyURL, subject, productID, order)
	if err != nil {
		return "", err
	}

	return payURL, nil
}

// generateSubject 生成订单主题
// 参考 Python: 从 product.subject 或 get_plugin_subject 获取
func (p *AlipayPhonePlugin) generateSubject(req *CreateOrderRequest) string {
	// 从产品表获取 subject
	// Python: product.subject.format(money=format_money(money), order_no=order_no, out_order_no=out_order_no)
	if req.ProductID != "" {
		productIDInt, err := strconv.ParseInt(req.ProductID, 10, 64)
		if err == nil {
			var product models.AlipayProduct
			if err := database.DB.Select("subject").Where("id = ?", productIDInt).First(&product).Error; err == nil {
				if product.Subject != "" {
					// 格式化主题：替换占位符
					// Python: product.subject.format(money=format_money(money), order_no=order_no, out_order_no=out_order_no)
					moneyStr := fmt.Sprintf("%.2f", float64(req.Money)/100)
					subject := product.Subject
					subject = strings.ReplaceAll(subject, "{money}", moneyStr)
					subject = strings.ReplaceAll(subject, "{order_no}", req.OrderNo)
					subject = strings.ReplaceAll(subject, "{out_order_no}", req.OutOrderNo)
					// 支持 Python 风格的格式化
					subject = strings.ReplaceAll(subject, "{{money}}", moneyStr)
					subject = strings.ReplaceAll(subject, "{{order_no}}", req.OrderNo)
					subject = strings.ReplaceAll(subject, "{{out_order_no}}", req.OutOrderNo)
					if subject != product.Subject {
						return subject
					}
				}
			}
		}
	}

	// 默认主题
	moneyStr := fmt.Sprintf("%.2f", float64(req.Money)/100)
	return fmt.Sprintf("订单支付-%s-%s元", req.OrderNo, moneyStr)
}

// buildAlipayWapPayURL 构建支付宝手机网站支付URL
// 参考 Python: AlipayWapPluginResponder.get_pay_url
// 需要调用支付宝SDK的 api_alipay_trade_wap_pay 方法，并处理重定向（redirects=True）
func (p *AlipayPhonePlugin) buildAlipayWapPayURL(req *CreateOrderRequest, domainURL, notifyURL, subject, productID string, order *models.Order) (string, error) {
	// 解析产品ID
	productIDInt, err := strconv.ParseInt(productID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("产品ID格式错误: %w", err)
	}

	var product models.AlipayProduct
	if err := database.DB.Where("id = ?", productIDInt).First(&product).Error; err != nil {
		return "", fmt.Errorf("产品不存在: %w", err)
	}

	// 创建支付宝客户端
	alipayClient, err := alipay.NewClient(&product, notifyURL, true)
	if err != nil {
		return "", fmt.Errorf("创建支付宝客户端失败: %w", err)
	}

	// 格式化金额（分转元）
	totalAmount := fmt.Sprintf("%.2f", float64(req.Money)/100)

	// 构建其他参数
	others := make(map[string]interface{})

	// 根据产品账户类型添加参数
	if product.AccountType == 0 || product.AccountType == 7 {
		others["seller_id"] = product.UID
	} else if product.AccountType == 6 {
		// 分账模式
		others["settle_info"] = map[string]interface{}{
			"settle_detail_infos": []map[string]interface{}{
				{
					"amount":        totalAmount,
					"trans_in_type": "defaultSettle",
				},
			},
		}
		others["sub_merchant"] = map[string]interface{}{
			"merchant_id": product.AppID,
		}
	}

	// 获取通道的 extra_arg（B2B 模式）
	// 从 Channel 获取 extra_arg
	if req.Channel != nil {
		if extraArg, exists := req.Channel["extra_arg"]; exists {
			// 处理不同类型的 extra_arg
			var extraArgInt int
			switch v := extraArg.(type) {
			case int:
				extraArgInt = v
			case int64:
				extraArgInt = int(v)
			case float64:
				extraArgInt = int(v)
			case *int:
				if v != nil {
					extraArgInt = *v
				}
			}

			if extraArgInt == 3 {
				// B2B 模式
				others["extend_params"] = map[string]interface{}{
					"paySolution":       "E_PAY",
					"paySolutionConfig": "{\"paySolutionScene\":\"ENTERPRISE_PAY\"}",
				}
			}
		}
	}

	// 调用支付宝 SDK 生成支付 URL
	payURL, err := alipayClient.TradeWapPay(subject, req.OrderNo, totalAmount, notifyURL, others)
	if err != nil {
		return "", fmt.Errorf("生成支付URL失败: %w", err)
	}

	// 处理重定向（redirects=True）
	// 参考 Python: 发送 GET 请求获取重定向后的 URL
	redirectURL, err := alipayClient.GetRedirectURL(payURL)
	if err != nil {
		// 如果重定向失败，返回原始 URL
		return payURL, nil
	}

	return redirectURL, nil
}

// 实现 PluginCapabilities 接口
var _ PluginCapabilities = (*AlipayPhonePlugin)(nil)

// CanHandleExtra 是否可以处理额外参数
func (p *AlipayPhonePlugin) CanHandleExtra() bool {
	return false
}

// AutoExtra 是否自动处理额外参数
func (p *AlipayPhonePlugin) AutoExtra() bool {
	return false
}

// ExtraNeedProduct 额外参数是否需要产品
func (p *AlipayPhonePlugin) ExtraNeedProduct() bool {
	return false
}

// ExtraNeedCookie 额外参数是否需要Cookie
func (p *AlipayPhonePlugin) ExtraNeedCookie() bool {
	return false
}

// GetTimeout 获取订单超时时间（秒）
func (p *AlipayPhonePlugin) GetTimeout(ctx context.Context, pluginID int64) int {
	// 默认5分钟
	return 300
}

// WaitProduct 等待产品（获取产品ID、核销ID等）
// 参考 Python: BasePluginResponder.wait_product
// alipay_phone 插件嵌入 BasePlugin，直接使用基类的通用实现
// 如果需要自定义逻辑，可以覆盖此方法
// 当前实现：不覆盖，通过嵌入直接使用基类方法

// getSystemConfigByPath 通过路径获取系统配置（避免循环依赖）
// path 格式：如 "alipay.inline_notify_domain"
func getSystemConfigByPath(ctx context.Context, path string) string {
	// 先尝试直接获取
	var config models.SystemConfig
	if err := database.DB.Where("key = ? AND status = ? AND parent_id IS NULL", path, true).
		First(&config).Error; err == nil {
		return parseSystemConfigValue(config.Value)
	}

	// 如果直接获取失败，尝试按点分割路径
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return ""
	}

	// 先找父配置
	var parentConfig models.SystemConfig
	if err := database.DB.Where("key = ? AND status = ? AND parent_id IS NULL", parts[0], true).
		First(&parentConfig).Error; err != nil {
		return ""
	}

	// 再找子配置
	if err := database.DB.Where("key = ? AND status = ? AND parent_id = ?", parts[1], true, parentConfig.ID).
		First(&config).Error; err != nil {
		return ""
	}

	return parseSystemConfigValue(config.Value)
}

// parseSystemConfigValue 解析系统配置的 JSON 值
func parseSystemConfigValue(valueStr string) string {
	if valueStr == "" {
		return ""
	}

	// 尝试解析 JSON
	var valueMap map[string]interface{}
	if err := json.Unmarshal([]byte(valueStr), &valueMap); err != nil {
		// 如果解析失败，尝试直接返回原始值
		return valueStr
	}

	// 尝试获取 value 字段
	if value, ok := valueMap["value"].(string); ok {
		return value
	}

	// 如果 value 不是字符串，返回空
	return ""
}
