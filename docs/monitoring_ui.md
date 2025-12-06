# 监控告警 UI 界面指南

## 监控 UI 选项

当前项目集成了 Prometheus 监控，有以下 UI 界面可用：

1. **Prometheus Web UI** - Prometheus 自带的查询界面
2. **Grafana** - 强大的可视化仪表板和告警界面（推荐）
3. **Alertmanager** - 告警管理界面

## 1. Prometheus Web UI

### 功能特点

- ✅ 查询和可视化 Prometheus 指标
- ✅ 执行 PromQL 查询
- ✅ 查看告警规则状态
- ✅ 查看目标（targets）状态
- ✅ 内置图形展示

### 访问方式

启动 Prometheus 后，访问：
```
http://localhost:9090
```

### 基本查询示例

在 Prometheus UI 的查询框中输入：

```promql
# QPS（每秒请求数）
rate(http_requests_total[5m])

# P95 延迟
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# 错误率
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])

# 当前并发数
http_requests_in_flight
```

## 2. Grafana（推荐）

### 功能特点

- ✅ 丰富的可视化图表（折线图、柱状图、仪表盘等）
- ✅ 自定义仪表板
- ✅ 告警通知（邮件、钉钉、企业微信、Slack 等）
- ✅ 模板变量和动态查询
- ✅ 数据源集成（Prometheus、MySQL、Redis 等）

### 快速开始

#### 方式一：Docker Compose（推荐）

创建 `docker-compose.monitoring.yml`:

```yaml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./monitoring/alerts.yml:/etc/prometheus/alerts.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards
    restart: unless-stopped
    depends_on:
      - prometheus

  alertmanager:
    image: prom/alertmanager:latest
    container_name: alertmanager
    ports:
      - "9093:9093"
    volumes:
      - ./monitoring/alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
    restart: unless-stopped

volumes:
  prometheus_data:
  grafana_data:
  alertmanager_data:
```

启动监控栈：
```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

#### 方式二：单独安装

**安装 Prometheus:**
```bash
# 下载 Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar xvfz prometheus-*.tar.gz
cd prometheus-*

# 启动
./prometheus --config.file=prometheus.yml
```

**安装 Grafana:**
```bash
# Ubuntu/Debian
sudo apt-get install -y software-properties-common
sudo add-apt-repository "deb https://packages.grafana.com/oss/deb stable main"
sudo apt-get update
sudo apt-get install grafana

# 启动
sudo systemctl start grafana-server
sudo systemctl enable grafana-server
```

### 访问 Grafana

启动后访问：
```
http://localhost:3000
```

默认账号：
- 用户名: `admin`
- 密码: `admin`（首次登录会要求修改）

## 3. 配置文件

### Prometheus 配置

创建 `monitoring/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'golang-pay-core'
    environment: 'production'

# 告警规则文件
rule_files:
  - 'alerts.yml'

# Alertmanager 配置
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

# 抓取配置
scrape_configs:
  # 应用指标
  - job_name: 'golang-pay-core'
    scrape_interval: 5s
    metrics_path: '/metrics'
    static_configs:
      - targets: ['host.docker.internal:8080']  # Docker 环境
        # - targets: ['localhost:8080']          # 本地环境
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        replacement: 'golang-pay-core:8080'

  # Prometheus 自身指标
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
```

### 告警规则配置

创建 `monitoring/alerts.yml`:

```yaml
groups:
  - name: golang-pay-core
    interval: 30s
    rules:
      # 高错误率告警
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m])) 
          / 
          sum(rate(http_requests_total[5m])) > 0.01
        for: 5m
        labels:
          severity: critical
          service: golang-pay-core
        annotations:
          summary: "错误率过高"
          description: "服务 {{ $labels.instance }} 的错误率超过 1%，当前值: {{ $value | humanizePercentage }}"

      # 高延迟告警
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95, 
            sum(rate(http_request_duration_seconds_bucket[5m])) by (le)
          ) > 1
        for: 5m
        labels:
          severity: warning
          service: golang-pay-core
        annotations:
          summary: "请求延迟过高"
          description: "服务 {{ $labels.instance }} 的 P95 延迟超过 1 秒，当前值: {{ $value }}s"

      # 服务不可用告警
      - alert: ServiceDown
        expr: up{job="golang-pay-core"} == 0
        for: 1m
        labels:
          severity: critical
          service: golang-pay-core
        annotations:
          summary: "服务不可用"
          description: "服务 {{ $labels.instance }} 已下线超过 1 分钟"

      # 高并发告警
      - alert: HighConcurrency
        expr: http_requests_in_flight > 1000
        for: 5m
        labels:
          severity: warning
          service: golang-pay-core
        annotations:
          summary: "并发请求数过高"
          description: "服务 {{ $labels.instance }} 的并发请求数超过 1000，当前值: {{ $value }}"

      # 数据库连接池耗尽告警
      - alert: DatabaseConnectionPoolExhausted
        expr: |
          (http_requests_total{path="/health"} == 0) 
          OR 
          (up{job="golang-pay-core"} == 1 AND rate(http_requests_total[5m]) == 0)
        for: 2m
        labels:
          severity: warning
          service: golang-pay-core
        annotations:
          summary: "可能数据库连接池耗尽"
          description: "服务可能无法处理请求，检查数据库连接池状态"
```

### Alertmanager 配置

创建 `monitoring/alertmanager.yml`:

```yaml
global:
  resolve_timeout: 5m
  # 邮件配置（可选）
  # smtp_smarthost: 'smtp.example.com:587'
  # smtp_from: 'alerts@example.com'
  # smtp_auth_username: 'alerts@example.com'
  # smtp_auth_password: 'password'

# 路由配置
route:
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'default'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
    - match:
        severity: warning
      receiver: 'warning-alerts'

# 接收器配置
receivers:
  - name: 'default'
    webhook_configs:
      - url: 'http://localhost:5001/webhook'  # 自定义 webhook
        send_resolved: true

  - name: 'critical-alerts'
    # 邮件通知
    # email_configs:
    #   - to: 'ops@example.com'
    #     headers:
    #       Subject: '{{ .GroupLabels.alertname }} 告警'
    
    # Webhook 通知（钉钉、企业微信等）
    webhook_configs:
      - url: 'https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN'
        send_resolved: true

  - name: 'warning-alerts'
    webhook_configs:
      - url: 'http://localhost:5001/webhook'
        send_resolved: true

# 抑制规则
inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'instance']
```

## 4. Grafana 仪表板配置

### 创建数据源

1. 登录 Grafana (http://localhost:3000)
2. 进入 **Configuration** > **Data Sources**
3. 点击 **Add data source**
4. 选择 **Prometheus**
5. 配置 URL: `http://prometheus:9090` (Docker) 或 `http://localhost:9090` (本地)
6. 点击 **Save & Test**

### 导入预定义仪表板

创建 `monitoring/grafana/dashboards/pay-core-dashboard.json`:

```json
{
  "dashboard": {
    "title": "支付系统核心 API 监控",
    "tags": ["pay-core", "golang"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "QPS (每秒请求数)",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total[5m])) by (method, path)",
            "legendFormat": "{{method}} {{path}}"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "请求延迟 (P95)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path))",
            "legendFormat": "P95 {{path}}"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0}
      },
      {
        "id": 3,
        "title": "错误率",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m]))",
            "legendFormat": "错误率"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8}
      },
      {
        "id": 4,
        "title": "当前并发数",
        "type": "graph",
        "targets": [
          {
            "expr": "http_requests_in_flight",
            "legendFormat": "并发数"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8}
      }
    ]
  }
}
```

### Grafana 数据源自动配置

创建 `monitoring/grafana/provisioning/datasources/prometheus.yml`:

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
```

### Grafana 仪表板自动配置

创建 `monitoring/grafana/provisioning/dashboards/default.yml`:

```yaml
apiVersion: 1

providers:
  - name: 'Default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
```

## 5. 常用 Grafana 查询

### QPS 查询

```promql
# 总 QPS
sum(rate(http_requests_total[5m]))

# 按方法分组的 QPS
sum(rate(http_requests_total[5m])) by (method)

# 按路径分组的 QPS
sum(rate(http_requests_total[5m])) by (path)
```

### 延迟查询

```promql
# P50 延迟
histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))

# P95 延迟
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))

# P99 延迟
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
```

### 错误率查询

```promql
# 5xx 错误率
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))

# 4xx 错误率
sum(rate(http_requests_total{status=~"4.."}[5m])) / sum(rate(http_requests_total[5m]))
```

### 并发数查询

```promql
# 当前并发数
http_requests_in_flight

# 平均并发数
avg_over_time(http_requests_in_flight[5m])
```

## 6. 告警通知配置

### 邮件通知

在 `alertmanager.yml` 中配置：

```yaml
receivers:
  - name: 'email-alerts'
    email_configs:
      - to: 'ops@example.com'
        from: 'alerts@example.com'
        smarthost: 'smtp.example.com:587'
        auth_username: 'alerts@example.com'
        auth_password: 'password'
        headers:
          Subject: '{{ .GroupLabels.alertname }} 告警'
        html: |
          <h2>告警详情</h2>
          <p><strong>告警名称:</strong> {{ .GroupLabels.alertname }}</p>
          <p><strong>服务:</strong> {{ .GroupLabels.service }}</p>
          <p><strong>描述:</strong> {{ .Annotations.description }}</p>
```

### 钉钉通知

创建 Webhook 接收器：

```yaml
receivers:
  - name: 'dingtalk-alerts'
    webhook_configs:
      - url: 'https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN'
        send_resolved: true
```

### 企业微信通知

```yaml
receivers:
  - name: 'wechat-alerts'
    wechat_configs:
      - corp_id: 'YOUR_CORP_ID'
        api_secret: 'YOUR_API_SECRET'
        to_user: '@all'
        agent_id: 'YOUR_AGENT_ID'
        message: '{{ .GroupLabels.alertname }}: {{ .Annotations.description }}'
```

## 7. 快速启动脚本

创建 `scripts/start-monitoring.sh`:

```bash
#!/bin/bash

echo "启动监控栈..."

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "错误: Docker 未运行，请先启动 Docker"
    exit 1
fi

# 启动监控服务
docker-compose -f docker-compose.monitoring.yml up -d

echo "监控服务已启动:"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3000 (admin/admin)"
echo "  - Alertmanager: http://localhost:9093"
```

## 8. 访问地址总结

| 服务 | 地址 | 说明 |
|------|------|------|
| **Prometheus UI** | http://localhost:9090 | Prometheus 查询界面 |
| **Grafana** | http://localhost:3000 | 可视化仪表板（推荐） |
| **Alertmanager** | http://localhost:9093 | 告警管理界面 |
| **应用指标** | http://localhost:8080/metrics | 原始指标数据 |
| **健康检查** | http://localhost:8080/health | 服务健康状态 |

## 9. 最佳实践

1. **使用 Grafana 作为主要监控界面** - 更强大的可视化和告警功能
2. **设置合理的告警阈值** - 避免告警疲劳
3. **配置告警抑制规则** - 避免重复告警
4. **定期审查告警规则** - 确保告警的有效性
5. **保存 Grafana 仪表板** - 便于团队共享和版本控制

## 10. 故障排查

### Prometheus 无法抓取指标

1. 检查应用是否运行: `curl http://localhost:8080/metrics`
2. 检查 Prometheus 配置中的 targets 是否正确
3. 查看 Prometheus 日志: `docker logs prometheus`

### Grafana 无法连接 Prometheus

1. 检查数据源配置中的 URL
2. 确保 Prometheus 服务可访问
3. 测试连接: 在 Grafana 数据源配置中点击 "Save & Test"

### 告警未触发

1. 检查告警规则语法是否正确
2. 查看 Prometheus 的 Alerts 页面确认规则状态
3. 检查 Alertmanager 配置和日志

