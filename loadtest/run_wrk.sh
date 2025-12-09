#!/bin/bash

# wrk 压测运行脚本
# 使用方法: ./run_wrk.sh [scenario] [base_url] [threads] [connections] [duration]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WRK_DIR="$SCRIPT_DIR/wrk"
RESULTS_DIR="$SCRIPT_DIR/results"

# 创建结果目录
mkdir -p "$RESULTS_DIR"

# 检查wrk是否安装
if ! command -v wrk &> /dev/null; then
    echo -e "${RED}错误: wrk 未安装${NC}"
    echo "安装方法:"
    echo "  macOS: brew install wrk"
    echo "  Linux: sudo apt-get install wrk 或从源码编译"
    exit 1
fi

# 加载配置
CONFIG_FILE="$SCRIPT_DIR/config.env"
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
fi

# 参数处理
SCENARIO=${1:-health_check}
BASE_URL=${2:-${BASE_URL:-http://localhost:8080}}
THREADS=${3:-4}
CONNECTIONS=${4:-100}
DURATION=${5:-30s}

# 验证场景文件是否存在
SCENARIO_FILE="$WRK_DIR/${SCENARIO}.lua"
if [ ! -f "$SCENARIO_FILE" ]; then
    echo -e "${RED}错误: 场景文件 $SCENARIO_FILE 不存在${NC}"
    echo "可用场景:"
    ls -1 "$WRK_DIR"/*.lua | xargs -n1 basename | sed 's/.lua$//'
    exit 1
fi

echo -e "${GREEN}开始压测...${NC}"
echo "场景: $SCENARIO"
echo "目标URL: $BASE_URL"
echo "线程数: $THREADS"
echo "连接数: $CONNECTIONS"
echo "持续时间: $DURATION"
echo "结果目录: $RESULTS_DIR"
echo ""

# 设置环境变量
export MERCHANT_ID=${MERCHANT_ID:-1}
export CHANNEL_ID=${CHANNEL_ID:-1}
export MERCHANT_KEY=${MERCHANT_KEY:-your_merchant_key}

# 运行wrk
RESULT_FILE="$RESULTS_DIR/${SCENARIO}_wrk_$(date +%Y%m%d_%H%M%S).txt"
wrk -t"$THREADS" -c"$CONNECTIONS" -d"$DURATION" -s"$SCENARIO_FILE" "$BASE_URL" | tee "$RESULT_FILE"

echo ""
echo -e "${GREEN}压测完成！${NC}"
echo "结果文件保存在: $RESULT_FILE"
