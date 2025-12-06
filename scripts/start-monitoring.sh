#!/bin/bash

echo "=========================================="
echo "启动监控栈..."
echo "=========================================="

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ 错误: Docker 未运行，请先启动 Docker"
    exit 1
fi

# 检查配置文件是否存在
if [ ! -f "monitoring/prometheus.yml" ]; then
    echo "❌ 错误: monitoring/prometheus.yml 文件不存在"
    exit 1
fi

# 启动监控服务
echo "正在启动监控服务..."
docker-compose -f docker-compose.monitoring.yml up -d

# 等待服务启动
echo "等待服务启动..."
sleep 5

# 检查服务状态
echo ""
echo "=========================================="
echo "监控服务状态:"
echo "=========================================="
docker-compose -f docker-compose.monitoring.yml ps

echo ""
echo "=========================================="
echo "✅ 监控服务已启动!"
echo "=========================================="
echo ""
echo "访问地址:"
echo "  📊 Prometheus:    http://localhost:9090"
echo "  📈 Grafana:       http://localhost:3000"
echo "     - 用户名: admin"
echo "     - 密码:   admin (首次登录会要求修改)"
echo "  🔔 Alertmanager: http://localhost:9093"
echo ""
echo "停止服务: docker-compose -f docker-compose.monitoring.yml down"
echo "查看日志: docker-compose -f docker-compose.monitoring.yml logs -f"
echo ""

