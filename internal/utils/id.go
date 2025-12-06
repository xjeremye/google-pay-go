package utils

import (
	"fmt"
	"time"
)

// GenerateID 生成唯一ID（基于时间戳和随机数）
func GenerateID() string {
	timestamp := time.Now().UnixNano() / 1e6 // 毫秒时间戳
	return fmt.Sprintf("%d", timestamp)
}

// GenerateOrderNo 生成订单号
// 订单号前缀应从数据库系统配置表（dvadmin_system_config）读取
// 当前使用默认值 "PAY"，后续可从数据库获取
func GenerateOrderNo() string {
	// TODO: 从数据库 dvadmin_system_config 表读取订单号前缀配置
	// 当前使用默认值
	prefix := "PAY"
	timestamp := time.Now().Format("20060102150405")
	random := time.Now().UnixNano() % 10000
	return fmt.Sprintf("%s%s%04d", prefix, timestamp, random)
}

