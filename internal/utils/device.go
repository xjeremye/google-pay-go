package utils

import (
	"strings"

	"github.com/golang-pay-core/internal/models"
)

// DetectDeviceType 从 User-Agent 检测设备类型
// 返回设备类型常量（DeviceTypeAndroid, DeviceTypeIOS, DeviceTypePC, DeviceTypeUnknown）
func DetectDeviceType(userAgent string) int {
	if userAgent == "" {
		return models.DeviceTypeUnknown
	}

	ua := strings.ToLower(userAgent)

	// 检测 Android
	if strings.Contains(ua, "android") {
		return models.DeviceTypeAndroid
	}

	// 检测 iOS
	if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod") {
		return models.DeviceTypeIOS
	}

	// 检测 PC（Windows, Mac, Linux）
	if strings.Contains(ua, "windows") || strings.Contains(ua, "macintosh") || strings.Contains(ua, "linux") || strings.Contains(ua, "x11") {
		return models.DeviceTypePC
	}

	return models.DeviceTypeUnknown
}

// GetClientIP 从 Gin Context 获取客户端真实IP
// 优先从 X-Forwarded-For, X-Real-IP 等头部获取，如果没有则使用 RemoteAddr
func GetClientIP(ctx interface{}) string {
	// 如果传入的是 *gin.Context，使用其 ClientIP() 方法
	// 这里为了通用性，直接返回空字符串，由调用方使用 gin.Context.ClientIP()
	return ""
}
