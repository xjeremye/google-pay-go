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
