# 单元测试指南

## 概述

项目使用 Go 标准 `testing` 包和 `testify` 库进行单元测试。

## 运行测试

### 运行所有测试

```bash
go test ./...
```

### 运行特定包的测试

```bash
go test ./internal/service
```

### 运行特定测试函数

```bash
go test ./internal/service -run TestOrderService_CreateOrder_AmountInvalid
```

### 查看测试覆盖率

```bash
go test ./internal/service -cover
```

### 生成详细覆盖率报告

```bash
go test ./internal/service -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 测试结构

### 订单服务测试

测试文件：`internal/service/order_create_test.go`

#### 已实现的测试用例

1. **TestOrderService_CreateOrder_AmountInvalid**
   - 测试金额无效的情况
   - 验证错误码和错误消息

2. **TestOrderService_validateMerchant**
   - 测试商户验证逻辑
   - 包括商户不存在、商户已禁用等情况

3. **TestOrderService_validateSign**
   - 测试签名验证
   - 包括签名正确、签名错误、缺少签名等情况

4. **TestOrderService_validateOutOrderNo**
   - 测试商户订单号验证
   - 包括订单号为空、订单号已存在等情况

5. **TestOrderService_validateChannel**
   - 测试渠道验证
   - 包括渠道不存在、渠道已禁用、金额范围检查等

6. **TestOrderService_validatePlugin**
   - 测试插件验证逻辑

7. **TestOrderService_checkChannelTime**
   - 测试渠道时间检查
   - 包括全天可用、正常时间范围、跨零点等情况

8. **TestOrderService_checkChannelAmount**
   - 测试渠道金额检查
   - 包括固定金额模式、金额范围检查等

9. **TestOrderService_applyFloatAmount**
   - 测试浮动加价逻辑

10. **TestOrderService_buildResponse**
    - 测试响应构建
    - 包括标准模式和兼容模式

11. **TestOrderService_GetOrderByOrderNo**
    - 测试根据订单号获取订单

12. **TestOrderService_GetOrderByOutOrderNo**
    - 测试根据商户订单号获取订单

## 测试工具

### 内存数据库

测试使用 SQLite 内存数据库，避免依赖外部数据库：

```go
db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
```

### 测试辅助函数

- `setupTestDB(t *testing.T) *gorm.DB`：设置测试数据库
- `setupTestRedis()`：设置测试 Redis（可选）

## 编写新测试

### 基本结构

```go
func TestOrderService_NewFeature(t *testing.T) {
    // 1. 设置测试环境
    db := setupTestDB(t)
    originalDB := database.DB
    database.DB = db
    defer func() {
        database.DB = originalDB
    }()

    // 2. 准备测试数据
    service := NewOrderService()
    
    // 3. 执行测试
    result, err := service.SomeMethod(...)
    
    // 4. 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 测试最佳实践

1. **独立性**：每个测试应该独立，不依赖其他测试
2. **可重复性**：测试应该可以重复运行，结果一致
3. **快速执行**：使用内存数据库和 Mock，避免外部依赖
4. **清晰命名**：测试函数名应该清晰描述测试内容
5. **完整覆盖**：测试应该覆盖正常流程和异常流程

## Mock 对象

### Redis Mock

对于需要 Redis 的测试，可以使用：
- 真实的 Redis 实例（测试环境）
- 或者使用 `miniredis` 等测试库

### 数据库 Mock

使用 SQLite 内存数据库，每个测试独立数据库实例。

## 持续集成

建议在 CI/CD 流程中运行测试：

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: go test ./... -v -cover
```

## 注意事项

1. **数据库状态**：每个测试应该清理自己的数据，或使用独立的数据库实例
2. **并发安全**：如果测试涉及并发，确保测试是线程安全的
3. **资源清理**：使用 `defer` 确保资源被正确清理
4. **测试数据**：使用有意义的测试数据，避免魔法数字

