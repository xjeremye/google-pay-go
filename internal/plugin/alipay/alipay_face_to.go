package alipay

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/plugin"
	"github.com/golang-pay-core/internal/utils"
)

// FaceToPlugin 支付宝当面付插件
// 参考 Python: AlipayFacePluginResponder
type FaceToPlugin struct {
	*BasePlugin // 嵌入 BasePlugin（支付宝基类）
}

// NewFaceToPlugin 创建支付宝当面付插件
func NewFaceToPlugin(pluginID int64) *FaceToPlugin {
	return &FaceToPlugin{
		BasePlugin: NewBasePlugin(pluginID),
	}
}

// CreateOrder 创建订单
// 参考 Python: AlipayFacePluginResponder.create_order
func (p *FaceToPlugin) CreateOrder(ctx context.Context, req *plugin.CreateOrderRequest) (*plugin.CreateOrderResponse, error) {
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

	// 生成支付URL（授权URL）
	payURL, err := p.generatePayURL(ctx, req, productID, order)
	if err != nil {
		return plugin.NewErrorResponse(7320, fmt.Sprintf("生成支付URL失败: %v", err)), nil
	}

	return plugin.NewSuccessResponse(payURL), nil
}

// getOrderInfo 获取订单详情和订单信息（公共逻辑）
func (p *FaceToPlugin) getOrderInfo(req *plugin.CreateOrderRequest) (*models.OrderDetail, *models.Order, error) {
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

// generatePayURL 生成支付URL（授权URL）
// 参考 Python: AlipayFacePluginResponder.get_pay_url
func (p *FaceToPlugin) generatePayURL(ctx context.Context, req *plugin.CreateOrderRequest, productID string, order *models.Order) (string, error) {
	// 获取域名信息
	if req.DomainID == nil {
		return "", fmt.Errorf("域名ID不能为空")
	}

	var domain models.PayDomain
	if err := database.DB.Where("id = ?", *req.DomainID).First(&domain).Error; err != nil {
		return "", fmt.Errorf("域名不存在: %w", err)
	}

	// 解析域名URL
	domainURL, err := url.Parse(domain.URL)
	if err != nil {
		return "", fmt.Errorf("域名URL格式错误: %w", err)
	}

	host := fmt.Sprintf("%s://%s", domainURL.Scheme, domainURL.Host)

	// 构建授权回调URL
	// Python: url = f"{host}/api/pay/order/{self._face_type}/{raw_order_no}"
	// _face_type = "face_pay"
	authCallbackURL := fmt.Sprintf("%s/api/pay/order/face_pay/%s", host, req.OutOrderNo)

	// 添加鉴权（如果有 auth_key）
	if domain.AuthKey != "" {
		// 使用 buildAuthURL 生成带鉴权的URL
		authKey := utils.GetAuthKey(req.OutOrderNo, domain.AuthKey, 30)
		authCallbackURL = buildAuthURL(authCallbackURL, authKey, domain.AuthTimeout)
	}

	// URL编码（对整个URL进行编码）
	encodedURL := url.QueryEscape(authCallbackURL)

	// 构建支付宝授权URL
	// Python: "alipays://platformapi/startapp?appId=20000067&url=" + urllib.parse.quote(...)
	authURL := fmt.Sprintf("alipays://platformapi/startapp?appId=20000067&url=%s",
		url.QueryEscape(fmt.Sprintf("https://openauth.alipay.com/oauth2/publicAppAuthorize.htm?app_id=%s&scope=auth_base&redirect_uri=%s",
			domain.AppID, encodedURL)))

	return authURL, nil
}

// 实现 PluginCapabilities 接口
var _ plugin.PluginCapabilities = (*FaceToPlugin)(nil)

// CanHandleExtra 是否可以处理额外参数
func (p *FaceToPlugin) CanHandleExtra() bool {
	return false
}

// AutoExtra 是否自动处理额外参数
func (p *FaceToPlugin) AutoExtra() bool {
	return false
}

// ExtraNeedProduct 额外参数是否需要产品
func (p *FaceToPlugin) ExtraNeedProduct() bool {
	return false
}

// ExtraNeedCookie 额外参数是否需要Cookie
func (p *FaceToPlugin) ExtraNeedCookie() bool {
	return false
}

// GetTimeout 获取订单超时时间（秒）
func (p *FaceToPlugin) GetTimeout(ctx context.Context, pluginID int64) int {
	return 300
}
