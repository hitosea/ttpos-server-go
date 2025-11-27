#!/bin/bash
# 健康检查集成测试脚本

set -e

echo "========================================="
echo "健康检查集成测试"
echo "========================================="

# 服务地址
BASE_URL="http://localhost:14051"
HEALTH_URL="${BASE_URL}/health"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试计数
TOTAL=0
PASSED=0
FAILED=0

# 测试函数
test_case() {
    TOTAL=$((TOTAL + 1))
    local name="$1"
    local url="$2"
    local expected_code="$3"
    local expected_status="$4"
    
    echo -n "测试 ${TOTAL}: ${name} ... "
    
    # 发送请求 (使用临时文件存储响应体，避免混入状态码)
    local temp_file=$(mktemp)
    http_code=$(curl -s -w "%{http_code}" -o "$temp_file" "${url}")
    body=$(cat "$temp_file")
    rm -f "$temp_file"
    
    # 检查 HTTP 状态码
    if [ "$http_code" != "$expected_code" ]; then
        echo -e "${RED}失败${NC}"
        echo "  期望 HTTP 状态码: $expected_code"
        echo "  实际 HTTP 状态码: $http_code"
        FAILED=$((FAILED + 1))
        return 1
    fi
    
    # 检查响应中的 status 字段（只取第一个匹配，即顶层 status）
    if [ -n "$expected_status" ]; then
        actual_status=$(echo "$body" | grep -o '"status"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"\([^"]*\)".*/\1/')
        if [ "$actual_status" != "$expected_status" ]; then
            echo -e "${RED}失败${NC}"
            echo "  期望 status: $expected_status"
            echo "  实际 status: $actual_status"
            echo "  响应体: ${body:0:200}"
            FAILED=$((FAILED + 1))
            return 1
        fi
    fi
    
    echo -e "${GREEN}通过${NC}"
    echo "  响应: ${body:0:100}..."
    PASSED=$((PASSED + 1))
    return 0
}

echo ""
echo "前置检查："
echo "- 确保 ttpos-websocket 服务已启动"
echo "- 确保 MySQL 和 Redis 正常运行"
echo ""

# 测试 1: 基本健康检查
test_case "基本健康检查" \
    "${HEALTH_URL}" \
    "200" \
    "UP"

# 测试 2: 不带详细信息
test_case "不带详细信息" \
    "${HEALTH_URL}?detail=false" \
    "200" \
    "UP"

# 测试 3: 指定 scope
test_case "指定 scope=liveness" \
    "${HEALTH_URL}?scope=liveness" \
    "200" \
    "UP"

# 测试 4: 检查响应格式（包含必需字段）
echo -n "测试 ${TOTAL}: 检查响应格式 ... "
TOTAL=$((TOTAL + 1))
response=$(curl -s "${HEALTH_URL}")
if echo "$response" | grep -q '"status"' && \
   echo "$response" | grep -q '"timestamp"' && \
   echo "$response" | grep -q '"components"'; then
    echo -e "${GREEN}通过${NC}"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}失败${NC}"
    echo "  响应缺少必需字段"
    FAILED=$((FAILED + 1))
fi

# 测试 5: 检查 MySQL 组件
echo -n "测试 ${TOTAL}: 检查 MySQL 组件 ... "
TOTAL=$((TOTAL + 1))
response=$(curl -s "${HEALTH_URL}")
if echo "$response" | grep -q '"mysql"'; then
    echo -e "${GREEN}通过${NC}"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}失败${NC}"
    echo "  响应中缺少 mysql 组件"
    FAILED=$((FAILED + 1))
fi

# 测试 6: 检查 Redis 组件
echo -n "测试 ${TOTAL}: 检查 Redis 组件 ... "
TOTAL=$((TOTAL + 1))
response=$(curl -s "${HEALTH_URL}")
if echo "$response" | grep -q '"redis"'; then
    echo -e "${GREEN}通过${NC}"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}失败${NC}"
    echo "  响应中缺少 redis 组件"
    FAILED=$((FAILED + 1))
fi

# 测试 7: 性能测试（响应时间）
echo -n "测试 ${TOTAL}: 响应时间检查 ... "
TOTAL=$((TOTAL + 1))
start_time=$(date +%s%N)
curl -s "${HEALTH_URL}" > /dev/null
end_time=$(date +%s%N)
duration=$(( (end_time - start_time) / 1000000 ))  # 转换为毫秒

if [ $duration -lt 1000 ]; then
    echo -e "${GREEN}通过${NC} (${duration}ms)"
    PASSED=$((PASSED + 1))
else
    echo -e "${YELLOW}警告${NC} (${duration}ms)"
    echo "  响应时间超过 1 秒"
    PASSED=$((PASSED + 1))
fi

# 输出测试结果
echo ""
echo "========================================="
echo "测试结果汇总"
echo "========================================="
echo "总计: ${TOTAL}"
echo -e "${GREEN}通过: ${PASSED}${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}失败: ${FAILED}${NC}"
fi
echo "========================================="

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ 所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}❌ 部分测试失败${NC}"
    exit 1
fi

