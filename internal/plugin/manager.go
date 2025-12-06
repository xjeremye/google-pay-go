package plugin

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/golang-pay-core/internal/models"
)

// PluginInfoProvider 插件信息提供者接口（避免循环依赖）
type PluginInfoProvider interface {
	GetPlugin(ctx context.Context, pluginID int64) (interface{}, error)
	GetPluginUpstream(ctx context.Context, pluginID int64) (int, error)
	GetPluginPayTypes(ctx context.Context, pluginID int64) ([]interface{}, error)
}

// Manager 插件管理器
type Manager struct {
	redis        *redis.Client
	infoProvider PluginInfoProvider
	plugins      map[string]Plugin // 插件实例缓存，key 为 pluginType（如 alipay_wap）
	mu           sync.RWMutex
	registry     *Registry
}

// NewManager 创建插件管理器
func NewManager(redis *redis.Client) *Manager {
	return &Manager{
		redis:    redis,
		plugins:  make(map[string]Plugin),
		registry: GetRegistry(),
	}
}

// SetInfoProvider 设置插件信息提供者
func (m *Manager) SetInfoProvider(provider PluginInfoProvider) {
	m.infoProvider = provider
}

// GetPluginByCtx 根据上下文获取插件实例
// 使用 PluginType（支付方式的 key，如 alipay_wap）来创建插件，而不是 PluginID
func (m *Manager) GetPluginByCtx(ctx context.Context, orderCtx OrderContext) (Plugin, error) {
	pluginType := orderCtx.GetPluginType()
	if pluginType == "" {
		return nil, fmt.Errorf("插件类型不能为空")
	}

	pluginID := orderCtx.GetPluginID()
	if pluginID == 0 {
		return nil, fmt.Errorf("插件ID不能为空")
	}

	// 先从缓存获取（使用 pluginType 作为 key）
	m.mu.RLock()
	if plugin, ok := m.plugins[pluginType]; ok {
		m.mu.RUnlock()
		return plugin, nil
	}
	m.mu.RUnlock()

	// 验证插件状态（如果需要）
	if m.infoProvider != nil {
		pluginInfo, err := m.infoProvider.GetPlugin(ctx, pluginID)
		if err != nil {
			return nil, fmt.Errorf("获取插件信息失败: %w", err)
		}

		// 检查插件状态
		if plugin, ok := pluginInfo.(*models.PayPlugin); ok {
			if !plugin.Status {
				return nil, fmt.Errorf("插件已禁用")
			}
		} else if pluginMap, ok := pluginInfo.(map[string]interface{}); ok {
			if status, ok := pluginMap["status"].(bool); ok && !status {
				return nil, fmt.Errorf("插件已禁用")
			}
		}
	}

	// 从注册表获取插件工厂
	factory, err := m.registry.GetFactory(pluginType)
	if err != nil {
		return nil, fmt.Errorf("获取插件工厂失败: %w", err)
	}

	// 使用工厂创建插件实例
	plugin, err := factory(ctx, pluginID, pluginType)
	if err != nil {
		return nil, fmt.Errorf("创建插件实例失败: %w", err)
	}

	// 缓存插件实例（使用 pluginType 作为 key）
	m.mu.Lock()
	m.plugins[pluginType] = plugin
	m.mu.Unlock()

	return plugin, nil
}

// ClearCache 清除插件缓存（使用 pluginType）
func (m *Manager) ClearCache(pluginType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.plugins, pluginType)
}

// ClearAllCache 清除所有插件缓存
func (m *Manager) ClearAllCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plugins = make(map[string]Plugin)
}
