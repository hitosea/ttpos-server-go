#!/usr/bin/env bash
# dev-server.sh - Worktree 开发服务管理器
# 管理 worktree 服务启停、HTTP 代理配置、主服务重启
set -euo pipefail

# ─── 常量 ───────────────────────────────────────
MAIN_REPO="${HOME}/ttpos-server-go"
PROXY_CONTAINER="http-proxy-debug-view"
PROXY_IMAGE="602666178/http-proxy-debug-view:latest"
DEFAULT_WT_PORT=8082
MAIN_PORT=8080
PROXY_HOST_PORT=8081
WEB_UI_PORT=8091

# ─── 工具函数 ───────────────────────────────────
get_local_ip() {
  bash "${MAIN_REPO}/ttpos-scripts/get_ip.sh" 2>/dev/null || echo "127.0.0.1"
}

get_nginx_port() {
  local dir="${1:-$MAIN_REPO}"
  grep '^NGINX_PORT=' "${dir}/.env" 2>/dev/null | cut -d'=' -f2 | tr -d ' ' || echo "8888"
}

get_pids_on_port() {
  lsof -ti:"${1}" 2>/dev/null || true
}

kill_port() {
  local port=$1
  local pids
  pids=$(get_pids_on_port "${port}")
  if [[ -n "$pids" ]]; then
    echo "🔪 Kill port ${port} (PIDs: ${pids})"
    echo "$pids" | xargs kill -9 2>/dev/null || true
    sleep 1
  fi
}

wait_for_port() {
  local port=$1
  local max_wait=${2:-15}
  local waited=0
  while [[ $waited -lt $max_wait ]]; do
    if lsof -ti:"${port}" > /dev/null 2>&1; then
      return 0
    fi
    sleep 1
    waited=$((waited + 1))
  done
  return 1
}

# ─── serve: 启动 worktree 服务 ──────────────────
cmd_serve() {
  local port="${1:-$DEFAULT_WT_PORT}"
  local wt_dir
  wt_dir="$(pwd)"

  if [[ ! -f "${wt_dir}/.env" ]]; then
    echo "❌ No .env in ${wt_dir}"
    echo "   Run: cp .env.example .env"
    exit 1
  fi

  # 停止已有 worktree 服务
  if [[ -f "${wt_dir}/.dev-server.port" ]]; then
    local old_port
    old_port=$(cat "${wt_dir}/.dev-server.port")
    kill_port "${old_port}"
  fi
  kill_port "${port}"

  echo "🔨 Building..."
  cd "${wt_dir}/main"
  go build -o "${wt_dir}/.dev-server-bin" .

  echo "🚀 Starting on port ${port}..."
  # 从 main/ 目录启动，使 ../.env 正确解析到 worktree 根目录
  # 通过环境变量覆盖端口（godotenv.Load 不覆盖已存在的环境变量）
  SERVER_PORT="${port}" nohup "${wt_dir}/.dev-server-bin" > "${wt_dir}/.dev-server.log" 2>&1 &
  local pid=$!

  echo "${port}" > "${wt_dir}/.dev-server.port"
  echo "${pid}" > "${wt_dir}/.dev-server.pid"

  if wait_for_port "${port}" 15; then
    echo "✅ Worktree service running"
    echo "   Port: ${port} | PID: ${pid}"
    echo "   Log:  ${wt_dir}/.dev-server.log"
  else
    echo "⚠️  Service may still be starting..."
    echo "   Check: tail -f ${wt_dir}/.dev-server.log"
  fi
}

# ─── proxy: 将 worktree 设为代理首选上游 ────────
cmd_proxy() {
  local wt_dir
  wt_dir="$(pwd)"
  local port="${1:-}"

  # 从状态文件读取端口
  if [[ -z "$port" && -f "${wt_dir}/.dev-server.port" ]]; then
    port=$(cat "${wt_dir}/.dev-server.port")
  fi

  if [[ -z "$port" ]]; then
    echo "❌ No worktree port. Run 'serve' first or specify port."
    exit 1
  fi

  local local_ip
  local_ip=$(get_local_ip)
  local nginx_port
  nginx_port=$(get_nginx_port "${MAIN_REPO}")
  # worktree 作为第一优先级上游
  local target_url="http://${local_ip}:${port},http://${local_ip}:${MAIN_PORT},http://${local_ip}:${nginx_port}"

  echo "🔄 Updating proxy → worktree:${port} (primary)..."
  docker rm -f "${PROXY_CONTAINER}" > /dev/null 2>&1 || true
  docker run -d \
    --name "${PROXY_CONTAINER}" \
    -p "${PROXY_HOST_PORT}:8080" \
    -p "${WEB_UI_PORT}:8091" \
    -e TARGET_URL="${target_url}" \
    -e PROXY_PORT=8080 \
    -e WEB_PORT=8091 \
    --restart unless-stopped \
    "${PROXY_IMAGE}" > /dev/null 2>&1

  echo "✅ Proxy updated"
  echo "   localhost:${PROXY_HOST_PORT} → wt:${port} → main:${MAIN_PORT} → nginx:${nginx_port}"
  echo "   Debug UI: http://localhost:${WEB_UI_PORT}"
}

# ─── restore: 恢复代理为仅代理主服务 ────────────
cmd_restore() {
  local local_ip
  local_ip=$(get_local_ip)
  local nginx_port
  nginx_port=$(get_nginx_port "${MAIN_REPO}")
  local target_url="http://${local_ip}:${MAIN_PORT},http://${local_ip}:${nginx_port}"

  echo "🔄 Restoring proxy → main:${MAIN_PORT}..."
  docker rm -f "${PROXY_CONTAINER}" > /dev/null 2>&1 || true
  docker run -d \
    --name "${PROXY_CONTAINER}" \
    -p "${PROXY_HOST_PORT}:8080" \
    -p "${WEB_UI_PORT}:8091" \
    -e TARGET_URL="${target_url}" \
    -e PROXY_PORT=8080 \
    -e WEB_PORT=8091 \
    --restart unless-stopped \
    "${PROXY_IMAGE}" > /dev/null 2>&1

  echo "✅ Proxy restored: localhost:${PROXY_HOST_PORT} → main:${MAIN_PORT}"
}

# ─── stop: 停止 worktree 服务 ───────────────────
cmd_stop() {
  local wt_dir
  wt_dir="$(pwd)"

  if [[ -f "${wt_dir}/.dev-server.port" ]]; then
    local port
    port=$(cat "${wt_dir}/.dev-server.port")
    kill_port "${port}"
  fi

  rm -f "${wt_dir}/.dev-server.port" "${wt_dir}/.dev-server.pid" "${wt_dir}/.dev-server-bin"
  echo "✅ Worktree service stopped"
}

# ─── cleanup: 停止服务 + 恢复代理 ───────────────
cmd_cleanup() {
  cmd_stop
  cmd_restore
}

# ─── main: 重启主仓库 8080 服务 ─────────────────
cmd_main() {
  echo "🔄 Restarting main service on port ${MAIN_PORT}..."

  kill_port "${MAIN_PORT}"

  if [[ ! -f "${MAIN_REPO}/.env" ]]; then
    echo "❌ No .env in ${MAIN_REPO}"
    exit 1
  fi

  echo "🔨 Building main service..."
  cd "${MAIN_REPO}/main"
  go build -o "${MAIN_REPO}/.main-server-bin" .

  echo "🚀 Starting main service..."
  # 从 main/ 目录启动
  nohup "${MAIN_REPO}/.main-server-bin" > "${MAIN_REPO}/.main-server.log" 2>&1 &
  local pid=$!

  if wait_for_port "${MAIN_PORT}" 15; then
    echo "✅ Main service running"
    echo "   Port: ${MAIN_PORT} | PID: ${pid}"
    echo "   Log:  ${MAIN_REPO}/.main-server.log"
  else
    echo "⚠️  Service may still be starting..."
    echo "   Check: tail -f ${MAIN_REPO}/.main-server.log"
  fi
}

# ─── status: 显示所有服务状态 ───────────────────
cmd_status() {
  local wt_dir
  wt_dir="$(pwd)"

  echo "📊 Dev Server Status"
  echo "────────────────────"

  # 主服务
  local main_pids
  main_pids=$(get_pids_on_port "${MAIN_PORT}")
  if [[ -n "$main_pids" ]]; then
    echo "✅ Main     (${MAIN_PORT}): running (PIDs: ${main_pids})"
  else
    echo "❌ Main     (${MAIN_PORT}): not running"
  fi

  # Worktree 服务
  if [[ -f "${wt_dir}/.dev-server.port" ]]; then
    local wt_port
    wt_port=$(cat "${wt_dir}/.dev-server.port")
    local wt_pids
    wt_pids=$(get_pids_on_port "${wt_port}")
    if [[ -n "$wt_pids" ]]; then
      echo "✅ Worktree (${wt_port}): running (PIDs: ${wt_pids})"
    else
      echo "⚠️  Worktree (${wt_port}): configured but not running"
    fi
  else
    echo "⚪ Worktree: not configured"
  fi

  # 代理
  echo ""
  local proxy_status
  proxy_status=$(docker ps --filter "name=${PROXY_CONTAINER}" --format "{{.Status}}" 2>/dev/null || true)
  if [[ -n "$proxy_status" ]]; then
    local targets
    targets=$(docker inspect "${PROXY_CONTAINER}" --format '{{range .Config.Env}}{{println .}}{{end}}' 2>/dev/null | grep TARGET_URL | cut -d= -f2-)
    echo "✅ Proxy    (${PROXY_HOST_PORT}): ${proxy_status}"
    echo "   Targets: ${targets}"
  else
    echo "❌ Proxy: not running"
  fi
}

# ─── 入口 ──────────────────────────────────────
case "${1:-help}" in
  serve)   shift; cmd_serve "$@" ;;
  proxy)   shift; cmd_proxy "$@" ;;
  restore) cmd_restore ;;
  stop)    cmd_stop ;;
  cleanup) cmd_cleanup ;;
  main)    cmd_main ;;
  status)  cmd_status ;;
  *)
    cat <<'HELP'
dev-server - Worktree 开发服务管理器

Usage: dev-server <command> [args]

Commands:
  serve [port]  启动 worktree 服务 (默认端口: 8082)
  proxy [port]  将 worktree 设为代理首选上游
  restore       恢复代理为仅代理主服务 (8080)
  stop          停止 worktree 服务
  cleanup       停止服务 + 恢复代理
  main          重启主仓库 8080 服务
  status        查看所有服务状态

Workflow:
  1. dev-server serve 8082   # 启动 worktree 服务测试
  2. dev-server proxy        # 测试通过后，代理到 worktree
  3. dev-server cleanup      # 开发完成，恢复原状
HELP
    ;;
esac
