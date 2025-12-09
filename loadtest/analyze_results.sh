#!/bin/bash

# 压测结果分析脚本
# 使用方法: ./analyze_results.sh [result_file]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_DIR="$SCRIPT_DIR/results"

# 检查jq是否安装（用于JSON解析）
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}警告: jq 未安装，部分功能可能不可用${NC}"
    echo "安装方法:"
    echo "  macOS: brew install jq"
    echo "  Linux: sudo apt-get install jq"
fi

# 分析k6 JSON结果
analyze_k6_json() {
    local file=$1
    echo -e "${BLUE}=== k6 压测结果分析 ===${NC}"
    
    if command -v jq &> /dev/null; then
        echo ""
        echo -e "${GREEN}HTTP请求指标:${NC}"
        jq -r '.metrics.http_reqs | "总请求数: \(.values.count)\nQPS: \(.values.rate | tostring | .[0:5])\n"' "$file" 2>/dev/null || echo "无法解析"
        
        echo -e "${GREEN}响应时间指标:${NC}"
        jq -r '.metrics.http_req_duration | "平均响应时间: \(.values.avg)ms\n最小响应时间: \(.values.min)ms\n最大响应时间: \(.values.max)ms\nP95响应时间: \(.values["p(95)"])ms\nP99响应时间: \(.values["p(99)"])ms\n"' "$file" 2>/dev/null || echo "无法解析"
        
        echo -e "${GREEN}错误率:${NC}"
        jq -r '.metrics.http_req_failed | "错误率: \(.values.rate * 100)%\n失败请求数: \(.values.fails)\n"' "$file" 2>/dev/null || echo "无法解析"
        
        echo -e "${GREEN}阈值检查:${NC}"
        jq -r '.root_group.checks[]? | "\(.name): \(.passes)/\(.passes + .fails) 通过"' "$file" 2>/dev/null || echo "无法解析"
    else
        echo "请安装jq以查看详细分析"
    fi
}

# 分析wrk文本结果
analyze_wrk_text() {
    local file=$1
    echo -e "${BLUE}=== wrk 压测结果分析 ===${NC}"
    echo ""
    cat "$file"
}

# 主函数
main() {
    if [ $# -eq 0 ]; then
        # 如果没有指定文件，列出所有结果文件
        echo -e "${GREEN}可用的压测结果文件:${NC}"
        echo ""
        
        local k6_files=$(find "$RESULTS_DIR" -name "*.json" -type f 2>/dev/null | sort -r)
        local wrk_files=$(find "$RESULTS_DIR" -name "*_wrk_*.txt" -type f 2>/dev/null | sort -r)
        
        if [ -n "$k6_files" ] || [ -n "$wrk_files" ]; then
            if [ -n "$k6_files" ]; then
                echo -e "${BLUE}k6 结果文件:${NC}"
                echo "$k6_files" | head -5 | nl
                echo ""
            fi
            
            if [ -n "$wrk_files" ]; then
                echo -e "${BLUE}wrk 结果文件:${NC}"
                echo "$wrk_files" | head -5 | nl
                echo ""
            fi
            
            echo "使用方法: $0 <result_file>"
        else
            echo -e "${YELLOW}未找到结果文件${NC}"
        fi
        exit 0
    fi
    
    local file=$1
    
    if [ ! -f "$file" ]; then
        echo -e "${RED}错误: 文件 $file 不存在${NC}"
        exit 1
    fi
    
    # 根据文件类型选择分析方式
    if [[ "$file" == *.json ]]; then
        analyze_k6_json "$file"
    elif [[ "$file" == *.txt ]]; then
        analyze_wrk_text "$file"
    else
        echo -e "${RED}错误: 不支持的文件格式${NC}"
        exit 1
    fi
}

main "$@"
