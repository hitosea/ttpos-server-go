#!/bin/bash

# MCP Token 自动更新脚本
# 功能：定时获取 DooTask 登录 token 并更新 mcp.json 配置文件

set -e

# 确保 PATH 包含标准路径
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"

# 加载 .env 文件
if [ -f "$ENV_FILE" ]; then
    # 加载环境变量，忽略注释和空行
    export $(grep -E '^(DOOTASK_EMAIL|DOOTASK_PASSWORD|MCP_JSON_PATHS)=' "$ENV_FILE" | xargs)
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

# MCP_JSON_PATHS 支持多个文件，用逗号分隔
MCP_JSON_PATHS="${MCP_JSON_PATHS:-/workspace/project/ttpos-server-go/.mcp.private.json}"
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

# 创建 MCP 配置文件（包含 DooTask 配置）
create_mcp_json() {
    local mcp_path="$1"
    local new_token="$2"

    cat > "$mcp_path" << EOF
{
  "mcpServers": {
    "DooTask": {
      "type": "http",
      "url": "https://t.hitosea.com/apps/mcp_server/mcp",
      "headers": {
        "Authorization": "Bearer ${new_token}"
      }
    }
  }
}
EOF
    log "已创建: $mcp_path"
}

# 更新单个 mcp.json 文件
update_single_mcp_json() {
    local mcp_path="$1"
    local new_token="$2"

    # 检查文件是否存在，不存在则创建
    if [ ! -f "$mcp_path" ]; then
        log "文件不存在，正在创建: $mcp_path"
        create_mcp_json "$mcp_path" "$new_token"
        return 0
    fi

    # 检查文件中是否有 DooTask 配置，没有则添加
    if ! jq -e '.mcpServers.DooTask' "$mcp_path" > /dev/null 2>&1; then
        log "添加 DooTask 配置到: $mcp_path"
        if jq --arg token "$new_token" \
           '.mcpServers.DooTask = {"type": "http", "url": "https://t.hitosea.com/apps/mcp_server/mcp", "headers": {"Authorization": "Bearer \($token)"}}' \
           "$mcp_path" > "${mcp_path}.tmp" && \
        mv "${mcp_path}.tmp" "$mcp_path"; then
            log "成功添加 DooTask 配置: $mcp_path"
            return 0
        else
            log "错误: 添加配置失败: $mcp_path"
            return 1
        fi
    fi

    # 使用 jq 更新 Authorization header
    if jq --arg token "$new_token" \
       '.mcpServers.DooTask.headers.Authorization = "Bearer \($token)"' \
       "$mcp_path" > "${mcp_path}.tmp" && \
    mv "${mcp_path}.tmp" "$mcp_path"; then
        log "成功更新: $mcp_path"
        return 0
    else
        log "错误: 更新失败: $mcp_path"
        return 1
    fi
}

# 更新所有 mcp.json 文件
update_all_mcp_json() {
    local new_token="$1"
    local has_error=0

    if [ -z "$new_token" ]; then
        log "错误: token 为空，无法更新"
        return 1
    fi

    # 按逗号分隔遍历所有路径
    IFS=',' read -ra MCP_PATHS <<< "$MCP_JSON_PATHS"
    for mcp_path in "${MCP_PATHS[@]}"; do
        mcp_path=$(echo "$mcp_path" | xargs)
        if [ -n "$mcp_path" ]; then
            update_single_mcp_json "$mcp_path" "$new_token" || has_error=1
        fi
    done

    return $has_error
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

    update_all_mcp_json "$NEW_TOKEN"
    if [ $? -eq 0 ]; then
        log "========== Token 更新完成 =========="
        exit 0
    else
        log "========== Token 更新失败 =========="
        exit 1
    fi
}

main "$@"
