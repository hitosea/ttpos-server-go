#!/bin/bash

# DooTask MCP 配置脚本
# 功能：获取 DooTask 登录 token 并通过 claude mcp add 命令配置 MCP 服务器

set -e

# 确保 PATH 包含标准路径
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"

# 加载 .env 文件
if [ -f "$ENV_FILE" ]; then
    # 加载环境变量，忽略注释和空行
    export $(grep -E '^(DOOTASK_EMAIL|DOOTASK_PASSWORD)=' "$ENV_FILE" | xargs)
else
    echo "错误: 未找到 .env 文件: $ENV_FILE"
    echo "请复制 .env.example 并填写配置: cp ${SCRIPT_DIR}/.env.example ${SCRIPT_DIR}/.env"
    exit 1
fi

# 验证必要的环境变量
if [ -z "$DOOTASK_EMAIL" ] || [ -z "$DOOTASK_PASSWORD" ]; then
    echo "错误: DOOTASK_EMAIL 或 DOOTASK_PASSWORD 未在 .env 文件中配置"
    exit 1
fi

# DooTask MCP 配置
DOOTASK_MCP_URL="https://t.hitosea.com/apps/mcp_server/mcp"
LOGIN_URL="https://t.hitosea.com/api/users/login"
LOG_FILE="/tmp/update_mcp_token.log"

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# 检查依赖
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        log "错误: 未找到 curl 命令"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        log "错误: 未找到 jq 命令，请安装: sudo apt-get install jq"
        exit 1
    fi

    if ! command -v claude &> /dev/null; then
        log "错误: 未找到 claude 命令，请先安装 Claude Code CLI"
        exit 1
    fi
}

# 获取新的 token
get_token() {
    SECONDS=$(date +%s)
    MILLIS=$((SECONDS * 1000))
    TIMESTAMP=$MILLIS

    # URL 编码邮箱和密码
    ENCODED_EMAIL=$(printf '%s' "$DOOTASK_EMAIL" | jq -sRr @uri)
    ENCODED_PASSWORD=$(printf '%s' "$DOOTASK_PASSWORD" | jq -sRr @uri)

    LOGIN_URL_FULL="${LOGIN_URL}?type=login&email=${ENCODED_EMAIL}&password=${ENCODED_PASSWORD}&_nocache=${TIMESTAMP}"

    # 调用登录 API
    RESPONSE=$(curl -s "$LOGIN_URL_FULL")

    # 检查响应是否为空
    if [ -z "$RESPONSE" ]; then
        log "错误: API 响应为空"
        return 1
    fi

    # 解析 JSON 获取 token
    TOKEN=$(echo "$RESPONSE" | jq -r '.data.token // empty')

    # 检查 token 是否获取成功
    if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
        log "错误: 未能从响应中提取 token"
        log "响应内容: $RESPONSE"
        return 1
    fi

    echo "$TOKEN"
}

# 使用 claude mcp add 命令添加/更新 DooTask MCP 服务器
add_dootask_mcp() {
    local new_token="$1"

    if [ -z "$new_token" ]; then
        log "错误: token 为空，无法更新"
        return 1
    fi

    log "正在使用 claude mcp add 命令配置 DooTask..."

    # 使用 claude mcp add 命令添加 MCP 服务器
    # --transport http 指定传输类型
    # --header 设置 Authorization header
    if claude mcp add --transport http DooTask "$DOOTASK_MCP_URL" \
        --header "Authorization: Bearer ${new_token}" 2>&1 | tee -a "$LOG_FILE"; then
        log "成功配置 DooTask MCP 服务器"
        return 0
    else
        log "错误: claude mcp add 命令执行失败"
        return 1
    fi
}

# 主函数
main() {
    log "========== 开始更新 MCP Token =========="

    check_dependencies

    NEW_TOKEN=$(get_token)
    if [ $? -ne 0 ] || [ -z "$NEW_TOKEN" ]; then
        log "获取 token 失败，退出"
        exit 1
    fi

    add_dootask_mcp "$NEW_TOKEN"
    if [ $? -eq 0 ]; then
        log "========== Token 更新完成 =========="
        exit 0
    else
        log "========== Token 更新失败 =========="
        exit 1
    fi
}

main "$@"
