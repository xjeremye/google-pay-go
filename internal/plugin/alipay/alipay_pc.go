package alipay

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golang-pay-core/internal/alipay"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/plugin"
)

// PcPlugin 支付宝电脑网站支付插件
// 参考 Python: AlipayPcPluginResponder
type PcPlugin struct {
	*BasePlugin // 嵌入 BasePlugin（支付宝基类）
}

// NewPcPlugin 创建支付宝电脑网站支付插件
func NewPcPlugin(pluginID int64) *PcPlugin {
	return &PcPlugin{
		BasePlugin: NewBasePlugin(pluginID),
	}
}

// CreateOrder 创建订单
// 参考 Python: AlipayPcPluginResponder.create_order
func (p *PcPlugin) CreateOrder(ctx context.Context, req *plugin.CreateOrderRequest) (*plugin.CreateOrderResponse, error) {
	// 获取订单详情和订单信息
	orderDetail, order, err := p.getOrderInfo(req)
	if err != nil {
		return plugin.NewErrorResponse(7320, err.Error()), nil
	}

	// 获取产品ID
	productID := req.ProductID
	if productID == "" {
		productID = orderDetail.ProductID
	}
	if productID == "" {
		return plugin.NewErrorResponse(7320, "产品ID不能为空"), nil
	}

	// 获取域名信息
	if req.DomainID == nil {
		if orderDetail.DomainID != nil {
			req.DomainID = orderDetail.DomainID
		} else {
			return plugin.NewErrorResponse(7320, "域名ID不能为空"), nil
		}
	}

	// 生成支付URL
	payURL, err := p.generatePayURL(ctx, req, productID, order)
	if err != nil {
		return plugin.NewErrorResponse(7320, fmt.Sprintf("生成支付URL失败: %v", err)), nil
	}

	// PC支付返回原始URL（is_raw=True）
	response := plugin.NewSuccessResponse(payURL)
	response.ExtraData = map[string]interface{}{
		"is_raw": true, // 标记为原始URL，前端直接跳转
	}

	return response, nil
}

// getOrderInfo 获取订单详情和订单信息（公共逻辑）
func (p *PcPlugin) getOrderInfo(req *plugin.CreateOrderRequest) (*models.OrderDetail, *models.Order, error) {
	var orderDetail models.OrderDetail
	if req.DetailID > 0 {
		if err := database.DB.Where("id = ?", req.DetailID).First(&orderDetail).Error; err != nil {
			return nil, nil, fmt.Errorf("订单详情不存在")
		}
	} else if req.OrderID != "" {
		if err := database.DB.Where("order_id = ?", req.OrderID).First(&orderDetail).Error; err != nil {
			return nil, nil, fmt.Errorf("订单详情不存在")
		}
	} else {
		return nil, nil, fmt.Errorf("缺少订单ID或详情ID")
	}

	var order models.Order
	if req.OrderID != "" {
		if err := database.DB.Where("id = ?", req.OrderID).First(&order).Error; err != nil {
			return nil, nil, fmt.Errorf("订单不存在")
		}
	} else if req.OrderNo != "" {
		if err := database.DB.Where("order_no = ?", req.OrderNo).First(&order).Error; err != nil {
			return nil, nil, fmt.Errorf("订单不存在")
		}
	} else {
		return nil, nil, fmt.Errorf("缺少订单ID或订单号")
	}

	return &orderDetail, &order, nil
}

// generatePayURL 生成支付URL
// 参考 Python: AlipayPcPluginResponder.get_pay_url
func (p *PcPlugin) generatePayURL(ctx context.Context, req *plugin.CreateOrderRequest, productID string, order *models.Order) (string, error) {
	// 获取域名信息
	domainURL := req.DomainURL
	if domainURL == "" {
		return "", fmt.Errorf("域名URL不能为空")
	}

	// 获取通知域名
	notifyDomain := getSystemConfigByPath(ctx, "alipay.notify_domain")
	if notifyDomain == "" {
		notifyDomain = domainURL
	}

	// 构建通知URL
	notifyURL := fmt.Sprintf("%s/api/pay/order/notify/alipay_pc/%s/", notifyDomain, productID)

	// 生成订单主题
	subject := generateSubject(ctx, req, productID)

	// 构建支付URL
	payURL, err := p.buildAlipayPcPayURL(req, notifyURL, subject, productID)
	if err != nil {
		return "", err
	}

	return payURL, nil
}

// buildAlipayPcPayURL 构建支付宝PC网站支付URL
func (p *PcPlugin) buildAlipayPcPayURL(req *plugin.CreateOrderRequest, notifyURL, subject, productID string) (string, error) {
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
	var alipayClient *alipay.Client
	alipayClient, err = createAlipayClient(&product, notifyURL)
	if err != nil {
		return "", fmt.Errorf("创建支付宝客户端失败: %w", err)
	}

	// 设置订单信息
	alipayClient.OrderNo = req.OrderNo
	alipayClient.OutOrderNo = req.OutOrderNo

	// 格式化金额
	totalAmount := fmt.Sprintf("%.2f", float64(req.Money)/100)

	// 构建其他参数
	others := buildAlipayPayParams(&product, req, totalAmount)

	// 调用支付宝 SDK 生成支付 URL
	payURL, err := alipayClient.TradePagePay(subject, req.OrderNo, totalAmount, notifyURL, others)
	if err != nil {
		return "", fmt.Errorf("生成支付URL失败: %w", err)
	}

	return payURL, nil
}

// 实现 PluginCapabilities 接口
var _ plugin.PluginCapabilities = (*PcPlugin)(nil)

// CanHandleExtra 是否可以处理额外参数
func (p *PcPlugin) CanHandleExtra() bool {
	return false
}

// AutoExtra 是否自动处理额外参数
func (p *PcPlugin) AutoExtra() bool {
	return false
}

// ExtraNeedProduct 额外参数是否需要产品
func (p *PcPlugin) ExtraNeedProduct() bool {
	return false
}

// ExtraNeedCookie 额外参数是否需要Cookie
func (p *PcPlugin) ExtraNeedCookie() bool {
	return false
}

// GetTimeout 获取订单超时时间（秒）
func (p *PcPlugin) GetTimeout(ctx context.Context, pluginID int64) int {
	return 300
}
