# Mock 插件使用说明

## 概述

Mock 插件是专门为压测设计的模拟插件，用于模拟支付宝网关的延迟和回调，**不会调用真实的支付宝API**。

## 使用方式

Mock 插件需要通过数据库配置来使用，压测时使用**正常的订单创建接口**，通过配置的通道ID自动使用 mock 插件：

1. **创建插件**：在 `pay_plugin` 表中创建插件，`name` 设置为 "Mock插件"，`status` 为 1
2. **创建支付类型**：在 `pay_type` 表中创建支付类型，`key` 设置为 `mock`，关联到插件ID
3. **创建支付通道**：在 `pay_channel` 表中创建通道，`plugin_id` 关联到插件ID，`status` 为 1
4. **压测配置**：在 `config.env` 中配置 `CHANNEL_ID` 为 mock 通道ID
5. **运行压测**：使用正常的 `/api/v1/orders` 接口，系统会自动使用 mock 插件

## 特性

### 1. 模拟网关延迟

Mock 插件会模拟支付宝网关的响应延迟：
- **默认延迟**: 50-200ms（随机）
- **可配置**: 修改 `internal/plugin/mock/mock_plugin.go` 中的 `MinDelay` 和 `MaxDelay`

### 2. 模拟产品选择

- 返回模拟产品ID（格式：`MOCK_PRODUCT_{随机数}`）
- 返回模拟核销ID（随机生成）
- 模拟产品选择延迟（约为支付URL生成延迟的一半）

### 3. 模拟支付URL

返回格式：`mock://pay.example.com/pay?order_no=xxx&out_order_no=xxx&amount=xxx&plugin_type=mock`

### 4. 模拟回调（已启用）

Mock 插件默认启用回调模拟功能：
- `SimulateCallback`: 是否模拟回调（**默认 true**）
- `CallbackDelay`: 回调延迟时间（默认 5 秒）

**回调流程**：
1. 订单创建成功后，等待 `CallbackDelay` 秒
2. 自动发送HTTP POST请求到系统回调接口：`/api/pay/order/notify/mock/{product_id}/`
3. 模拟支付宝回调参数格式，包含订单号、交易号、交易状态等
4. 系统处理回调，更新订单状态为已支付
5. 触发商户通知（如果配置了NotifyURL）

## 配置

### 修改延迟时间

编辑 `internal/plugin/mock/mock_plugin.go`：

```go
func NewMockPlugin(pluginID int64) *MockPlugin {
	return &MockPlugin{
		BasePlugin:       plugin.NewBasePlugin(pluginID),
		MinDelay:         50,  // 修改最小延迟（毫秒）
		MaxDelay:         200, // 修改最大延迟（毫秒）
		SimulateCallback: false,
		CallbackDelay:    5,
	}
}
```

### 禁用模拟回调

如果不需要模拟支付回调，修改 `SimulateCallback` 为 `false`：

```go
SimulateCallback: false, // 禁用模拟回调
CallbackDelay:    5,     // 回调延迟时间（如果启用）
```

## 使用示例

### 压测脚本自动使用

压测脚本中的订单号会自动以 `LOADTEST_` 开头，系统会自动使用 mock 插件：

```javascript
// k6 脚本中
const outOrderNo = `LOADTEST_${timestamp}_${random}`;
```

### 手动测试

如果需要手动测试 mock 插件，需要先配置数据库：

1. **创建插件**：在 `pay_plugin` 表中创建插件，`name` 设置为 "Mock插件"，`status` 为 1
2. **创建支付类型**：在 `pay_type` 表中创建支付类型，`key` 设置为 `mock`，关联到插件ID
3. **创建支付通道**：在 `pay_channel` 表中创建通道，`plugin_id` 关联到插件ID，`status` 为 1
4. **创建订单**：使用该通道ID创建订单

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "mchId": 1,
    "channelId": <mock通道ID>,
    "mchOrderNo": "TEST_001",
    "amount": 10000,
    "notifyUrl": "http://localhost:8080/notify",
    "jumpUrl": "https://example.com/jump",
    "sign": "..."
  }'
```

**注意**：`notifyUrl` 应该指向你的服务器地址，mock插件会从该URL提取基础地址，然后发送回调到 `/api/pay/order/notify/mock/{product_id}/`

## 与真实插件的区别

| 特性 | Mock 插件 | 真实插件（如 alipay_wap） |
|------|-----------|---------------------------|
| 调用真实API | ❌ 否 | ✅ 是 |
| 模拟延迟 | ✅ 是（可配置） | ❌ 否（真实网络延迟） |
| 产生费用 | ❌ 否 | ✅ 可能 |
| 产品选择 | ✅ 模拟产品ID | ✅ 真实产品选择 |
| 支付URL | ✅ 模拟URL | ✅ 真实支付宝URL |
| 回调处理 | ✅ 可选模拟 | ✅ 真实回调 |

## 回调模拟说明

Mock 插件会自动模拟支付回调，完整流程如下：

1. **订单创建**：调用 `CreateOrder`，返回模拟支付URL
2. **等待延迟**：等待 `CallbackDelay` 秒（默认5秒）
3. **发送回调**：自动发送POST请求到系统回调接口
4. **处理回调**：系统处理回调，更新订单状态为已支付
5. **商户通知**：如果配置了 `NotifyURL`，系统会通知商户

**回调URL格式**：
```
POST {notifyUrl的基础地址}/api/pay/order/notify/mock/{product_id}/
```

**回调参数**（模拟支付宝格式）：
- `out_trade_no`: 商户订单号
- `trade_no`: 模拟支付宝交易号（格式：MOCK_{timestamp}）
- `trade_status`: 交易状态（固定为 `TRADE_SUCCESS`）
- `total_amount`: 订单金额（元）
- `sign`: 模拟签名（mock插件不需要真实签名验证）

## 注意事项

1. **仅用于压测**: Mock 插件仅用于性能测试，不应用于生产环境
2. **数据库配置**: 需要在数据库中创建对应的插件、支付类型和通道
3. **延迟配置**: 根据实际压测需求调整延迟时间，模拟真实场景
4. **回调地址**: `notifyUrl` 应该指向你的服务器，mock插件会从中提取基础地址
5. **数据清理**: 压测产生的订单数据需要定期清理
6. **回调处理**: Mock插件的回调不需要真实签名验证，系统会自动处理

## 技术实现

Mock 插件实现了 `plugin.Plugin` 接口：

- `CreateOrder`: 模拟创建订单，返回模拟支付URL
- `WaitProduct`: 模拟产品选择，返回模拟产品ID和核销ID
- `CallbackSubmit`: 模拟回调处理（可选）

插件注册在 `internal/plugin/mock/register.go` 中，通过 `init()` 函数自动注册。
