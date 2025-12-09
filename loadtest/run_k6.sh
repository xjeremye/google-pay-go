#!/bin/bash

# k6 压测运行脚本
# 使用方法: ./run_k6.sh [scenario] [base_url]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K6_DIR="$SCRIPT_DIR/k6"
RESULTS_DIR="$SCRIPT_DIR/results"

# 创建结果目录
mkdir -p "$RESULTS_DIR"

# 检查k6是否安装
if ! command -v k6 &> /dev/null; then
    echo -e "${RED}错误: k6 未安装${NC}"
    echo "安装方法:"
    echo "  macOS: brew install k6"
    echo "  Linux: https://k6.io/docs/getting-started/installation/"
    exit 1
fi

# 加载配置
CONFIG_FILE="$SCRIPT_DIR/config.env"
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo -e "${YELLOW}警告: 配置文件 $CONFIG_FILE 不存在，使用默认配置${NC}"
fi

# 参数处理
SCENARIO=${1:-${SCENARIO:-create_order}}
BASE_URL=${2:-${BASE_URL:-http://localhost:8080}}

# 验证场景文件是否存在
SCENARIO_FILE="$K6_DIR/${SCENARIO}.js"
if [ ! -f "$SCENARIO_FILE" ]; then
    echo -e "${RED}错误: 场景文件 $SCENARIO_FILE 不存在${NC}"
    echo "可用场景:"
    ls -1 "$K6_DIR"/*.js | xargs -n1 basename | sed 's/.js$//'
    exit 1
fi

echo -e "${GREEN}开始压测...${NC}"
echo "场景: $SCENARIO"
echo "目标URL: $BASE_URL"
echo "结果目录: $RESULTS_DIR"
echo ""

# 设置环境变量
export BASE_URL
export MERCHANT_ID=${MERCHANT_ID:-1}
export CHANNEL_ID=${CHANNEL_ID:-1}
export MERCHANT_KEY=${MERCHANT_KEY:-your_merchant_key}

# 运行k6
k6 run \
    --out json="$RESULTS_DIR/${SCENARIO}_$(date +%Y%m%d_%H%M%S).json" \
    "$SCENARIO_FILE"

echo ""
echo -e "${GREEN}压测完成！${NC}"
echo "结果文件保存在: $RESULTS_DIR"
