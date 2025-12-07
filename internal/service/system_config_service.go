package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"gorm.io/gorm"
)

// SystemConfigService 系统配置服务
type SystemConfigService struct {
	redis *redis.Client
}

// NewSystemConfigService 创建系统配置服务
func NewSystemConfigService() *SystemConfigService {
	return &SystemConfigService{
		redis: database.RDB,
	}
}

// GetSystemConfig 获取系统配置值
// key 格式：如 "alipay.inline_notify_domain"
// parentID 为 nil 时表示顶级配置
func (s *SystemConfigService) GetSystemConfig(ctx context.Context, key string, parentID *int64) (string, error) {
	cacheKey := fmt.Sprintf("system_config:%s", key)
	if parentID != nil {
		cacheKey = fmt.Sprintf("system_config:%s:%d", key, *parentID)
	}

	// 尝试从缓存获取
	if val, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		return val, nil
	}

	// 从数据库获取
	var config models.SystemConfig
	query := database.DB.Where("key = ? AND status = ?", key, true)
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	if err := query.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil // 配置不存在，返回空字符串
		}
		return "", err
	}

	// 解析 JSON 值
	var valueMap map[string]interface{}
	if err := json.Unmarshal([]byte(config.Value), &valueMap); err != nil {
		// 如果解析失败，尝试直接返回原始值
		return config.Value, nil
	}

	// 尝试获取 value 字段
	if value, ok := valueMap["value"].(string); ok {
		// 缓存配置值（1小时）
		s.redis.Set(ctx, cacheKey, value, 1*time.Hour)
		return value, nil
	}

	// 如果 value 不是字符串，返回整个 JSON
	valueBytes, _ := json.Marshal(valueMap)
	valueStr := string(valueBytes)
	s.redis.Set(ctx, cacheKey, valueStr, 1*time.Hour)
	return valueStr, nil
}

// GetSystemConfigByPath 通过路径获取系统配置
// path 格式：如 "alipay.inline_notify_domain"
// 支持嵌套路径，如 "alipay.notify.domain"
func (s *SystemConfigService) GetSystemConfigByPath(ctx context.Context, path string) (string, error) {
	// 先尝试直接获取
	value, err := s.GetSystemConfig(ctx, path, nil)
	if err != nil {
		return "", err
	}
	if value != "" {
		return value, nil
	}

	// 如果直接获取失败，尝试按点分割路径
	// 例如 "alipay.inline_notify_domain" -> 先找 "alipay"，再找 "inline_notify_domain"
	parts := splitPath(path)
	if len(parts) < 2 {
		return "", nil
	}

	// 先找父配置
	var parentConfig models.SystemConfig
	if err := database.DB.Where("key = ? AND status = ? AND parent_id IS NULL", parts[0], true).
		First(&parentConfig).Error; err != nil {
		return "", nil
	}

	// 再找子配置
	parentID := parentConfig.ID
	return s.GetSystemConfig(ctx, parts[1], &parentID)
}

// splitPath 分割配置路径
func splitPath(path string) []string {
	var parts []string
	var current string
	for _, char := range path {
		if char == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
