package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/golang-pay-core/internal/logger"
	"go.uber.org/zap"
)

var (
	// globalConfigProvider 全局插件配置提供者（避免循环依赖）
	globalConfigProvider PluginConfigProvider
	configProviderOnce   sync.Once
	configProviderMu     sync.RWMutex
)

// SetConfigProvider 设置全局插件配置提供者（应在应用启动时调用）
func SetConfigProvider(provider PluginConfigProvider) {
	configProviderMu.Lock()
	defer configProviderMu.Unlock()
	globalConfigProvider = provider
}

// getConfigProvider 获取插件配置提供者
func getConfigProvider() PluginConfigProvider {
	configProviderMu.RLock()
	defer configProviderMu.RUnlock()
	return globalConfigProvider
}

// BasePlugin 基础插件实现（所有插件的基类）
// 提供通用的插件功能，不包含任何第三方支付平台特定的逻辑
// 第三方支付平台（支付宝、微信、京东等）的插件应该继承或嵌入此基类
type BasePlugin struct {
	pluginID int64
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(pluginID int64) *BasePlugin {
	return &BasePlugin{
		pluginID: pluginID,
	}
}

// GetPluginID 获取插件ID
func (p *BasePlugin) GetPluginID() int64 {
	return p.pluginID
}

// CreateOrder 创建订单（基础实现）
// 子类应该覆盖此方法以实现具体的支付逻辑
func (p *BasePlugin) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// 基础实现：生成一个占位符支付URL
	// 实际实现应该根据插件类型调用对应的支付接口
	payURL := fmt.Sprintf("https://pay.example.com/pay?order_no=%s&plugin_id=%d", req.OrderNo, req.PluginID)
	return NewSuccessResponse(payURL), nil
}

// WaitProduct 等待产品（基础实现）
// 子类应该覆盖此方法以实现具体的产品选择逻辑
func (p *BasePlugin) WaitProduct(ctx context.Context, req *WaitProductRequest) (*WaitProductResponse, error) {
	// 基础实现：返回错误，提示子类需要实现
	return NewWaitProductErrorResponse(7318, "WaitProduct 方法需要由子类实现"), nil
}

// CallbackSubmit 下单回调（基础实现）
// 子类应该覆盖此方法以实现具体的回调逻辑
func (p *BasePlugin) CallbackSubmit(ctx context.Context, req *CallbackSubmitRequest) error {
	// 基础实现：什么都不做
	// 子类可以覆盖此方法以实现统计更新等逻辑
	return nil
}

// 实现 PluginCapabilities 接口（可选）
var _ PluginCapabilities = (*BasePlugin)(nil)

// CanHandleExtra 是否可以处理额外参数
func (p *BasePlugin) CanHandleExtra() bool {
	return false
}

// AutoExtra 是否自动处理额外参数
func (p *BasePlugin) AutoExtra() bool {
	return false
}

// ExtraNeedProduct 额外参数是否需要产品
func (p *BasePlugin) ExtraNeedProduct() bool {
	return false
}

// ExtraNeedCookie 额外参数是否需要Cookie
func (p *BasePlugin) ExtraNeedCookie() bool {
	return false
}

// GetTimeout 获取订单超时时间（秒）
// 参考 Python: get_plugin_out_time(plugin_id)
// 从插件配置中获取 out_time，如果没有配置则返回默认值 300 秒（5分钟）
// 使用 CacheService 获取配置（带缓存）
func (p *BasePlugin) GetTimeout(ctx context.Context, pluginID int64) int {
	// 获取配置提供者（通过 CacheService）
	provider := getConfigProvider()
	if provider == nil {
		logger.Logger.Warn("插件配置提供者未设置，使用默认超时时间",
			zap.Int64("plugin_id", pluginID),
			zap.Int("default_timeout", 300))
		return 300
	}

	// 从缓存服务获取配置
	config, err := provider.GetPluginConfigByKey(ctx, pluginID, "out_time")
	if err != nil {
		logger.Logger.Debug("未找到插件超时配置，使用默认值",
			zap.Int64("plugin_id", pluginID),
			zap.Error(err),
			zap.Int("default_timeout", 300))
		return 300
	}

	return p.parseTimeoutValue(ctx, pluginID, config.Value)
}

// parseTimeoutValue 解析超时配置值
func (p *BasePlugin) parseTimeoutValue(ctx context.Context, pluginID int64, value string) int {
	logger.Logger.Debug("解析插件超时配置值",
		zap.Int64("plugin_id", pluginID),
		zap.String("config_value", value),
		zap.Int("value_length", len(value)))

	var timeout int

	// 先尝试直接解析为整数（如果值是 "30" 这样的字符串）
	if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
		logger.Logger.Info("从插件配置获取超时时间（直接解析）",
			zap.Int64("plugin_id", pluginID),
			zap.String("config_value", value),
			zap.Int("timeout_seconds", parsed))
		return parsed
	}

	// 尝试解析为 JSON 数字（如果值是 JSON 格式的数字，如 30）
	if err := json.Unmarshal([]byte(value), &timeout); err == nil {
		if timeout > 0 {
			logger.Logger.Info("从插件配置获取超时时间（JSON 数字）",
				zap.Int64("plugin_id", pluginID),
				zap.String("config_value", value),
				zap.Int("timeout_seconds", timeout))
			return timeout
		}
	}

	// 尝试解析为 JSON 对象（如果值是 JSON 对象，如 {"value": 30}）
	var valueMap map[string]interface{}
	if err := json.Unmarshal([]byte(value), &valueMap); err == nil {
		if val, ok := valueMap["value"]; ok {
			logger.Logger.Debug("找到 JSON 对象中的 value 字段",
				zap.Int64("plugin_id", pluginID),
				zap.String("value_type", fmt.Sprintf("%T", val)),
				zap.Any("value", val))

			// 尝试多种类型转换
			switch v := val.(type) {
			case float64:
				timeout = int(v)
			case int:
				timeout = v
			case int64:
				timeout = int(v)
			case string:
				if parsed, err := strconv.Atoi(v); err == nil {
					timeout = parsed
				}
			}

			if timeout > 0 {
				logger.Logger.Info("从插件配置获取超时时间（JSON 对象）",
					zap.Int64("plugin_id", pluginID),
					zap.String("config_value", value),
					zap.Int("timeout_seconds", timeout))
				return timeout
			}
		}
	}

	logger.Logger.Warn("插件配置值解析失败，使用默认值",
		zap.Int64("plugin_id", pluginID),
		zap.String("config_value", value),
		zap.Int("value_length", len(value)),
		zap.Int("default_timeout", 300))

	// 如果都没有配置，返回默认值 300 秒（5分钟）
	// 参考 Python: return 300
	return 300
}
