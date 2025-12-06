package plugin

import (
	"context"
)

// 示例：如何注册新的插件类型

// ExampleRegisterAlipayPlugin 示例：注册支付宝插件
func ExampleRegisterAlipayPlugin() {
	registry := GetRegistry()

	// 注册 alipay_wap 插件
	registry.Register("alipay_wap", func(ctx context.Context, pluginID int64, pluginType string) (Plugin, error) {
		// 这里可以初始化支付宝插件
		// 例如：从数据库加载配置、初始化客户端等
		return NewAlipayWapPlugin(pluginID), nil
	})

	// 注册 alipay_app 插件
	registry.Register("alipay_app", func(ctx context.Context, pluginID int64, pluginType string) (Plugin, error) {
		return NewAlipayAppPlugin(pluginID), nil
	})
}

// ExampleRegisterWechatPlugin 示例：注册微信支付插件
func ExampleRegisterWechatPlugin() {
	registry := GetRegistry()

	// 注册 wechat_wap 插件
	registry.Register("wechat_wap", func(ctx context.Context, pluginID int64, pluginType string) (Plugin, error) {
		return NewWechatWapPlugin(pluginID), nil
	})

	// 注册 wechat_app 插件
	registry.Register("wechat_app", func(ctx context.Context, pluginID int64, pluginType string) (Plugin, error) {
		return NewWechatAppPlugin(pluginID), nil
	})
}

// 这些函数需要在应用启动时调用，例如在 main.go 或初始化函数中
// func init() {
//     ExampleRegisterAlipayPlugin()
//     ExampleRegisterWechatPlugin()
// }

// 占位符函数（实际实现需要根据具体支付方式）
func NewAlipayWapPlugin(pluginID int64) Plugin {
	return &BasePlugin{pluginID: pluginID}
}

func NewAlipayAppPlugin(pluginID int64) Plugin {
	return &BasePlugin{pluginID: pluginID}
}

func NewWechatWapPlugin(pluginID int64) Plugin {
	return &BasePlugin{pluginID: pluginID}
}

func NewWechatAppPlugin(pluginID int64) Plugin {
	return &BasePlugin{pluginID: pluginID}
}
