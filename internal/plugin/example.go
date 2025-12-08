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

// NewAlipayWapPlugin 创建支付宝手机网站支付插件（示例）
func NewAlipayWapPlugin(pluginID int64) Plugin {
	// 实际应该返回 AlipayWapPlugin 实例
	// 这里使用 BasePlugin 作为示例（应该使用 alipay.NewBasePlugin）
	// 注意：实际实现应该在 internal/plugin/alipay 包中
	return NewBasePlugin(pluginID)
}

// NewAlipayAppPlugin 创建支付宝APP支付插件（示例）
func NewAlipayAppPlugin(pluginID int64) Plugin {
	// 实际应该返回 AlipayAppPlugin 实例
	// 这里使用 BasePlugin 作为示例（应该使用 alipay.NewBasePlugin）
	// 注意：实际实现应该在 internal/plugin/alipay 包中
	return NewBasePlugin(pluginID)
}

// NewWechatWapPlugin 创建微信手机网站支付插件（示例）
// 微信插件应该继承 BasePlugin，而不是 alipay.BasePlugin
// 微信插件应该有自己的基类（WechatBasePlugin），继承自 plugin.BasePlugin
func NewWechatWapPlugin(pluginID int64) Plugin {
	// 微信插件应该有自己的基类（WechatBasePlugin），继承自 BasePlugin
	// 这里使用 BasePlugin 作为示例
	return NewBasePlugin(pluginID)
}

// NewWechatAppPlugin 创建微信APP支付插件（示例）
// 微信插件应该继承 BasePlugin，而不是 alipay.BasePlugin
// 微信插件应该有自己的基类（WechatBasePlugin），继承自 plugin.BasePlugin
func NewWechatAppPlugin(pluginID int64) Plugin {
	// 微信插件应该有自己的基类（WechatBasePlugin），继承自 BasePlugin
	// 这里使用 BasePlugin 作为示例
	return NewBasePlugin(pluginID)
}
