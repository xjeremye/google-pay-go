package utils

import (
	"fmt"
	"time"

	"github.com/golang-pay-core/config"
)

// GenerateID 生成唯一ID（基于时间戳和随机数）
func GenerateID() string {
	timestamp := time.Now().UnixNano() / 1e6 // 毫秒时间戳
	return fmt.Sprintf("%d", timestamp)
}

// GenerateOrderNo 生成订单号
func GenerateOrderNo() string {
	prefix := config.Cfg.Payment.OrderPrefix
	if prefix == "" {
		prefix = "PAY"
	}
	timestamp := time.Now().Format("20060102150405")
	random := time.Now().UnixNano() % 10000
	return fmt.Sprintf("%s%s%04d", prefix, timestamp, random)
}

