# 插件目录结构说明

## 一、目录结构

```
internal/plugin/
├── base_plugin.go          # BasePlugin（通用基类，所有插件的基类）
├── interfaces.go           # 插件接口定义
├── registry.go             # 插件注册表
├── manager.go              # 插件管理器
├── writeoff.go             # 核销相关工具（通用功能）
├── example.go              # 使用示例
└── alipay/                 # 支付宝插件目录
    ├── base_plugin.go      # BasePlugin（支付宝基类，继承 plugin.BasePlugin）
    ├── alipay_phone.go     # PhonePlugin（支付宝手机网站支付）
    ├── product_selector.go # 产品选择逻辑（支付宝特定）
    ├── statistics.go       # 日统计服务（支付宝特定）
    └── register.go         # 插件注册函数
```

## 二、设计原则

### 2.1 按第三方平台分组

- **通用功能**：放在 `plugin/` 根目录
- **平台特定功能**：放在 `plugin/{platform}/` 目录下
  - 支付宝：`plugin/alipay/`
  - 微信（未来）：`plugin/wechat/`
  - 京东（未来）：`plugin/jd/`

### 2.2 命名规范

- **通用基类**：`BasePlugin`（在 `plugin/` 目录）
- **平台基类**：`BasePlugin`（在 `plugin/{platform}/` 目录，如 `alipay.BasePlugin`）
- **具体插件文件**：文件名与 pay type 保持一致（如 `alipay_phone.go` 对应 pay type `alipay_phone`）
- **具体插件类型**：使用简洁名称（如 `PhonePlugin` 而不是 `AlipayPhonePlugin`）

### 2.3 包导入

```go
// 导入通用插件包
import "github.com/golang-pay-core/internal/plugin"

// 导入支付宝插件包
import alipayplugin "github.com/golang-pay-core/internal/plugin/alipay"
```

## 三、文件说明

### 3.1 通用文件（plugin/ 目录）

- **base_plugin.go**: 所有插件的通用基类，不包含任何第三方平台特定逻辑
- **interfaces.go**: 插件接口定义（Plugin、PluginCapabilities 等）
- **registry.go**: 插件注册表，管理所有插件的工厂函数
- **manager.go**: 插件管理器，负责创建和管理插件实例
- **writeoff.go**: 核销相关工具函数（通用功能，所有平台都可以使用）

### 3.2 支付宝文件（plugin/alipay/ 目录）

- **base_plugin.go**: 支付宝基础插件，继承 `plugin.BasePlugin`
  - 实现支付宝产品选择逻辑
  - 实现支付宝日统计更新逻辑
  
- **alipay_phone.go**: 支付宝手机网站支付插件
  - 实现具体的支付URL生成逻辑
  - 处理支付宝手机网站支付的回调
  - 文件名与 pay type (`alipay_phone`) 保持一致，方便查找

- **product_selector.go**: 支付宝产品选择逻辑
  - `getAlipayProduct()`: 获取支付宝产品
  - 包含产品筛选、限额检查、日笔数限制等逻辑

- **statistics.go**: 支付宝日统计服务
  - `DayStatisticsService`: 日统计服务
  - 支持普通模式、公池模式、神码模式的统计

- **register.go**: 插件注册函数
  - `RegisterPhonePlugin()`: 注册支付宝手机网站支付插件

## 四、使用示例

### 4.1 注册支付宝插件

```go
// main.go
import (
    _ "github.com/golang-pay-core/internal/plugin/alipay" // 导入以触发自动注册
)

func main() {
    // 插件已通过 init() 函数自动注册，无需手动调用
    // 所有插件的注册逻辑都在各自的包中管理，保持 main.go 的简洁性
}
```

**自动注册机制**：
- 每个平台的 `register.go` 文件中的 `init()` 函数会在包导入时自动执行
- 无需在 `main.go` 中手动调用注册函数
- 保持入口文件的简洁性

### 4.2 创建新的支付宝插件

```go
// internal/plugin/alipay/alipay_wap.go (文件名与 pay type 保持一致)
package alipay

import "github.com/golang-pay-core/internal/plugin"

type WapPlugin struct {
    *BasePlugin  // 嵌入支付宝基类
}

func NewWapPlugin(pluginID int64) *WapPlugin {
    return &WapPlugin{
        BasePlugin: NewBasePlugin(pluginID),
    }
}

// 实现 CreateOrder 方法
func (p *WapPlugin) CreateOrder(ctx context.Context, req *plugin.CreateOrderRequest) (*plugin.CreateOrderResponse, error) {
    // 实现具体的支付逻辑
}
```

### 4.3 创建新的第三方平台插件（微信示例）

```go
// internal/plugin/wechat/base_plugin.go
package wechat

import "github.com/golang-pay-core/internal/plugin"

type BasePlugin struct {
    *plugin.BasePlugin  // 嵌入通用基类
}

func NewBasePlugin(pluginID int64) *BasePlugin {
    return &BasePlugin{
        BasePlugin: plugin.NewBasePlugin(pluginID),
    }
}

// 实现微信特定的产品选择逻辑
func (p *BasePlugin) WaitProduct(...) { ... }
```

## 五、优势

### 5.1 清晰的目录结构

- 按平台分组，易于查找和维护
- 通用功能与平台特定功能分离

### 5.2 易于扩展

- 添加新平台只需创建新目录
- 不影响现有代码

### 5.3 避免命名冲突

- 使用包名区分不同平台的基类
- `plugin.BasePlugin` vs `alipay.BasePlugin`

### 5.4 代码组织

- 相关功能集中在一个目录
- 减少文件查找时间

## 六、迁移说明

### 6.1 已移动的文件

- `alipay_base_plugin.go` → `alipay/base_plugin.go`
- `alipay_phone.go` → `alipay/alipay_phone.go`
- `register_alipay_phone.go` → `alipay/register.go`
- `product_selector.go` → `alipay/product_selector.go`
- `statistics.go` → `alipay/statistics.go`

### 6.2 已更新的引用

- `main.go`: 更新了插件注册调用
- `example.go`: 更新了示例代码注释
- 所有导入路径已更新

### 6.3 命名变更

- `AlipayBasePlugin` → `alipay.BasePlugin`
- `AlipayPhonePlugin` → `alipay.PhonePlugin`
- `RegisterAlipayPhonePlugin()` → `alipay.RegisterPhonePlugin()`

## 七、未来扩展

### 7.1 添加微信支付

```
internal/plugin/
└── wechat/
    ├── base_plugin.go
    ├── wechat_wap.go      # 文件名与 pay type 保持一致
    ├── wechat_app.go      # 文件名与 pay type 保持一致
    └── register.go
```

### 7.2 添加京东支付

```
internal/plugin/
└── jd/
    ├── base_plugin.go
    ├── jd_wap.go          # 文件名与 pay type 保持一致
    └── register.go
```

每个平台都有自己独立的目录，互不干扰，便于维护和扩展。
