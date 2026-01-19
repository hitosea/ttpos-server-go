#!/bin/bash
# Bug-251127-005 测试验证脚本

set -e

echo "======================================"
echo "Bug-251127-005 测试验证"
echo "物料详情接口 allow_substore_visible 字段"
echo "======================================"
echo ""

# 配置
BASE_URL="${BASE_URL:-https://ttpos-test1.ttpos.com}"
MATERIAL_UUID="${MATERIAL_UUID:-3699861597323265}"
TOKEN="${TOKEN:-}"

# 检查 TOKEN
if [ -z "$TOKEN" ]; then
    echo "❌ 错误：未设置 TOKEN"
    echo ""
    echo "使用方法："
    echo "  export TOKEN='your-jwt-token'"
    echo "  ./test_bug_251127_005.sh"
    echo ""
    echo "或："
    echo "  TOKEN='your-jwt-token' ./test_bug_251127_005.sh"
    echo ""
    exit 1
fi

echo "测试环境: $BASE_URL"
echo "物料 UUID: $MATERIAL_UUID"
echo ""

# 测试接口
echo "📡 调用接口..."
RESPONSE=$(curl -s -w "\n%{http_code}" \
    -H "Authorization: Bearer $TOKEN" \
    "$BASE_URL/api/v1/shop/material/detail?uuid=$MATERIAL_UUID")

# 分离响应体和状态码
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "HTTP 状态码: $HTTP_CODE"
echo ""

if [ "$HTTP_CODE" != "200" ]; then
    echo "❌ 接口调用失败"
    echo "响应内容:"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    exit 1
fi

# 解析响应
echo "✅ 接口调用成功"
echo ""

# 检查 code 字段
CODE=$(echo "$BODY" | jq -r '.code' 2>/dev/null)
if [ "$CODE" != "0" ]; then
    echo "❌ 业务错误: code = $CODE"
    MESSAGE=$(echo "$BODY" | jq -r '.message' 2>/dev/null)
    echo "错误信息: $MESSAGE"
    exit 1
fi

echo "✅ 业务状态正常 (code = 0)"
echo ""

# 检查 allow_substore_visible 字段
ALLOW_SUBSTORE_VISIBLE=$(echo "$BODY" | jq -r '.data.allow_substore_visible' 2>/dev/null)

echo "🔍 字段验证结果:"
echo "===================="

if [ "$ALLOW_SUBSTORE_VISIBLE" = "null" ] || [ -z "$ALLOW_SUBSTORE_VISIBLE" ]; then
    echo "❌ allow_substore_visible 字段不存在或为 null"
    echo ""
    echo "完整响应:"
    echo "$BODY" | jq '.data' 2>/dev/null || echo "$BODY"
    exit 1
else
    echo "✅ allow_substore_visible 字段存在"
    echo "   值: $ALLOW_SUBSTORE_VISIBLE"
    
    # 验证值的有效性
    if [ "$ALLOW_SUBSTORE_VISIBLE" = "0" ] || [ "$ALLOW_SUBSTORE_VISIBLE" = "1" ]; then
        echo "✅ 字段值有效 (0 或 1)"
    else
        echo "⚠️  字段值异常: $ALLOW_SUBSTORE_VISIBLE (预期 0 或 1)"
    fi
fi

echo ""
echo "===================="
echo "📋 测试结果汇总"
echo "===================="
echo "✅ HTTP 状态码: 200"
echo "✅ 业务状态码: 0"
echo "✅ allow_substore_visible 字段存在: $ALLOW_SUBSTORE_VISIBLE"
echo ""
echo "🎉 Bug-251127-005 修复验证通过！"

