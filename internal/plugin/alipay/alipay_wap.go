package alipay

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/plugin"
	"github.com/golang-pay-core/internal/utils"
)

// WapPlugin 支付宝手机网站支付插件
// 参考 Python: AlipayWapPluginResponder
// 与 alipay_phone 的区别：alipay_wap 返回授权URL，alipay_phone 返回直接支付URL
type WapPlugin struct {
	*BasePlugin // 嵌入 BasePlugin（支付宝基类）
}

// NewWapPlugin 创建支付宝手机网站支付插件
func NewWapPlugin(pluginID int64) *WapPlugin {
	return &WapPlugin{
		BasePlugin: NewBasePlugin(pluginID),
	}
}

// CreateOrder 创建订单
// 参考 Python: AlipayWapPluginResponder.create_order
func (p *WapPlugin) CreateOrder(ctx context.Context, req *plugin.CreateOrderRequest) (*plugin.CreateOrderResponse, error) {
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

	return plugin.NewSuccessResponse(payURL), nil
}

// getOrderInfo 获取订单详情和订单信息（公共逻辑）
func (p *WapPlugin) getOrderInfo(req *plugin.CreateOrderRequest) (*models.OrderDetail, *models.Order, error) {
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
// 参考 Python: AlipayWapPluginResponder.get_pay_url
// 与 alipay_phone 的区别：不处理重定向，直接返回支付URL，然后生成授权URL
func (p *WapPlugin) generatePayURL(ctx context.Context, req *plugin.CreateOrderRequest, productID string, order *models.Order) (string, error) {
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
	notifyURL := fmt.Sprintf("%s/api/pay/order/notify/alipay_wap/%s/", notifyDomain, productID)

	// 生成订单主题
	subject := generateSubject(ctx, req, productID)

	// 构建支付URL（不重定向）
	payURL, err := p.buildAlipayWapPayURL(req, notifyURL, subject, productID, order, false)
	if err != nil {
		return "", err
	}

	// 生成授权URL
	// 参考 Python: 生成 alipays://platformapi/startapp?appId=20000067&url=...
	domain, err := p.getDomain(req.DomainID)
	if err != nil {
		return "", fmt.Errorf("获取域名信息失败: %w", err)
	}

	// 构建授权URL
	authURL := p.buildAuthURL(domain, req.OrderNo, payURL)

	return authURL, nil
}

// buildAlipayWapPayURL 构建支付宝手机网站支付URL（不重定向）
func (p *WapPlugin) buildAlipayWapPayURL(req *plugin.CreateOrderRequest, notifyURL, subject, productID string, order *models.Order, redirects bool) (string, error) {
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
	alipayClient, err := createAlipayClient(&product, notifyURL)
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
	payURL, err := alipayClient.TradeWapPay(subject, req.OrderNo, totalAmount, notifyURL, others)
	if err != nil {
		return "", fmt.Errorf("生成支付URL失败: %w", err)
	}

	// 如果需要重定向，获取重定向后的URL
	if redirects {
		redirectURL, err := alipayClient.GetRedirectURL(payURL)
		if err != nil {
			return payURL, nil // 如果重定向失败，返回原始URL
		}
		return redirectURL, nil
	}

	return payURL, nil
}

// getDomain 获取域名信息
func (p *WapPlugin) getDomain(domainID *int64) (*models.PayDomain, error) {
	if domainID == nil {
		return nil, fmt.Errorf("域名ID不能为空")
	}

	var domain models.PayDomain
	if err := database.DB.Where("id = ?", *domainID).First(&domain).Error; err != nil {
		return nil, fmt.Errorf("域名不存在: %w", err)
	}

	return &domain, nil
}

// buildAuthURL 构建授权URL
// 参考 Python: alipays://platformapi/startapp?appId=20000067&url=...
func (p *WapPlugin) buildAuthURL(domain *models.PayDomain, orderNo, payURL string) string {
	// 解析域名URL
	domainURL, err := url.Parse(domain.URL)
	if err != nil {
		return payURL // 如果解析失败，返回原始支付URL
	}

	host := fmt.Sprintf("%s://%s", domainURL.Scheme, domainURL.Host)

	// 构建授权回调URL
	authCallbackURL := fmt.Sprintf("%s/api/pay/order/alipay_wap/%s", host, orderNo)

	// 如果有 auth_key，添加鉴权
	if domain.AuthKey != "" {
		// 使用 buildAuthURL 生成带鉴权的URL
		authKey := utils.GetAuthKey(orderNo, domain.AuthKey, 30)
		authCallbackURL = buildAuthURL(authCallbackURL, authKey, domain.AuthTimeout)
	}

	// URL编码
	encodedURL := url.QueryEscape(authCallbackURL)

	// 构建支付宝授权URL
	authURL := fmt.Sprintf("alipays://platformapi/startapp?appId=20000067&url=%s",
		url.QueryEscape(fmt.Sprintf("https://openauth.alipay.com/oauth2/publicAppAuthorize.htm?app_id=%s&scope=auth_base&redirect_uri=%s",
			domain.AppID, encodedURL)))

	return authURL
}

// 实现 PluginCapabilities 接口
var _ plugin.PluginCapabilities = (*WapPlugin)(nil)

// CanHandleExtra 是否可以处理额外参数
func (p *WapPlugin) CanHandleExtra() bool {
	return false
}

// AutoExtra 是否自动处理额外参数
func (p *WapPlugin) AutoExtra() bool {
	return false
}

// ExtraNeedProduct 额外参数是否需要产品
func (p *WapPlugin) ExtraNeedProduct() bool {
	return false
}

// ExtraNeedCookie 额外参数是否需要Cookie
func (p *WapPlugin) ExtraNeedCookie() bool {
	return false
}

// GetTimeout 获取订单超时时间（秒）
func (p *WapPlugin) GetTimeout(ctx context.Context, pluginID int64) int {
	return 300
}
