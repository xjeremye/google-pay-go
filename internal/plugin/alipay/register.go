package alipay

import (
	"context"

	"github.com/golang-pay-core/internal/plugin"
)

// RegisterPhonePlugin 注册支付宝手机网站支付插件
func RegisterPhonePlugin() {
	registry := plugin.GetRegistry()

	// 注册 alipay_phone 插件
	registry.Register("alipay_phone", func(ctx context.Context, pluginID int64, pluginType string) (plugin.Plugin, error) {
		return NewPhonePlugin(pluginID), nil
	})
}

// init 自动注册支付宝插件
// 在包导入时自动执行，无需在 main.go 中手动调用
// 使用 init() 函数可以保持 main.go 的简洁性
// 注意：init() 函数在 logger 初始化之前执行，因此不在此处记录日志
func init() {
	RegisterPhonePlugin()
	// 未来可以在这里添加其他支付宝插件的自动注册：
	// RegisterWapPlugin()
	// RegisterAppPlugin()
	// RegisterQrPlugin()
}
