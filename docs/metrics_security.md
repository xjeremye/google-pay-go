# /metrics 端点安全保护指南

## 安全风险

`/metrics` 端点暴露了应用的详细监控指标，包括：
- 请求数量、延迟、错误率
- 系统资源使用情况
- 业务指标（订单数、交易量等）

如果未加保护，可能被恶意用户利用进行：
- 系统信息收集
- 业务数据窃取
- DDoS 攻击目标识别

## 安全保护方案

项目提供了多种安全保护方案，可以根据实际需求选择：

### 方案一：Token 认证（推荐）

#### 配置方式

在配置文件中设置 Token：

```yaml
monitoring:
  metrics_token: "your-strong-secret-token"
```

#### 访问方式

**方式 1: Authorization 头**
```bash
curl -H "Authorization: Bearer your-strong-secret-token" \
  http://localhost:8080/metrics
```

**方式 2: 查询参数**
```bash
curl http://localhost:8080/metrics?token=your-strong-secret-token
```

#### Prometheus 配置

在 `monitoring/prometheus.yml` 中配置：

```yaml
scrape_configs:
  - job_name: 'golang-pay-core'
    metrics_path: '/metrics'
    params:
      token: ['your-strong-secret-token']  # 使用查询参数方式
    static_configs:
      - targets: ['localhost:8080']
```

或者使用 Bearer Token：

```yaml
scrape_configs:
  - job_name: 'golang-pay-core'
    metrics_path: '/metrics'
    bearer_token: 'your-strong-secret-token'  # 使用 Bearer Token
    static_configs:
      - targets: ['localhost:8080']
```

### 方案二：IP 白名单

#### 配置方式

```yaml
monitoring:
  metrics_ip_whitelist:
    - "127.0.0.1"           # 本地访问
    - "::1"                 # IPv6 本地访问
    - "192.168.1.100"       # 特定 IP
    - "192.168.1.0/24"      # CIDR 网段
    - "10.0.0.0/8"          # 内网段
```

#### 特点

- ✅ 简单易用
- ✅ 适合内网环境
- ⚠️ 需要维护 IP 列表
- ⚠️ 不适用于动态 IP

### 方案三：Token + IP 白名单（最安全）

同时启用两种方式，任一通过即可：

```yaml
monitoring:
  metrics_token: "your-strong-secret-token"
  metrics_ip_whitelist:
    - "127.0.0.1"
    - "192.168.1.0/24"
```

### 方案四：网络隔离

#### 使用反向代理（Nginx）

```nginx
# 只允许内网访问
location /metrics {
    allow 127.0.0.1;
    allow 192.168.0.0/16;
    allow 10.0.0.0/8;
    deny all;
    
    proxy_pass http://localhost:8080;
    proxy_set_header Host $host;
}
```

#### 使用防火墙规则

```bash
# 只允许特定 IP 访问 8080 端口的 /metrics
iptables -A INPUT -p tcp --dport 8080 -s 192.168.1.0/24 -j ACCEPT
iptables -A INPUT -p tcp --dport 8080 -j DROP
```

## 环境配置建议

### 开发环境

```yaml
monitoring:
  metrics_token: ""              # 不启用 Token（方便开发）
  metrics_ip_whitelist: []        # 不限制 IP
  swagger_enabled: true
```

### 测试环境

```yaml
monitoring:
  metrics_token: "test-token"     # 简单 Token
  metrics_ip_whitelist:
    - "127.0.0.1"
    - "::1"
  swagger_enabled: true
```

### 生产环境（推荐配置）

```yaml
monitoring:
  metrics_token: "strong-random-token-here"  # 强随机 Token
  metrics_ip_whitelist:
    - "127.0.0.1"                # 本地
    - "10.0.0.0/8"               # 内网段
    - "172.16.0.0/12"            # 内网段
    - "192.168.0.0/16"           # 内网段
  swagger_enabled: false         # 关闭 Swagger
```

## 生成强 Token

```bash
# 方式一：使用 openssl
openssl rand -hex 32

# 方式二：使用 /dev/urandom
head -c 32 /dev/urandom | base64

# 方式三：使用 Go
go run -c 'package main; import ("crypto/rand"; "encoding/hex"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(hex.EncodeToString(b)) }'
```

## Prometheus 配置更新

### 使用查询参数方式

```yaml
scrape_configs:
  - job_name: 'golang-pay-core'
    metrics_path: '/metrics'
    params:
      token: ['${METRICS_TOKEN}']  # 使用环境变量
    static_configs:
      - targets: ['localhost:8080']
```

### 使用 Bearer Token 方式

```yaml
scrape_configs:
  - job_name: 'golang-pay-core'
    metrics_path: '/metrics'
    bearer_token: '${METRICS_TOKEN}'  # 使用环境变量
    static_configs:
      - targets: ['localhost:8080']
```

### 使用环境变量

在 `docker-compose.monitoring.yml` 中：

```yaml
services:
  prometheus:
    environment:
      - METRICS_TOKEN=${METRICS_TOKEN}
    # 在配置文件中使用 ${METRICS_TOKEN}
```

## 验证安全配置

### 测试未授权访问

```bash
# 应该返回 401 未授权
curl http://localhost:8080/metrics

# 应该返回 401 未授权（错误 Token）
curl -H "Authorization: Bearer wrong-token" http://localhost:8080/metrics
```

### 测试授权访问

```bash
# 应该返回指标数据
curl -H "Authorization: Bearer your-strong-secret-token" \
  http://localhost:8080/metrics

# 或使用查询参数
curl http://localhost:8080/metrics?token=your-strong-secret-token
```

## 最佳实践

1. **生产环境必须启用认证**
   - 设置强随机 Token
   - 定期轮换 Token

2. **结合多种方式**
   - Token 认证 + IP 白名单
   - 网络隔离 + 应用层认证

3. **使用环境变量**
   - Token 不要硬编码在配置文件中
   - 使用密钥管理服务（如 Vault、AWS Secrets Manager）

4. **监控访问日志**
   - 记录所有对 `/metrics` 的访问
   - 设置异常访问告警

5. **限制暴露范围**
   - 生产环境关闭 Swagger
   - 使用内网部署 Prometheus
   - 不对外暴露 `/metrics` 端点

## 安全等级对比

| 方案 | 安全等级 | 复杂度 | 适用场景 |
|------|---------|--------|----------|
| 无保护 | ⭐ | 低 | 仅开发环境 |
| IP 白名单 | ⭐⭐ | 低 | 内网环境 |
| Token 认证 | ⭐⭐⭐ | 中 | 通用场景 |
| Token + IP | ⭐⭐⭐⭐ | 中 | 生产环境 |
| 网络隔离 | ⭐⭐⭐⭐⭐ | 高 | 高安全要求 |

## 故障排查

### Prometheus 无法抓取指标

1. 检查 Token 是否正确
2. 检查 Prometheus 配置中的认证方式
3. 查看应用日志中的认证失败记录

### IP 白名单不生效

1. 检查客户端真实 IP（可能被代理转发）
2. 检查 CIDR 格式是否正确
3. 查看应用日志中的 IP 信息

