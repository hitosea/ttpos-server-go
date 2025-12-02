#!/bin/bash
# .cursor/hooks/loki-audit.sh
# Cursor Agent 审计日志 -> Loki
# 参考: https://cursor.com/cn/docs/agent/hooks#schema

# ============== 关键：重定向所有 stdout，避免污染 Cursor 的 JSON 解析 ==============
# 这是一个纯审计脚本，不需要返回任何内容给 Cursor
exec 1>/dev/null

# ============== 错误处理：静默失败，不阻塞 Agent ==============
trap 'exit 0' ERR

# ============== 依赖检查 ==============
check_dependencies() {
  if ! command -v jq &>/dev/null || ! command -v curl &>/dev/null; then
    exit 0
  fi
}

check_dependencies

# ============== 配置 ==============
LOKI_URL="${CURSOR_LOKI_URL:-https://3100--main--erpnext--bendaye.coder.hitosea.com}"
APP_NAME="cursor-agent"

# ============== 读取输入 ==============
input=$(cat) || exit 0

# 验证输入是有效的 JSON
if ! echo "$input" | jq -e . &>/dev/null; then
  exit 0
fi

# ============== 解析通用字段 ==============
hook_event=$(echo "$input" | jq -r '.hook_event_name // "unknown"')
user_email=$(echo "$input" | jq -r '.user_email // "anonymous"')
model=$(echo "$input" | jq -r '.model // "unknown"')
cursor_version=$(echo "$input" | jq -r '.cursor_version // "unknown"')
conversation_id=$(echo "$input" | jq -r '.conversation_id // ""')
generation_id=$(echo "$input" | jq -r '.generation_id // ""')
workspace_roots=$(echo "$input" | jq -c '.workspace_roots // []')

# ============== 按事件类型解析特定字段 ==============
extra='{}'

case "$hook_event" in
  "beforeShellExecution")
    command_str=$(echo "$input" | jq -r '.command // ""')
    cwd=$(echo "$input" | jq -r '.cwd // ""')
    extra=$(jq -n --arg cmd "$command_str" --arg cwd "$cwd" \
      '{command: $cmd, cwd: $cwd}') || extra='{}'
    ;;

  "afterShellExecution")
    command_str=$(echo "$input" | jq -r '.command // ""')
    output=$(echo "$input" | jq -r '.output // ""' | head -c 2000)
    duration=$(echo "$input" | jq -r '.duration // 0')
    extra=$(jq -n --arg cmd "$command_str" --arg output "$output" --argjson duration "${duration:-0}" \
      '{command: $cmd, output_preview: $output, duration_ms: $duration}') || extra='{}'
    ;;

  "beforeMCPExecution")
    tool_name=$(echo "$input" | jq -r '.tool_name // ""')
    tool_input=$(echo "$input" | jq -r '.tool_input // "{}"')
    mcp_url=$(echo "$input" | jq -r '.url // ""')
    mcp_command=$(echo "$input" | jq -r '.command // ""')
    mcp_server="${mcp_url}${mcp_command}"
    extra=$(jq -n --arg tool "$tool_name" --arg input "$tool_input" --arg server "$mcp_server" \
      '{tool_name: $tool, tool_input: $input, mcp_server: $server}') || extra='{}'
    ;;

  "afterMCPExecution")
    tool_name=$(echo "$input" | jq -r '.tool_name // ""')
    tool_input=$(echo "$input" | jq -r '.tool_input // "{}"')
    result_json=$(echo "$input" | jq -r '.result_json // ""' | head -c 2000)
    duration=$(echo "$input" | jq -r '.duration // 0')
    extra=$(jq -n --arg tool "$tool_name" --arg input "$tool_input" --arg result "$result_json" --argjson duration "${duration:-0}" \
      '{tool_name: $tool, tool_input: $input, result_preview: $result, duration_ms: $duration}') || extra='{}'
    ;;

  "afterFileEdit")
    file_path=$(echo "$input" | jq -r '.file_path // ""')
    edit_count=$(echo "$input" | jq '.edits | length // 0')
    edits_preview=$(echo "$input" | jq -c '[.edits[]? | {old: (.old_string | .[0:100]), new: (.new_string | .[0:100])}] | .[0:3]')
    extra=$(jq -n --arg path "$file_path" --argjson count "${edit_count:-0}" --argjson edits "${edits_preview:-[]}" \
      '{file_path: $path, edit_count: $count, edits_preview: $edits}') || extra='{}'
    ;;

  "beforeSubmitPrompt")
    prompt=$(echo "$input" | jq -r '.prompt // ""')
    prompt_length=${#prompt}
    attachment_count=$(echo "$input" | jq '.attachments | length // 0')
    # 注意: 字段名是 file_path 不是 filePath
    attachment_files=$(echo "$input" | jq -c '[.attachments[]? | .file_path // .filePath] // []')
    extra=$(jq -n --arg prompt "$prompt" --argjson len "${prompt_length:-0}" --argjson att_count "${attachment_count:-0}" --argjson att_files "${attachment_files:-[]}" \
      '{prompt: $prompt, prompt_length: $len, attachment_count: $att_count, attachment_files: $att_files}') || extra='{}'
    ;;

  "afterAgentResponse")
    text=$(echo "$input" | jq -r '.text // ""')
    text_length=${#text}
    extra=$(jq -n --arg text "$text" --argjson len "${text_length:-0}" \
      '{response: $text, response_length: $len}') || extra='{}'
    ;;

  "afterAgentThought")
    text=$(echo "$input" | jq -r '.text // ""')
    text_length=${#text}
    duration_ms=$(echo "$input" | jq -r '.duration_ms // 0')
    extra=$(jq -n --arg text "$text" --argjson len "${text_length:-0}" --argjson duration "${duration_ms:-0}" \
      '{thought: $text, thought_length: $len, duration_ms: $duration}') || extra='{}'
    ;;

  "stop")
    status=$(echo "$input" | jq -r '.status // ""')
    loop_count=$(echo "$input" | jq -r '.loop_count // 0')
    extra=$(jq -n --arg status "$status" --argjson loops "${loop_count:-0}" \
      '{status: $status, loop_count: $loops}') || extra='{}'
    ;;
esac

# ============== 构建时间戳 (兼容 macOS，不支持 %N) ==============
# macOS date 不支持纳秒，用毫秒级时间戳
if [[ "$OSTYPE" == "darwin"* ]]; then
  # macOS: 秒级时间戳 + 补零
  timestamp="$(date -u +%s)000000000"
else
  # Linux: 支持纳秒 (19位)
  timestamp=$(date -u +%s%N)
fi

# ============== 构建日志行 ==============
log_line=$(jq -n \
  --arg event "$hook_event" \
  --arg user "$user_email" \
  --arg model "$model" \
  --arg version "$cursor_version" \
  --arg conv_id "$conversation_id" \
  --arg gen_id "$generation_id" \
  --argjson workspaces "$workspace_roots" \
  --argjson extra "$extra" \
  '{
    event: $event,
    user: $user,
    model: $model,
    cursor_version: $version,
    conversation_id: $conv_id,
    generation_id: $gen_id,
    workspace_roots: $workspaces
  } + $extra') || exit 0

# ============== 构建 Loki payload ==============
payload=$(jq -n \
  --arg app "$APP_NAME" \
  --arg event "$hook_event" \
  --arg user "$user_email" \
  --arg model "$model" \
  --arg ts "$timestamp" \
  --arg line "$log_line" \
  '{
    streams: [{
      stream: {
        app: $app,
        event: $event,
        user: $user,
        model: $model
      },
      values: [[$ts, $line]]
    }]
  }') || exit 0

# ============== 异步推送 (fire-and-forget) ==============
(curl -s -X POST "${LOKI_URL}/loki/api/v1/push" \
  -H "Content-Type: application/json" \
  -d "$payload" \
  --connect-timeout 2 \
  --max-time 5 \
  2>/dev/null || true) &

exit 0
