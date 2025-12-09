#!/bin/bash

# 单次订单创建测试脚本
# 使用方法: ./test_single.sh [base_url] [merchant_id] [channel_id] [merchant_key]

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
BASE_URL=${1:-http://localhost:8888}
MERCHANT_ID=${2:-20001}
CHANNEL_ID=${3:-8008}
MERCHANT_KEY=${4:-your_merchant_key}

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}单次订单创建测试${NC}"
echo -e "${GREEN}========================================${NC}"
echo "BASE_URL: $BASE_URL"
echo "MERCHANT_ID: $MERCHANT_ID"
echo "CHANNEL_ID: $CHANNEL_ID"
echo ""

# 生成订单号
TIMESTAMP=$(date +%s%3N)
RANDOM_NUM=$((RANDOM % 10000))
OUT_ORDER_NO="TEST_${TIMESTAMP}_${RANDOM_NUM}"
AMOUNT=10000  # 100元

echo "订单号: $OUT_ORDER_NO"
echo "金额: ${AMOUNT}分 (${AMOUNT}/100元)"
echo ""

# 构建请求参数（不包含sign）
PARAMS_JSON=$(cat <<EOF
{
  "mchId": $MERCHANT_ID,
  "channelId": $CHANNEL_ID,
  "mchOrderNo": "$OUT_ORDER_NO",
  "amount": $AMOUNT,
  "notifyUrl": "https://example.com/notify",
  "jumpUrl": "https://example.com/jump",
  "extra": "{}",
  "compatible": 0,
  "test": false
}
EOF
)

echo "请求参数:"
echo "$PARAMS_JSON" | python3 -m json.tool 2>/dev/null || echo "$PARAMS_JSON"
echo ""

# 注意：这里需要生成签名，但 bash 脚本生成 MD5 签名比较复杂
# 建议使用 Python 或 Node.js 来生成签名
echo -e "${YELLOW}注意: 此脚本需要手动生成签名或使用 k6 脚本${NC}"
echo ""
echo "使用 k6 测试脚本:"
echo "  k6 run loadtest/k6/test_single_order.js"
echo ""
echo "或使用 curl（需要先手动生成签名）:"
echo "  curl -X POST $BASE_URL/api/v1/orders \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '$PARAMS_JSON'"
