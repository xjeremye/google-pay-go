# 压测工具使用指南

本目录包含用于支付系统上线前压测的工具和脚本。

## 目录结构

```
loadtest/
├── k6/                    # k6 压测脚本
│   ├── create_order.js           # 创建订单（GET方式）
│   ├── create_order_post.js      # 创建订单（POST方式）
│   ├── health_check.js           # 健康检查
│   └── mixed_scenario.js         # 混合场景
├── wrk/                   # wrk 压测脚本
│   ├── create_order.lua          # 创建订单（GET方式）
│   └── health_check.lua          # 健康检查
├── results/              # 压测结果目录（自动创建）
├── config.env.example    # 配置文件示例
├── run_k6.sh            # k6 压测运行脚本
├── run_wrk.sh           # wrk 压测运行脚本
├── analyze_results.sh   # 结果分析脚本
└── README.md            # 本文档
```

## 前置要求

### k6 安装

**macOS:**
```bash
brew install k6
```

**Linux:**
```bash
# Ubuntu/Debian
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

**Windows:**
```bash
choco install k6
```

### wrk 安装

**macOS:**
```bash
brew install wrk
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install wrk

# 或从源码编译
git clone https://github.com/wg/wrk.git
cd wrk
make
sudo cp wrk /usr/local/bin/
```

## 快速开始

### 1. 配置压测参数

复制配置文件并修改：

```bash
cp config.env.example config.env
```

编辑 `config.env`，设置以下参数：

```bash
# 服务器地址
BASE_URL=http://localhost:8080

# 商户配置
MERCHANT_ID=1
CHANNEL_ID=1
MERCHANT_KEY=your_merchant_key_here
```

### 2. 运行 k6 压测

**创建订单压测（GET方式）:**
```bash
chmod +x run_k6.sh
./run_k6.sh create_order
```

**创建订单压测（POST方式）:**
```bash
./run_k6.sh create_order_post
```

**健康检查压测:**
```bash
./run_k6.sh health_check
```

**混合场景压测:**
```bash
./run_k6.sh mixed_scenario
```

**指定服务器地址:**
```bash
./run_k6.sh create_order http://your-server:8080
```

### 3. 运行 wrk 压测

**健康检查压测:**
```bash
chmod +x run_wrk.sh
./run_wrk.sh health_check
```

**创建订单压测:**
```bash
./run_wrk.sh create_order http://localhost:8080 4 100 30s
```

参数说明：
- `create_order`: 场景名称
- `http://localhost:8080`: 服务器地址
- `4`: 线程数
- `100`: 连接数
- `30s`: 持续时间

### 4. 分析压测结果

**查看所有结果文件:**
```bash
chmod +x analyze_results.sh
./analyze_results.sh
```

**分析特定结果文件:**
```bash
./analyze_results.sh results/create_order_20240101_120000.json
```

## 压测场景说明

### 1. 创建订单（GET方式）

- **场景文件**: `k6/create_order.js`
- **说明**: 模拟通过GET请求创建订单
- **压测阶段**:
  - 预热: 30秒内增加到50并发
  - 稳定: 1分钟保持100并发
  - 增长: 2分钟增加到200并发
  - 峰值: 2分钟增加到300并发
  - 下降: 1分钟降到200并发
  - 冷却: 30秒降到0并发

### 2. 创建订单（POST方式）

- **场景文件**: `k6/create_order_post.js`
- **说明**: 模拟通过POST请求创建订单
- **压测阶段**: 与GET方式相同

### 3. 健康检查

- **场景文件**: `k6/health_check.js`
- **说明**: 测试健康检查接口的性能
- **压测阶段**:
  - 快速预热: 10秒内增加到100并发
  - 稳定: 1分钟保持500并发
  - 增长: 2分钟增加到1000并发
  - 下降: 1分钟降到500并发
  - 冷却: 10秒降到0并发

### 4. 混合场景

- **场景文件**: `k6/mixed_scenario.js`
- **说明**: 混合多种操作（70%创建订单，20%查询订单，10%健康检查）
- **压测阶段**:
  - 预热: 30秒内增加到50并发
  - 稳定: 2分钟保持150并发
  - 增长: 3分钟增加到250并发
  - 下降: 2分钟降到150并发
  - 冷却: 30秒降到0并发

## 性能指标说明

### k6 指标

- **http_req_duration**: HTTP请求响应时间
  - `avg`: 平均响应时间
  - `min`: 最小响应时间
  - `max`: 最大响应时间
  - `p(95)`: 95%的请求响应时间
  - `p(99)`: 99%的请求响应时间

- **http_reqs**: HTTP请求统计
  - `count`: 总请求数
  - `rate`: 每秒请求数（QPS）

- **http_req_failed**: HTTP请求失败率
  - `rate`: 失败率
  - `fails`: 失败请求数

### 阈值设置

默认阈值（可在脚本中修改）:
- `http_req_duration`: P95 < 500ms, P99 < 1000ms
- `http_req_failed`: 错误率 < 1%

## 压测建议

### 1. 压测前准备

- [ ] 确保测试环境与生产环境配置相似
- [ ] 检查数据库连接池配置
- [ ] 检查Redis连接池配置
- [ ] 确保有足够的测试数据
- [ ] 监控服务器资源（CPU、内存、网络）

### 2. 压测步骤

1. **基础压测**: 先运行健康检查压测，验证服务器基础性能
2. **单接口压测**: 对每个主要接口进行单独压测
3. **混合场景压测**: 模拟真实业务场景
4. **峰值压测**: 逐步增加并发，找到系统瓶颈

### 3. 压测监控

建议同时监控以下指标：

- **应用层**:
  - QPS（每秒请求数）
  - 响应时间（平均、P95、P99）
  - 错误率

- **系统层**:
  - CPU使用率
  - 内存使用率
  - 网络IO
  - 磁盘IO

- **数据库层**:
  - 连接数
  - 慢查询
  - 锁等待

- **Redis层**:
  - 连接数
  - 命中率
  - 内存使用

### 4. 压测结果分析

1. **查看响应时间分布**: 确认P95和P99是否在可接受范围内
2. **查看错误率**: 确认错误率是否低于阈值
3. **查看QPS**: 确认系统能处理的峰值QPS
4. **查看资源使用**: 确认是否有资源瓶颈

### 5. 优化建议

根据压测结果，可能的优化方向：

- **数据库优化**:
  - 增加连接池大小
  - 优化慢查询
  - 添加索引
  - 读写分离

- **缓存优化**:
  - 增加缓存命中率
  - 优化缓存策略
  - 增加Redis连接池

- **代码优化**:
  - 减少数据库查询
  - 优化业务逻辑
  - 异步处理非关键操作

- **架构优化**:
  - 水平扩展（多实例）
  - 负载均衡
  - 消息队列（异步处理）

## 常见问题

### 1. 签名验证失败

确保 `MERCHANT_KEY` 配置正确，并且签名算法与后端一致。

### 2. 连接数不足

增加数据库和Redis的连接池大小，参考 `config/config.prod.yaml` 中的配置建议。

### 3. 响应时间过长

- 检查数据库慢查询
- 检查Redis连接
- 检查网络延迟
- 优化业务逻辑

### 4. 错误率过高

- 检查服务器日志
- 检查数据库连接
- 检查资源使用情况
- 逐步降低并发数，找到稳定点

## Mock 插件说明

压测时通过配置 **mock 插件**（模拟插件）来模拟真实下单流程，排查系统问题和性能瓶颈：

- **数据库配置**: 需要在数据库中创建使用 mock 插件的支付通道
- **模拟延迟**: mock 插件会模拟支付宝网关的延迟（默认 50-200ms）
- **不调用真实API**: 不会调用真实的支付宝接口，避免产生费用
- **模拟回调**: 自动模拟支付回调，完整模拟真实下单流程
- **真实流程**: 使用正常的订单创建接口，走完整的业务逻辑

### Mock 插件特性

- **CreateOrder**: 模拟支付宝网关延迟，返回模拟支付URL
- **WaitProduct**: 返回模拟产品ID和核销ID
- **CallbackSubmit**: 模拟回调处理延迟

### 配置 Mock 插件延迟

如果需要调整模拟延迟，可以修改 `internal/plugin/mock/mock_plugin.go`：

```go
MinDelay: 50,  // 最小延迟（毫秒）
MaxDelay: 200, // 最大延迟（毫秒）
```

## 使用 Mock 插件

### 1. 数据库配置

在数据库中创建使用 mock 插件的支付通道：

1. **创建插件**：在 `pay_plugin` 表中创建插件，`name` 设置为 "Mock插件"，`status` 为 1
2. **创建支付类型**：在 `pay_type` 表中创建支付类型，`key` 设置为 `mock`，关联到插件ID
3. **创建支付通道**：在 `pay_channel` 表中创建通道，`plugin_id` 关联到插件ID，`status` 为 1

### 2. 压测配置

在 `config.env` 中配置使用 mock 插件的通道ID：

```bash
CHANNEL_ID=<mock通道ID>  # 使用mock插件的通道ID
```

### 3. 运行压测

压测脚本会使用正常的 `/api/v1/orders` 接口，通过配置的通道ID自动使用 mock 插件：

```bash
./run_k6.sh create_order
```

## 注意事项

1. **压测环境**: 建议在独立的测试环境进行压测，避免影响生产环境
2. **数据清理**: 压测会产生大量测试数据，需要定期清理
3. **资源监控**: 压测时要密切监控服务器资源，避免系统崩溃
4. **逐步加压**: 不要一开始就使用高并发，应该逐步增加
5. **签名算法**: 确保压测脚本中的签名算法与后端一致
6. **Mock插件**: 通过数据库配置使用 mock 插件，模拟真实下单流程但不调用真实支付接口
7. **真实流程**: 压测走完整的业务逻辑，包括订单创建、产品选择、回调处理等

## 参考资源

- [k6 官方文档](https://k6.io/docs/)
- [wrk 官方文档](https://github.com/wg/wrk)
- [性能测试最佳实践](https://k6.io/docs/test-types/load-testing/)
