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

// PluginService 插件服务，用于获取插件相关信息
type PluginService struct {
	redis *redis.Client
}

// NewPluginService 创建插件服务
func NewPluginService() *PluginService {
	return &PluginService{
		redis: database.RDB,
	}
}

// GetPlugin 获取插件信息（带缓存）
func (s *PluginService) GetPlugin(ctx context.Context, pluginID int64) (*models.PayPlugin, error) {
	cacheKey := fmt.Sprintf("plugin:%d", pluginID)

	// 尝试从缓存获取
	if val, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var plugin models.PayPlugin
		if err := json.Unmarshal([]byte(val), &plugin); err == nil {
			return &plugin, nil
		}
	}

	// 从数据库获取
	var plugin models.PayPlugin
	if err := database.DB.First(&plugin, pluginID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("插件不存在")
		}
		return nil, err
	}

	// 缓存插件信息
	if data, err := json.Marshal(plugin); err == nil {
		s.redis.Set(ctx, cacheKey, data, 1*time.Hour)
	}

	return &plugin, nil
}

// GetPluginConfigs 获取插件配置列表（带缓存）
func (s *PluginService) GetPluginConfigs(ctx context.Context, pluginID int64) ([]models.PayPluginConfig, error) {
	cacheKey := fmt.Sprintf("plugin_config:%d", pluginID)

	// 尝试从缓存获取
	if val, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var configs []models.PayPluginConfig
		if err := json.Unmarshal([]byte(val), &configs); err == nil {
			return configs, nil
		}
	}

	// 从数据库获取
	var configs []models.PayPluginConfig
	if err := database.DB.Where("parent_id = ? AND status = ?", pluginID, true).
		Order("sort ASC").
		Find(&configs).Error; err != nil {
		return nil, err
	}

	// 缓存配置列表
	if data, err := json.Marshal(configs); err == nil {
		s.redis.Set(ctx, cacheKey, data, 1*time.Hour)
	}

	return configs, nil
}

// GetPluginConfigByKey 根据key获取插件配置
func (s *PluginService) GetPluginConfigByKey(ctx context.Context, pluginID int64, key string) (*models.PayPluginConfig, error) {
	configs, err := s.GetPluginConfigs(ctx, pluginID)
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.Key == key {
			return &config, nil
		}
	}

	return nil, fmt.Errorf("配置不存在: %s", key)
}

// GetPluginPayTypes 获取插件支持的支付类型（带缓存）
func (s *PluginService) GetPluginPayTypes(ctx context.Context, pluginID int64) ([]models.PayType, error) {
	cacheKey := fmt.Sprintf("plugin_pay_types:%d", pluginID)

	// 尝试从缓存获取
	if val, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var payTypes []models.PayType
		if err := json.Unmarshal([]byte(val), &payTypes); err == nil {
			return payTypes, nil
		}
	}

	// 从数据库获取（通过关联表）
	var payTypes []models.PayType
	if err := database.DB.Table("dvadmin_pay_type").
		Joins("INNER JOIN dvadmin_pay_plugin_pay_types ON dvadmin_pay_type.id = dvadmin_pay_plugin_pay_types.paytype_id").
		Where("dvadmin_pay_plugin_pay_types.payplugin_id = ? AND dvadmin_pay_type.status = ?", pluginID, true).
		Find(&payTypes).Error; err != nil {
		return nil, err
	}

	// 缓存支付类型列表
	if data, err := json.Marshal(payTypes); err == nil {
		s.redis.Set(ctx, cacheKey, data, 1*time.Hour)
	}

	return payTypes, nil
}

// GetPluginUpstream 获取插件上游类型
func (s *PluginService) GetPluginUpstream(ctx context.Context, pluginID int64) (int, error) {
	config, err := s.GetPluginConfigByKey(ctx, pluginID, "type")
	if err != nil {
		return 0, err
	}

	// 解析 value 字段（可能是 JSON 字符串或直接的值）
	var upstream int
	if err := json.Unmarshal([]byte(config.Value), &upstream); err != nil {
		// 如果不是 JSON，尝试直接解析为整数
		if _, err := fmt.Sscanf(config.Value, "%d", &upstream); err != nil {
			return 0, fmt.Errorf("无法解析插件类型: %v", err)
		}
	}

	return upstream, nil
}
