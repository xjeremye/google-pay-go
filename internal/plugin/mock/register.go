package mock

import (
	"context"

	"github.com/golang-pay-core/internal/plugin"
)

// RegisterMockPlugin 注册模拟测试插件
func RegisterMockPlugin() {
	registry := plugin.GetRegistry()

	// 注册 mock 插件（用于压测）
	registry.Register("mock", func(ctx context.Context, pluginID int64, pluginType string) (plugin.Plugin, error) {
		return NewMockPlugin(pluginID), nil
	})

	// 也可以注册为 mock_alipay，用于替换支付宝插件
	registry.Register("mock_alipay", func(ctx context.Context, pluginID int64, pluginType string) (plugin.Plugin, error) {
		return NewMockPlugin(pluginID), nil
	})
}

// init 自动注册模拟插件
// 在包导入时自动执行
func init() {
	RegisterMockPlugin()
}
