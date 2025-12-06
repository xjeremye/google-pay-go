package plugin

import (
	"context"
	"fmt"
	"sync"
)

// PluginFactory 插件工厂函数类型
type PluginFactory func(ctx context.Context, pluginID int64, pluginType string) (Plugin, error)

// Registry 插件注册表
type Registry struct {
	factories map[string]PluginFactory
	mu        sync.RWMutex
}

var (
	globalRegistry *Registry
	registryOnce   sync.Once
)

// GetRegistry 获取全局插件注册表
func GetRegistry() *Registry {
	registryOnce.Do(func() {
		globalRegistry = &Registry{
			factories: make(map[string]PluginFactory),
		}
		// 注册默认的基础插件
		globalRegistry.Register("default", func(ctx context.Context, pluginID int64, pluginType string) (Plugin, error) {
			return NewBasePlugin(pluginID), nil
		})
	})
	return globalRegistry
}

// Register 注册插件工厂
func (r *Registry) Register(pluginType string, factory PluginFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[pluginType] = factory
}

// GetFactory 获取插件工厂
func (r *Registry) GetFactory(pluginType string) (PluginFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	factory, ok := r.factories[pluginType]
	if !ok {
		// 如果没有找到，返回默认工厂
		factory, ok = r.factories["default"]
		if !ok {
			return nil, fmt.Errorf("插件类型 %s 未注册且无默认工厂", pluginType)
		}
	}
	return factory, nil
}

// ListRegistered 列出所有已注册的插件类型
func (r *Registry) ListRegistered() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	types := make([]string, 0, len(r.factories))
	for t := range r.factories {
		if t != "default" {
			types = append(types, t)
		}
	}
	return types
}

