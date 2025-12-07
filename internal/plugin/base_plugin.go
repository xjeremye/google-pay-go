package plugin

import (
	"context"
	"fmt"
)

// BasePlugin 基础插件实现（示例）
type BasePlugin struct {
	pluginID int64
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(pluginID int64) *BasePlugin {
	return &BasePlugin{
		pluginID: pluginID,
	}
}

// CreateOrder 创建订单
func (p *BasePlugin) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// 基础实现：生成一个占位符支付URL
	// 实际实现应该根据插件类型调用对应的支付接口

	payURL := fmt.Sprintf("https://pay.example.com/pay?order_no=%s&plugin_id=%d", req.OrderNo, req.PluginID)

	return NewSuccessResponse(payURL), nil
}

// 实现 PluginCapabilities 接口（可选）
var _ PluginCapabilities = (*BasePlugin)(nil)

// CanHandleExtra 是否可以处理额外参数
func (p *BasePlugin) CanHandleExtra() bool {
	return false
}

// AutoExtra 是否自动处理额外参数
func (p *BasePlugin) AutoExtra() bool {
	return false
}

// ExtraNeedProduct 额外参数是否需要产品
func (p *BasePlugin) ExtraNeedProduct() bool {
	return false
}

// ExtraNeedCookie 额外参数是否需要Cookie
func (p *BasePlugin) ExtraNeedCookie() bool {
	return false
}

// GetTimeout 获取订单超时时间（秒）
func (p *BasePlugin) GetTimeout(ctx context.Context, pluginID int64) int {
	// 默认5分钟
	return 300
}

// WaitProduct 等待产品（基础实现）
// 参考 Python: BasePluginResponder.wait_product
// 通用实现：获取支付宝产品（适用于继承 BasePlugin 的支付宝插件）
// 如果插件需要自定义逻辑，可以覆盖此方法
func (p *BasePlugin) WaitProduct(ctx context.Context, req *WaitProductRequest) (*WaitProductResponse, error) {
	// 获取可用的核销ID列表
	writeoffIDs, err := getWriteoffIDsForPlugin(req.TenantID, req.Money, &req.ChannelID)
	if err != nil {
		return NewWaitProductErrorResponse(7318, fmt.Sprintf("获取核销ID失败: %v", err)), nil
	}
	if len(writeoffIDs) == 0 {
		return NewWaitProductErrorResponse(7318, "没有可选核销"), nil
	}

	// 获取产品（通用实现：支付宝产品）
	productID, writeoffID, money, err := getAlipayProduct(ctx, req, writeoffIDs)
	if err != nil {
		return NewWaitProductErrorResponse(7318, fmt.Sprintf("获取产品失败: %v", err)), nil
	}
	if productID == "" {
		return NewWaitProductErrorResponse(7318, "无货物库存"), nil
	}
	if writeoffID == nil {
		return NewWaitProductErrorResponse(7318, "无核销库存"), nil
	}
	return NewWaitProductSuccessResponse(productID, writeoffID, "", money), nil
}
