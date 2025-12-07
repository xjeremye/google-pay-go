package plugin

import (
	"context"
)

// RegisterAlipayPhonePlugin 注册支付宝手机网站支付插件
func RegisterAlipayPhonePlugin() {
	registry := GetRegistry()

	// 注册 alipay_phone 插件
	registry.Register("alipay_phone", func(ctx context.Context, pluginID int64, pluginType string) (Plugin, error) {
		return NewAlipayPhonePlugin(pluginID), nil
	})
}
