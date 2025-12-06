# 监控架构和数据流

## 监控数据流

```
┌─────────────────┐
│   应用服务       │
│  :8080          │
│  /metrics       │ ← 暴露 Prometheus 格式的指标
└────────┬────────┘
         │
         │ HTTP GET /metrics
         │ (每 5-15 秒抓取一次)
         ▼
┌─────────────────┐
│  Prometheus     │
│  :9090          │ ← 抓取并存储指标数据
│  (时间序列数据库)│
└────────┬────────┘
         │
         │ PromQL 查询
         │ (实时查询)
         ▼
    ┌────────┴────────┐
    │                 │
    ▼                 ▼
┌─────────┐    ┌──────────────┐
│ Grafana │    │ Alertmanager │
│ :3000   │    │ :9093        │
│ (可视化)│    │ (告警管理)    │
└─────────┘    └──────────────┘
```

## 详细说明

### 1. 应用层 - `/metrics` 端点

**位置**: `http://localhost:8080/metrics`

**作用**: 
- 应用暴露 Prometheus 格式的指标数据
- 由 `internal/middleware/metrics.go` 中间件自动收集
- 由 `internal/middleware/prometheus.go` 提供 HTTP 端点

**数据格式**: Prometheus 文本格式
```
http_requests_total{method="POST",path="/api/v1/orders",status="200"} 1234
http_request_duration_seconds_bucket{method="POST",path="/api/v1/orders",le="0.1"} 1000
http_requests_in_flight 5
```

### 2. Prometheus 服务器

**位置**: `http://localhost:9090`

**作用**:
- 定期抓取应用的 `/metrics` 端点（默认每 15 秒）
- 存储时间序列数据
- 执行告警规则评估
- 提供 PromQL 查询接口

**配置**: `monitoring/prometheus.yml`
```yaml
scrape_configs:
  - job_name: 'golang-pay-core'
    scrape_interval: 5s          # 抓取间隔
    metrics_path: '/metrics'    # 指标路径
    static_configs:
      - targets: ['localhost:8080']  # 应用地址
```

### 3. Grafana

**位置**: `http://localhost:3000`

**作用**:
- 从 Prometheus 读取数据（不是直接从 `/metrics`）
- 可视化展示（图表、仪表盘）
- 配置告警规则
- 发送告警通知

**数据源配置**: 
- 类型: Prometheus
- URL: `http://prometheus:9090` (Docker) 或 `http://localhost:9090` (本地)

### 4. Alertmanager

**位置**: `http://localhost:9093`

**作用**:
- 接收来自 Prometheus 的告警
- 路由和分组告警
- 发送通知（邮件、钉钉等）

## 关键点

### ✅ 是的，基于 `/metrics` 端点

1. **应用暴露指标**: `/metrics` 端点提供原始指标数据
2. **Prometheus 抓取**: Prometheus 定期访问 `/metrics` 获取数据
3. **Grafana 可视化**: Grafana 从 Prometheus 读取数据并展示

### 数据流向

```
应用 /metrics 
  → Prometheus 抓取和存储 
    → Grafana 查询和可视化
    → Alertmanager 告警评估
```

### 为什么需要 Prometheus？

1. **数据存储**: Prometheus 是时间序列数据库，可以存储历史数据
2. **查询能力**: 支持 PromQL 复杂查询
3. **告警评估**: 可以基于历史数据评估告警规则
4. **数据聚合**: 可以对多个实例的数据进行聚合

### 直接访问 vs 通过 Prometheus

| 方式 | 优点 | 缺点 |
|------|------|------|
| **直接访问 `/metrics`** | 实时数据，无需额外服务 | 无历史数据，无查询能力，无告警 |
| **通过 Prometheus** | 历史数据，强大查询，告警功能 | 需要额外服务，有延迟（抓取间隔） |

## 验证数据流

### 1. 检查应用指标端点

```bash
# 直接访问指标端点
curl http://localhost:8080/metrics

# 应该看到类似输出：
# http_requests_total{method="GET",path="/health",status="200"} 1
# http_request_duration_seconds_bucket{method="GET",path="/health",le="0.005"} 1
```

### 2. 检查 Prometheus 是否抓取成功

访问 Prometheus UI: http://localhost:9090

- 进入 **Status** > **Targets**
- 查看 `golang-pay-core` 的状态应该是 **UP**
- 如果显示 **DOWN**，检查配置中的 targets 地址

### 3. 在 Prometheus 中查询

在 Prometheus UI 的查询框中输入：
```promql
http_requests_total
```

应该能看到指标数据。

### 4. 在 Grafana 中查看

1. 登录 Grafana: http://localhost:3000
2. 创建新的 Dashboard
3. 添加 Panel，选择 Prometheus 数据源
4. 输入查询: `rate(http_requests_total[5m])`
5. 应该能看到图表

## 常见问题

### Q: 为什么 Grafana 看不到数据？

A: 检查以下几点：
1. 应用是否运行并暴露 `/metrics` 端点
2. Prometheus 是否成功抓取（检查 Targets 页面）
3. Grafana 数据源配置是否正确
4. 查询的时间范围是否正确

### Q: 可以直接在 Grafana 中连接 `/metrics` 吗？

A: 不可以。Grafana 需要时间序列数据库作为数据源，不能直接连接 HTTP 端点。必须通过 Prometheus。

### Q: Prometheus 抓取间隔是多少？

A: 默认 15 秒，可以在 `prometheus.yml` 中配置 `scrape_interval`。

### Q: 数据有延迟吗？

A: 是的，最多延迟一个抓取间隔（默认 15 秒）。这是正常的，因为 Prometheus 是定期抓取，不是实时推送。

## 总结

**是的，监控 UI 完全基于 `/metrics` 端点**，但需要通过 Prometheus 作为中间层：

1. ✅ 应用通过 `/metrics` 暴露指标
2. ✅ Prometheus 抓取 `/metrics` 并存储
3. ✅ Grafana 从 Prometheus 读取数据并可视化
4. ✅ Alertmanager 基于 Prometheus 数据评估告警

这种架构的优势：
- 历史数据存储
- 强大的查询能力
- 灵活的告警规则
- 可扩展性（可以监控多个服务实例）

