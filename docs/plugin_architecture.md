# 插件架构设计文档

## 一、架构概述

插件系统采用**分层继承架构**，支持多第三方支付平台的扩展。

```
BasePlugin (基础插件)
├── AlipayBasePlugin (支付宝基础插件)
│   ├── AlipayPhonePlugin (支付宝手机网站支付)
│   ├── AlipayWapPlugin (支付宝手机网站支付)
│   ├── AlipayAppPlugin (支付宝APP支付)
│   └── AlipayQrPlugin (支付宝扫码支付)
├── WechatBasePlugin (微信基础插件) [待实现]
│   ├── WechatWapPlugin (微信手机网站支付)
│   └── WechatAppPlugin (微信APP支付)
└── JdBasePlugin (京东基础插件) [待实现]
    └── JdWapPlugin (京东手机网站支付)
```

## 二、核心组件

### 2.1 BasePlugin（基础插件）

**位置**: `internal/plugin/base_plugin.go`

**职责**:
- 提供所有插件的通用功能
- 定义插件接口的默认实现
- **不包含任何第三方支付平台特定的逻辑**

**特点**:
- 通用性：适用于所有第三方支付平台
- 可扩展性：子类可以覆盖任何方法
- 最小化：只包含必要的通用功能

**方法**:
- `CreateOrder()`: 创建订单（默认返回占位符URL）
- `WaitProduct()`: 等待产品（默认返回错误，提示子类实现）
- `CallbackSubmit()`: 下单回调（默认什么都不做）
- `GetTimeout()`: 获取订单超时时间（默认5分钟）

### 2.2 AlipayBasePlugin（支付宝基础插件）

**位置**: `internal/plugin/alipay/base_plugin.go`

**职责**:
- 提供支付宝插件通用的功能
- 实现支付宝特定的产品选择和统计逻辑
- 作为所有支付宝插件的基类

**特点**:
- 继承自 `plugin.BasePlugin`
- 包含支付宝特定的业务逻辑：
  - 支付宝产品选择（`WaitProduct`）
  - 支付宝日统计更新（`CallbackSubmit`）
  - 订单备注更新

**方法**:
- `WaitProduct()`: 支付宝产品选择逻辑
- `CallbackSubmit()`: 支付宝下单回调（更新备注和统计）

### 2.3 具体插件实现

**示例**: `PhonePlugin`（支付宝手机网站支付插件）

**位置**: `internal/plugin/alipay/alipay_phone.go`

**职责**:
- 实现支付宝手机网站支付的具体逻辑
- 生成支付URL
- 处理支付回调

**特点**:
- 嵌入 `AlipayBasePlugin`
- 继承支付宝通用功能
- 实现具体的支付方式逻辑

## 三、架构优势

### 3.1 清晰的职责分离

- **BasePlugin**: 通用功能，不依赖任何第三方平台
- **AlipayBasePlugin**: 支付宝特定功能
- **具体插件**: 具体支付方式的实现

### 3.2 易于扩展

添加新的第三方支付平台只需：

1. 创建新的基础插件（如 `WechatBasePlugin`）
2. 继承 `BasePlugin`
3. 实现该平台特定的逻辑
4. 创建具体的支付方式插件

### 3.3 代码复用

- 支付宝的所有插件共享 `AlipayBasePlugin` 的功能
- 所有插件共享 `BasePlugin` 的通用功能
- 减少代码重复，提高可维护性

### 3.4 类型安全

- 通过嵌入和继承，确保类型安全
- 编译时检查，避免运行时错误

## 四、使用示例

### 4.1 创建支付宝插件

```go
// PhonePlugin 嵌入 BasePlugin（支付宝基类）
// 位置: internal/plugin/alipay/alipay_phone.go
package alipay

import "github.com/golang-pay-core/internal/plugin"

type PhonePlugin struct {
    *BasePlugin  // 嵌入 BasePlugin（支付宝基类）
}

func NewPhonePlugin(pluginID int64) *PhonePlugin {
    return &PhonePlugin{
        BasePlugin: NewBasePlugin(pluginID),
    }
}
```

### 4.2 创建微信插件（未来扩展）

```go
// 位置: internal/plugin/wechat/base_plugin.go
package wechat

import "github.com/golang-pay-core/internal/plugin"

// BasePlugin 微信基础插件
type BasePlugin struct {
    *plugin.BasePlugin
}

// 位置: internal/plugin/wechat/wap.go
package wechat

// WapPlugin 微信手机网站支付插件
type WapPlugin struct {
    *BasePlugin
}
```

## 五、文件结构

```
internal/plugin/
├── base_plugin.go          # BasePlugin（通用基类）
├── interfaces.go           # 插件接口定义
├── registry.go             # 插件注册表
├── manager.go              # 插件管理器
├── writeoff.go             # 核销相关工具（通用）
├── example.go              # 使用示例
└── alipay/                 # 支付宝插件目录
    ├── base_plugin.go      # BasePlugin（支付宝基类）
    ├── alipay_phone.go     # PhonePlugin（支付宝手机网站支付）
    ├── product_selector.go # 产品选择逻辑（支付宝特定）
    ├── statistics.go       # 日统计服务（支付宝特定）
    └── register.go         # 插件注册
```

## 六、设计原则

### 6.1 单一职责原则

- `BasePlugin`: 只负责通用功能
- `AlipayBasePlugin`: 只负责支付宝通用功能
- 具体插件: 只负责具体支付方式的实现

### 6.2 开闭原则

- 对扩展开放：可以轻松添加新的第三方平台
- 对修改封闭：不需要修改现有代码

### 6.3 依赖倒置原则

- 高层模块（具体插件）依赖抽象（BasePlugin）
- 不依赖具体实现

## 七、未来扩展

### 7.1 微信支付插件

```go
// 位置: internal/plugin/wechat/base_plugin.go
package wechat

import "github.com/golang-pay-core/internal/plugin"

// BasePlugin 微信基础插件
type BasePlugin struct {
    *plugin.BasePlugin
}

// 实现微信特定的产品选择和统计逻辑
func (p *BasePlugin) WaitProduct(...) { ... }
func (p *BasePlugin) CallbackSubmit(...) { ... }
```

### 7.2 京东支付插件

```go
// 位置: internal/plugin/jd/base_plugin.go
package jd

import "github.com/golang-pay-core/internal/plugin"

// BasePlugin 京东基础插件
type BasePlugin struct {
    *plugin.BasePlugin
}

// 实现京东特定的产品选择和统计逻辑
func (p *BasePlugin) WaitProduct(...) { ... }
func (p *BasePlugin) CallbackSubmit(...) { ... }
```

## 八、总结

通过分层继承架构，我们实现了：

1. ✅ **清晰的职责分离**：通用功能与平台特定功能分离
2. ✅ **易于扩展**：添加新平台只需创建新的基类
3. ✅ **代码复用**：共享通用功能，减少重复代码
4. ✅ **类型安全**：编译时检查，避免运行时错误
5. ✅ **可维护性**：结构清晰，易于理解和维护
