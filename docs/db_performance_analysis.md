# 数据库性能瓶颈分析与优化

## 📊 性能瓶颈分析总结

### 1. 已优化的性能问题

#### ✅ UpdateOrderStatus 方法优化

**优化前的问题：**
- 需要 2 次数据库查询：
  1. 查询订单信息（SELECT）
  2. 查询商户获取租户ID（SELECT）
- 增加了数据库往返次数和事务时间

**优化后：**
- 使用 JOIN 查询，一次性获取订单和租户ID
- 减少 1 次数据库查询
- 减少事务时间，降低锁竞争风险

**优化代码：**
```go
// 使用 LEFT JOIN 一次性获取订单信息和租户ID
tx.Table("dvadmin_order").
    Select("dvadmin_order.id as order_id, dvadmin_order.merchant_id, dvadmin_order.money, dvadmin_order.order_status, dvadmin_merchant.parent_id as tenant_id").
    Joins("LEFT JOIN dvadmin_merchant ON dvadmin_order.merchant_id = dvadmin_merchant.id").
    Where("dvadmin_order.id = ?", orderID).
    First(&orderInfo)
```

#### ✅ 减少锁持有时间

**优化前：**
- SELECT FOR UPDATE 查询租户信息
- 然后更新租户余额
- 锁持有时间较长

**优化后：**
- 直接使用原子操作更新（UPDATE ... SET balance = balance - ?）
- 不需要先查询再更新
- 减少锁持有时间

**优化代码：**
```go
// 直接使用原子操作，避免先查询再更新
tx.Model(&models.Tenant{}).
    Where("id = ?", *tenantID).
    Updates(map[string]interface{}{
        "balance": gorm.Expr("balance - ?", orderInfo.Money),
        "pre_tax": gorm.Expr("pre_tax - ?", orderInfo.Money),
    })
```

### 2. 当前性能特征

#### 订单创建流程（createOrderAndDetail）

**数据库操作：**
1. SELECT FOR UPDATE 锁定租户行（检查余额）
2. INSERT 创建订单
3. INSERT 创建订单详情
4. UPDATE 增加预占余额（原子操作）

**性能特点：**
- ✅ 使用事务确保一致性
- ✅ 使用原子操作避免并发问题
- ✅ SELECT FOR UPDATE 锁定时间较短（只查询必要字段）
- ⚠️ 同一租户的并发订单会有锁竞争（这是必要的，确保余额一致性）

#### 订单状态更新流程（UpdateOrderStatus）

**数据库操作（优化后）：**
1. SELECT ... JOIN 查询订单和租户ID（1次查询）
2. UPDATE 更新租户余额（原子操作，无需先查询）
3. UPDATE 更新订单版本号
4. UPDATE 更新订单状态
5. UPDATE 更新订单详情（可选）

**性能特点：**
- ✅ 使用 JOIN 减少查询次数
- ✅ 使用原子操作减少锁持有时间
- ✅ 事务时间较短
- ⚠️ 同一租户的并发订单状态更新会有锁竞争（这是必要的）

### 3. 潜在性能瓶颈

#### ⚠️ 租户行锁竞争

**问题：**
- 同一租户的并发订单创建/更新会竞争同一行锁
- 高并发场景下，可能导致锁等待

**影响：**
- 订单创建/更新延迟增加
- 数据库连接占用时间增加

**缓解措施：**
1. ✅ 已优化：减少锁持有时间（使用原子操作，避免先查询）
2. ✅ 已优化：减少事务时间（使用 JOIN 减少查询）
3. 💡 建议：监控锁等待时间，设置告警
4. 💡 建议：考虑读写分离（读操作使用从库）

#### ⚠️ 数据库连接池压力

**问题：**
- 每个事务占用一个数据库连接
- 高并发时可能耗尽连接池

**当前配置：**
- `max_open_conns: 100`（根据配置文件）

**建议：**
- 根据实际并发量调整连接池大小
- 监控连接池使用率
- 考虑使用连接池监控

### 4. 索引建议

#### 关键索引检查

**订单表（dvadmin_order）：**
```sql
-- 检查现有索引
SHOW INDEX FROM dvadmin_order;

-- 建议索引（如果不存在）：
-- 1. id (主键) - 已存在
-- 2. order_no (唯一索引) - 已存在
-- 3. merchant_id (索引) - 已存在
-- 4. order_status (索引) - 已存在
-- 5. create_datetime (索引) - 已存在
```

**租户表（dvadmin_tenant）：**
```sql
-- 检查现有索引
SHOW INDEX FROM dvadmin_tenant;

-- 建议索引（如果不存在）：
-- 1. id (主键) - 已存在
-- 2. 确保 balance, pre_tax 字段有适当的索引（如果经常查询）
```

**商户表（dvadmin_merchant）：**
```sql
-- 检查现有索引
SHOW INDEX FROM dvadmin_merchant;

-- 建议索引（如果不存在）：
-- 1. id (主键) - 已存在
-- 2. parent_id (索引) - 用于 JOIN 查询，建议添加
```

**建议执行：**
```sql
-- 如果 parent_id 没有索引，添加索引
CREATE INDEX idx_merchant_parent_id ON dvadmin_merchant(parent_id);
```

### 5. 性能监控建议

#### 关键指标

1. **事务执行时间**
   - 订单创建事务：目标 < 50ms
   - 订单状态更新事务：目标 < 30ms

2. **锁等待时间**
   - SELECT FOR UPDATE 等待时间：目标 < 10ms
   - 监控 `SHOW ENGINE INNODB STATUS` 中的锁等待信息

3. **数据库连接池使用率**
   - 目标：使用率 < 80%
   - 告警阈值：使用率 > 90%

4. **慢查询**
   - 启用慢查询日志
   - 监控执行时间 > 100ms 的查询

#### 监控 SQL

```sql
-- 查看当前连接数
SHOW STATUS LIKE 'Threads_connected';

-- 查看最大连接数
SHOW VARIABLES LIKE 'max_connections';

-- 查看锁等待
SHOW ENGINE INNODB STATUS;

-- 查看慢查询
SHOW VARIABLES LIKE 'slow_query_log';
SHOW VARIABLES LIKE 'long_query_time';
```

### 6. 进一步优化建议

#### 💡 短期优化（1-2周）

1. **添加索引**
   - 确保 `dvadmin_merchant.parent_id` 有索引
   - 根据慢查询日志添加缺失索引

2. **监控和告警**
   - 设置数据库性能监控
   - 设置锁等待时间告警
   - 设置连接池使用率告警

3. **查询优化**
   - 定期检查慢查询日志
   - 优化执行时间 > 100ms 的查询

#### 💡 中期优化（1-3月）

1. **读写分离**
   - 读操作使用从库（查询订单、查询余额等）
   - 写操作使用主库（创建订单、更新状态等）

2. **缓存优化**
   - 增加缓存命中率
   - 优化缓存失效策略

3. **批量操作优化**
   - 如果有多订单批量操作，考虑批量更新

#### 💡 长期优化（3-6月）

1. **分库分表**
   - 如果订单量非常大，考虑按租户分库
   - 按时间分表（历史订单归档）

2. **异步处理**
   - 非关键操作异步化（如订单日志）
   - 使用消息队列处理订单状态更新

3. **数据库优化**
   - 优化 MySQL 配置参数
   - 考虑使用 MySQL 8.0+ 的新特性

### 7. 性能测试建议

#### 压测场景

1. **订单创建压测**
   - 目标：1000 QPS
   - 监控：事务执行时间、锁等待时间、连接池使用率

2. **订单状态更新压测**
   - 目标：500 QPS
   - 监控：事务执行时间、锁等待时间

3. **混合场景压测**
   - 订单创建 + 状态更新
   - 模拟真实业务场景

#### 压测工具

- 使用 Apache Bench (ab) 或 wrk
- 使用 Go 的压测工具（如 vegeta）
- 监控数据库性能指标

### 8. 总结

#### ✅ 已完成的优化

1. ✅ 使用 JOIN 减少查询次数
2. ✅ 使用原子操作减少锁持有时间
3. ✅ 优化事务执行顺序
4. ✅ 减少数据库往返次数

#### ⚠️ 需要注意的问题

1. ⚠️ 租户行锁竞争（高并发场景）
2. ⚠️ 数据库连接池压力
3. ⚠️ 需要监控和告警

#### 💡 建议的下一步

1. 💡 添加索引（特别是 merchant.parent_id）
2. 💡 设置性能监控和告警
3. 💡 进行压测，验证优化效果
4. 💡 根据压测结果进一步优化

---

**最后更新：** 2024-01-XX  
**审查人：** AI Assistant  
**状态：** ✅ 已优化主要性能瓶颈
