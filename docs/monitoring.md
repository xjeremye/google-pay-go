# 监控和 Swagger 文档使用指南

## Swagger API 文档

### 访问 Swagger UI

启动应用后，访问以下地址查看 API 文档：

```
http://localhost:8080/swagger/index.html
```

### 生成 Swagger 文档

#### 方式一：使用 Makefile

```bash
# 生成 Swagger 文档
make swagger

# 生成文档并运行应用
make swagger-run
```

#### 方式二：使用 swag 命令

```bash
# 安装 swag
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g main.go -o docs --parseDependency --parseInternal
```

### Swagger 注释规范

在控制器方法上添加 Swagger 注释：

```go
// CreateOrder 创建订单
// @Summary 创建订单
// @Description 创建支付订单
// @Tags 订单
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "订单信息"
// @Success 200 {object} response.Response{data=object} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/orders [post]
func (c *OrderController) CreateOrder(ctx *gin.Context) {
    // ...
}
```

### Swagger 注释说明

- `@Summary`: API 简要说明
- `@Description`: API 详细描述
- `@Tags`: API 分组标签
- `@Accept`: 接受的请求类型（json, form-data 等）
- `@Produce`: 返回的内容类型
- `@Param`: 参数说明（path, query, body, header）
- `@Success`: 成功响应说明
- `@Failure`: 失败响应说明
- `@Router`: 路由路径和方法

## Prometheus 监控

### 访问监控指标

启动应用后，访问以下地址查看 Prometheus 指标：

```
http://localhost:8080/metrics
```

### 监控指标说明

#### HTTP 请求指标

- `http_requests_total`: HTTP 请求总数（按方法、路径、状态码分组）
- `http_request_duration_seconds`: HTTP 请求持续时间（秒）
- `http_request_size_bytes`: HTTP 请求大小（字节）
- `http_response_size_bytes`: HTTP 响应大小（字节）
- `http_requests_in_flight`: 当前正在处理的请求数

#### 指标标签

- `method`: HTTP 方法（GET, POST 等）
- `path`: 请求路径
- `status`: HTTP 状态码

### Prometheus 配置示例

在 `prometheus.yml` 中添加配置：

```yaml
scrape_configs:
  - job_name: 'golang-pay-core'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
```

### Grafana 仪表板

可以使用以下 PromQL 查询创建 Grafana 仪表板：

#### QPS（每秒请求数）

```promql
rate(http_requests_total[5m])
```

#### 请求延迟（P95）

```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

#### 错误率

```promql
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])
```

#### 当前并发数

```promql
http_requests_in_flight
```

## 健康检查

### 访问健康检查端点

```
http://localhost:8080/health
```

### 响应示例

```json
{
  "status": "ok",
  "service": "golang-pay-core",
  "version": "1.0.0",
  "mode": "release",
  "database": {
    "status": "ok",
    "open_connections": 5,
    "in_use": 2,
    "idle": 3,
    "wait_count": 0
  },
  "redis": {
    "status": "ok",
    "hits": 1000,
    "misses": 50,
    "timeouts": 0,
    "total_conns": 10,
    "idle_conns": 5
  }
}
```

### 健康检查状态

- `status`: 服务状态（ok/error）
- `database`: 数据库连接状态和统计信息
- `redis`: Redis 连接状态和统计信息

## 监控最佳实践

### 1. 设置告警规则

在 Prometheus 中设置告警规则：

```yaml
groups:
  - name: golang-pay-core
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.01
        for: 5m
        annotations:
          summary: "错误率过高"
      
      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        annotations:
          summary: "请求延迟过高"
```

### 2. 监控关键指标

- **QPS**: 每秒请求数
- **延迟**: P50, P95, P99 延迟
- **错误率**: 4xx 和 5xx 错误比例
- **数据库连接池**: 连接数使用情况
- **Redis 连接池**: 连接数使用情况

### 3. 日志集成

结合应用日志和监控指标，可以更好地定位问题：

```go
logger.Logger.Info("请求处理",
    zap.String("method", method),
    zap.String("path", path),
    zap.Int("status", status),
    zap.Duration("duration", duration),
)
```

## Docker 部署监控

### docker-compose.yml 示例

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=prod
  
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
  
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

## 常见问题

### Q: Swagger 文档不显示？

A: 确保已运行 `make swagger` 生成文档，并且 `docs/swagger.json` 文件存在。

### Q: 如何更新 Swagger 文档？

A: 修改代码中的 Swagger 注释后，重新运行 `make swagger`。

### Q: Prometheus 指标为空？

A: 确保应用已启动，并且 `/metrics` 端点可以访问。检查中间件是否正确注册。

### Q: 如何自定义监控指标？

A: 在 `internal/middleware/metrics.go` 中添加自定义指标，参考现有指标的实现方式。

