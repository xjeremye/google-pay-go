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

// DdmPlugin 支付宝扫码支付插件
// 参考 Python: AlipayQrPluginResponder
type DdmPlugin struct {
	*BasePlugin // 嵌入 BasePlugin（支付宝基类）
}

// NewDdmPlugin 创建支付宝扫码支付插件
func NewDdmPlugin(pluginID int64) *DdmPlugin {
	return &DdmPlugin{
		BasePlugin: NewBasePlugin(pluginID),
	}
}

// CreateOrder 创建订单
// 参考 Python: AlipayQrPluginResponder.create_order
func (p *DdmPlugin) CreateOrder(ctx context.Context, req *plugin.CreateOrderRequest) (*plugin.CreateOrderResponse, error) {
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

	// 生成支付URL（二维码）
	payURL, err := p.generatePayURL(ctx, req, productID, order)
	if err != nil {
		return plugin.NewErrorResponse(7320, fmt.Sprintf("生成支付URL失败: %v", err)), nil
	}

	return plugin.NewSuccessResponse(payURL), nil
}

// getOrderInfo 获取订单详情和订单信息（公共逻辑）
func (p *DdmPlugin) getOrderInfo(req *plugin.CreateOrderRequest) (*models.OrderDetail, *models.Order, error) {
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

// generatePayURL 生成支付URL（二维码）
// 参考 Python: AlipayQrPluginResponder.get_pay_url
func (p *DdmPlugin) generatePayURL(ctx context.Context, req *plugin.CreateOrderRequest, productID string, order *models.Order) (string, error) {
	// 获取通知域名
	// Python: host = cache.get(f"system_config.notify.host")
	notifyDomain := getSystemConfigByPath(ctx, "notify.host")
	if notifyDomain == "" {
		// 如果没有配置，使用域名URL
		notifyDomain = req.DomainURL
	}

	// 构建通知URL
	notifyURL := fmt.Sprintf("%s/api/pay/order/notify/alipay_ddm/%s/", notifyDomain, productID)

	// 生成订单主题
	subject := generateSubject(ctx, req, productID)

	// 构建支付URL（二维码）
	qrCode, err := p.buildAlipayQrPayURL(req, notifyURL, subject, productID)
	if err != nil {
		return "", err
	}

	return qrCode, nil
}

// buildAlipayQrPayURL 构建支付宝扫码支付URL（返回二维码内容）
func (p *DdmPlugin) buildAlipayQrPayURL(req *plugin.CreateOrderRequest, notifyURL, subject, productID string) (string, error) {
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

	// 构建其他参数（扫码支付只需要 seller_id，不需要分账等复杂参数）
	others := make(map[string]interface{})
	if product.AccountType == 0 || product.AccountType == 7 {
		others["seller_id"] = product.UID
	}

	// 调用支付宝 SDK 生成二维码
	qrCode, err := alipayClient.TradePrecreate(subject, req.OrderNo, totalAmount, notifyURL, others)
	if err != nil {
		return "", fmt.Errorf("生成二维码失败: %w", err)
	}

	return qrCode, nil
}

// 实现 PluginCapabilities 接口
var _ plugin.PluginCapabilities = (*DdmPlugin)(nil)

// CanHandleExtra 是否可以处理额外参数
func (p *DdmPlugin) CanHandleExtra() bool {
	return false
}

// AutoExtra 是否自动处理额外参数
func (p *DdmPlugin) AutoExtra() bool {
	return false
}

// ExtraNeedProduct 额外参数是否需要产品
func (p *DdmPlugin) ExtraNeedProduct() bool {
	return false
}

// ExtraNeedCookie 额外参数是否需要Cookie
func (p *DdmPlugin) ExtraNeedCookie() bool {
	return false
}

// GetTimeout 获取订单超时时间（秒）
func (p *DdmPlugin) GetTimeout(ctx context.Context, pluginID int64) int {
	return 300
}
