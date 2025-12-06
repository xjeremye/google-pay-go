# 插件系统使用指南

## 概述

插件系统使用支付方式的 **key**（如 `alipay_wap`、`wechat_app`）来创建插件实例，而不是依赖数据库的 ID。这样设计的好处是：

1. **稳定性**：ID 可能会变动，但 key 是稳定的业务标识
2. **可读性**：key 更直观，便于理解和调试
3. **灵活性**：可以轻松添加新的插件类型

## 插件注册

### 1. 在应用启动时注册插件

在 `main.go` 或初始化函数中注册插件：

```go
package main

import (
    "github.com/golang-pay-core/internal/plugin"
)

func init() {
    registry := plugin.GetRegistry()
    
    // 注册支付宝插件
    registry.Register("alipay_wap", func(ctx context.Context, pluginID int64, pluginType string) (plugin.Plugin, error) {
        return NewAlipayWapPlugin(pluginID), nil
    })
    
    registry.Register("alipay_app", func(ctx context.Context, pluginID int64, pluginType string) (plugin.Plugin, error) {
        return NewAlipayAppPlugin(pluginID), nil
    })
    
    // 注册微信支付插件
    registry.Register("wechat_wap", func(ctx context.Context, pluginID int64, pluginType string) (plugin.Plugin, error) {
        return NewWechatWapPlugin(pluginID), nil
    })
    
    registry.Register("wechat_app", func(ctx context.Context, pluginID int64, pluginType string) (plugin.Plugin, error) {
        return NewWechatAppPlugin(pluginID), nil
    })
}
```

### 2. 实现插件接口

每个插件需要实现 `Plugin` 接口：

```go
package plugin

import (
    "context"
    "github.com/golang-pay-core/internal/plugin"
)

type AlipayWapPlugin struct {
    pluginID int64
    // 其他字段，如配置、客户端等
}

func NewAlipayWapPlugin(pluginID int64) *AlipayWapPlugin {
    return &AlipayWapPlugin{
        pluginID: pluginID,
    }
}

// 实现 Plugin 接口
func (p *AlipayWapPlugin) CreateOrder(ctx context.Context, req *plugin.CreateOrderRequest) (*plugin.CreateOrderResponse, error) {
    // 实现创建订单逻辑
    // 1. 从数据库或缓存获取插件配置
    // 2. 构建支付请求
    // 3. 调用支付宝 API
    // 4. 返回支付 URL
    
    payURL := "https://openapi.alipay.com/gateway.do?..."
    return plugin.NewSuccessResponse(payURL), nil
}

// 可选：实现 PluginCapabilities 接口
func (p *AlipayWapPlugin) CanHandleExtra() bool {
    return false
}

func (p *AlipayWapPlugin) AutoExtra() bool {
    return false
}

func (p *AlipayWapPlugin) ExtraNeedProduct() bool {
    return false
}

func (p *AlipayWapPlugin) ExtraNeedCookie() bool {
    return false
}

func (p *AlipayWapPlugin) GetTimeout(ctx context.Context, pluginID int64) int {
    return 300 // 5分钟
}
```

## 插件类型映射

插件类型（key）与数据库中的 `dvadmin_pay_type` 表的 `key` 字段对应：

| 插件类型 (key) | 说明 | 示例 |
|--------------|------|------|
| `alipay_wap` | 支付宝手机网站支付 | 支付宝 H5 支付 |
| `alipay_app` | 支付宝 APP 支付 | 支付宝 APP 支付 |
| `wechat_wap` | 微信 H5 支付 | 微信 H5 支付 |
| `wechat_app` | 微信 APP 支付 | 微信 APP 支付 |

## 工作流程

1. **订单创建时**：
   - 系统从渠道（PayChannel）获取 `plugin_id`
   - 根据 `plugin_id` 查询插件信息
   - 获取插件关联的支付类型（PayType）
   - 使用 PayType 的 `key` 作为 `PluginType`（如 `alipay_wap`）

2. **插件实例化**：
   - 插件管理器根据 `PluginType` 从注册表查找工厂函数
   - 使用工厂函数创建插件实例
   - 插件实例会被缓存（使用 `PluginType` 作为 key）

3. **创建支付订单**：
   - 调用插件的 `CreateOrder` 方法
   - 插件从 `OrderContext` 中获取所需信息
   - 返回支付 URL

## 注意事项

1. **插件注册时机**：必须在应用启动时、订单服务初始化之前注册插件
2. **插件类型唯一性**：每个插件类型（key）只能注册一次
3. **默认插件**：如果未找到对应的插件类型，会使用默认插件（`default`）
4. **插件缓存**：插件实例会被缓存，避免重复创建

## 扩展插件

要添加新的支付方式插件：

1. 在数据库中创建对应的 `PayType` 记录，设置唯一的 `key`
2. 实现插件接口（`Plugin` 和可选的 `PluginCapabilities`）
3. 在应用启动时注册插件工厂函数
4. 确保插件类型与数据库中的 `key` 一致

