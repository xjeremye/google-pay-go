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

// RegisterWapPlugin 注册支付宝手机网站支付插件
func RegisterWapPlugin() {
	registry := plugin.GetRegistry()
	registry.Register("alipay_wap", func(ctx context.Context, pluginID int64, pluginType string) (plugin.Plugin, error) {
		return NewWapPlugin(pluginID), nil
	})
}

// RegisterDdmPlugin 注册支付宝扫码支付插件
func RegisterDdmPlugin() {
	registry := plugin.GetRegistry()
	registry.Register("alipay_ddm", func(ctx context.Context, pluginID int64, pluginType string) (plugin.Plugin, error) {
		return NewDdmPlugin(pluginID), nil
	})
}

// RegisterPcPlugin 注册支付宝电脑网站支付插件
func RegisterPcPlugin() {
	registry := plugin.GetRegistry()
	registry.Register("alipay_pc", func(ctx context.Context, pluginID int64, pluginType string) (plugin.Plugin, error) {
		return NewPcPlugin(pluginID), nil
	})
}

// RegisterAppPlugin 注册支付宝APP支付插件
func RegisterAppPlugin() {
	registry := plugin.GetRegistry()
	registry.Register("alipay_app", func(ctx context.Context, pluginID int64, pluginType string) (plugin.Plugin, error) {
		return NewAppPlugin(pluginID), nil
	})
}

// RegisterFaceToPlugin 注册支付宝当面付插件
func RegisterFaceToPlugin() {
	registry := plugin.GetRegistry()
	registry.Register("alipay_face_to", func(ctx context.Context, pluginID int64, pluginType string) (plugin.Plugin, error) {
		return NewFaceToPlugin(pluginID), nil
	})
}

// init 自动注册所有支付宝插件
// 在包导入时自动执行，无需在 main.go 中手动调用
// 使用 init() 函数可以保持 main.go 的简洁性
// 注意：init() 函数在 logger 初始化之前执行，因此不在此处记录日志
func init() {
	RegisterPhonePlugin()  // alipay_phone
	RegisterWapPlugin()    // alipay_wap
	RegisterDdmPlugin()    // alipay_ddm
	RegisterPcPlugin()     // alipay_pc
	RegisterAppPlugin()    // alipay_app
	RegisterFaceToPlugin() // alipay_face_to
}
