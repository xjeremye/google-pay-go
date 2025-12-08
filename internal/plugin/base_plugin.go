package plugin

import (
	"context"
	"fmt"
)

// BasePlugin 基础插件实现（所有插件的基类）
// 提供通用的插件功能，不包含任何第三方支付平台特定的逻辑
// 第三方支付平台（支付宝、微信、京东等）的插件应该继承或嵌入此基类
type BasePlugin struct {
	pluginID int64
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(pluginID int64) *BasePlugin {
	return &BasePlugin{
		pluginID: pluginID,
	}
}

// GetPluginID 获取插件ID
func (p *BasePlugin) GetPluginID() int64 {
	return p.pluginID
}

// CreateOrder 创建订单（基础实现）
// 子类应该覆盖此方法以实现具体的支付逻辑
func (p *BasePlugin) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// 基础实现：生成一个占位符支付URL
	// 实际实现应该根据插件类型调用对应的支付接口
	payURL := fmt.Sprintf("https://pay.example.com/pay?order_no=%s&plugin_id=%d", req.OrderNo, req.PluginID)
	return NewSuccessResponse(payURL), nil
}

// WaitProduct 等待产品（基础实现）
// 子类应该覆盖此方法以实现具体的产品选择逻辑
func (p *BasePlugin) WaitProduct(ctx context.Context, req *WaitProductRequest) (*WaitProductResponse, error) {
	// 基础实现：返回错误，提示子类需要实现
	return NewWaitProductErrorResponse(7318, "WaitProduct 方法需要由子类实现"), nil
}

// CallbackSubmit 下单回调（基础实现）
// 子类应该覆盖此方法以实现具体的回调逻辑
func (p *BasePlugin) CallbackSubmit(ctx context.Context, req *CallbackSubmitRequest) error {
	// 基础实现：什么都不做
	// 子类可以覆盖此方法以实现统计更新等逻辑
	return nil
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
