# 性能优化配置指南

## 1. 高并发配置推荐

### 1.1 生产环境推荐配置

创建 `config/config.prod.yaml`:

```yaml
# 应用配置
app:
  name: golang-pay-core
  version: 1.0.0
  mode: release
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

# 数据库配置（高并发优化）
database:
  host: localhost
  port: 3306
  user: root
  password: ""
  dbname: django_vue3_admin
  charset: utf8mb4
  max_idle_conns: 50        # 增加到 50
  max_open_conns: 500       # 增加到 500（主要优化点）
  conn_max_lifetime: 3600s
  log_mode: false           # 生产环境关闭 SQL 日志

# Redis 配置（高并发优化）
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 50             # 增加到 50
  min_idle_conns: 20        # 增加到 20

# 日志配置
log:
  level: info
  format: json
  output: file              # 生产环境输出到文件
  file_path: logs/app.log
  max_size: 500             # 增加到 500MB
  max_backups: 30           # 保留 30 个备份
  max_age: 90               # 保留 90 天
  compress: true
```

### 1.2 MySQL 服务器优化

**检查当前配置：**
```sql
-- 查看最大连接数
SHOW VARIABLES LIKE 'max_connections';

-- 查看当前连接数
SHOW STATUS LIKE 'Threads_connected';

-- 查看最大使用连接数
SHOW STATUS LIKE 'Max_used_connections';
```

**推荐配置（my.cnf）：**
```ini
[mysqld]
# 最大连接数（根据服务器内存调整）
max_connections = 1000

# 连接超时时间
wait_timeout = 600
interactive_timeout = 600

# 查询缓存（MySQL 5.7 及以下）
query_cache_size = 256M
query_cache_type = 1

# InnoDB 缓冲池（建议设置为内存的 70-80%）
innodb_buffer_pool_size = 4G

# InnoDB 日志文件大小
innodb_log_file_size = 512M

# 表打开缓存
table_open_cache = 4000

# 临时表大小
tmp_table_size = 256M
max_heap_table_size = 256M
```

### 1.3 Redis 服务器优化

**检查当前配置：**
```bash
redis-cli CONFIG GET maxclients
redis-cli INFO clients
```

**推荐配置（redis.conf）：**
```conf
# 最大客户端连接数
maxclients 10000

# 内存策略
maxmemory 4gb
maxmemory-policy allkeys-lru

# 持久化配置（根据需求选择）
# RDB 快照
save 900 1
save 300 10
save 60 10000

# AOF 持久化（可选）
appendonly yes
appendfsync everysec
```

## 2. 代码层面优化

### 2.1 数据库连接池预热

在 `internal/database/mysql.go` 中添加预热功能：

```go
// WarmupMySQL 预热数据库连接池
func WarmupMySQL() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 预热连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// 创建一些连接以填充连接池
	maxIdle := config.Cfg.Database.MaxIdleConns
	for i := 0; i < maxIdle; i++ {
		conn, err := sqlDB.Conn(context.Background())
		if err != nil {
			return err
		}
		conn.Close()
	}

	return nil
}
```

### 2.2 Redis 连接池预热

在 `internal/database/redis.go` 中添加预热功能：

```go
// WarmupRedis 预热 Redis 连接池
func WarmupRedis() error {
	if RDB == nil {
		return fmt.Errorf("Redis 未初始化")
	}

	ctx := context.Background()
	// 执行一些简单操作以建立连接
	for i := 0; i < config.Cfg.Redis.MinIdleConns; i++ {
		if err := RDB.Ping(ctx).Err(); err != nil {
			return err
		}
	}

	return nil
}
```

### 2.3 增加缓存使用

在 `internal/service/merchant_service.go` 中添加缓存：

```go
// GetMerchantByID 根据ID获取商户（带缓存）
func (s *MerchantService) GetMerchantByID(id int64) (*models.Merchant, error) {
	cacheKey := utils.GetCacheKey("merchant", fmt.Sprintf("%d", id))
	
	// 尝试从缓存获取
	var merchant models.Merchant
	cache := &utils.Cache{}
	ctx := database.GetContext()
	
	if err := cache.Get(ctx, cacheKey, &merchant); err == nil {
		return &merchant, nil
	}

	// 缓存未命中，从数据库查询
	merchant, err := s.getMerchantFromDB(id)
	if err != nil {
		return nil, err
	}

	// 写入缓存（TTL: 5分钟）
	cache.Set(ctx, cacheKey, merchant, 5*time.Minute)
	
	return &merchant, nil
}
```

## 3. 监控和告警

### 3.1 添加性能监控中间件

创建 `internal/middleware/metrics.go`:

```go
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

var (
	requestCount    int64
	requestDuration time.Duration
)

// Metrics 性能监控中间件
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start)
		requestCount++
		requestDuration += duration
	}
}

// GetMetrics 获取性能指标
func GetMetrics() map[string]interface{} {
	avgDuration := time.Duration(0)
	if requestCount > 0 {
		avgDuration = requestDuration / time.Duration(requestCount)
	}
	
	return map[string]interface{}{
		"request_count":    requestCount,
		"avg_duration_ms": avgDuration.Milliseconds(),
	}
}
```

### 3.2 健康检查端点增强

在 `internal/router/router.go` 中添加：

```go
// 健康检查（包含性能指标）
r.GET("/health", func(c *gin.Context) {
	sqlDB, _ := database.DB.DB()
	dbStats := sqlDB.Stats()
	
	c.JSON(200, gin.H{
		"status": "ok",
		"service": config.Cfg.App.Name,
		"version": config.Cfg.App.Version,
		"database": gin.H{
			"open_connections": dbStats.OpenConnections,
			"in_use": dbStats.InUse,
			"idle": dbStats.Idle,
			"wait_count": dbStats.WaitCount,
		},
		"redis": gin.H{
			"pool_size": database.RDB.PoolSize(),
			"pool_stats": database.RDB.PoolStats(),
		},
		"metrics": middleware.GetMetrics(),
	})
})
```

## 4. 压力测试脚本

### 4.1 wrk 测试脚本

创建 `scripts/test_create_order.lua`:

```lua
-- 创建订单压测脚本
math.randomseed(os.time())

request = function()
    local order_no = "TEST" .. math.random(1000000, 9999999)
    local body = string.format([[
{
    "out_order_no": "%s",
    "money": 10000,
    "tax": 100,
    "notify_url": "https://example.com/notify",
    "jump_url": "https://example.com/jump",
    "notify_money": 10000,
    "pay_channel_id": 1
}
]], order_no)
    
    return wrk.format("POST", "/api/v1/orders", {
        ["Content-Type"] = "application/json"
    }, body)
end
```

### 4.2 运行压力测试

```bash
# 创建订单压测
wrk -t12 -c400 -d30s --script=scripts/test_create_order.lua \
    --latency http://localhost:8080

# 查询订单压测
wrk -t12 -c1000 -d30s --latency \
    http://localhost:8080/api/v1/orders/PAY20240101000001
```

## 5. 部署建议

### 5.1 单实例部署

**服务器配置建议：**
- CPU: 4-8 核
- 内存: 8-16 GB
- 预期并发: 1,000 - 2,000 QPS

### 5.2 多实例部署

**负载均衡配置（Nginx）：**
```nginx
upstream pay_api {
    least_conn;  # 最少连接算法
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
}

server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://pay_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        
        # 超时设置
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
}
```

**预期并发：**
- 3 实例: 3,000 - 6,000 QPS
- 5 实例: 5,000 - 10,000 QPS

## 6. 性能调优检查清单

- [ ] 数据库连接池配置优化（max_open_conns >= 500）
- [ ] Redis 连接池配置优化（pool_size >= 50）
- [ ] MySQL 服务器 max_connections 配置
- [ ] Redis 服务器 maxclients 配置
- [ ] 启用缓存（商户、支付通道等）
- [ ] 数据库索引优化
- [ ] SQL 查询优化
- [ ] 异步处理非关键操作
- [ ] 日志级别调整（生产环境使用 info）
- [ ] 连接池预热
- [ ] 性能监控和告警
- [ ] 压力测试和性能基准
- [ ] 负载均衡配置（多实例）

