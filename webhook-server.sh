#!/bin/bash

# ./webhook-server.sh start         # 启动服务
# ./webhook-server.sh start-daemon  # 启动守护进程
# ./webhook-server.sh stop          # 停止服务
# ./webhook-server.sh status        # 查看服务状态
# ./webhook-server.sh config        # 配置服务

# 配置 (可通过 webhook-config.sh 覆盖)
PORT=9000                       # webhook服务监听端口
WEBHOOK_SECRET="sdasd21312412asdakhi2123" # webhook密钥，可以为空
LOG_FILE="webhook.log"         # 日志文件
REPO_PATH="$(pwd)"             # 仓库路径，默认当前目录
USE_SOCAT=false                # 是否使用socat代替netcat
API_PATH="/webhook"            # 默认API路径
PID_FILE="webhook.pid"         # PID文件路径

# 颜色
GREEN="\033[32m"
RED="\033[31m"
YELLOW="\033[33m"
GREEN_BG="\033[42;37m"
RED_BG="\033[41;37m"
YELLOW_BG="\033[43;37m"
FONT="\033[0m"

# 提示信息
OK="${GREEN}[OK]${FONT}"
ERROR="${RED}[错误]${FONT}"
WARN="${YELLOW}[警告]${FONT}"

# 加载配置文件（如果存在）
CONFIG_FILE="webhook-config.sh"
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
    echo "已加载配置文件: $CONFIG_FILE"
fi

# 函数: 记录日志
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# 函数: 成功消息
success() {
    echo -e "${OK} ${GREEN_BG}$1${FONT}" | tee -a "$LOG_FILE"
}

# 函数: 错误消息
error() {
    echo -e "${ERROR} ${RED_BG}$1${FONT}" | tee -a "$LOG_FILE"
}

# 函数: 警告消息
warn() {
    echo -e "${WARN} ${YELLOW_BG}$1${FONT}" | tee -a "$LOG_FILE"
}

# 函数: 检查依赖
check_dependency() {
    local cmd="$1"
    if ! command -v "$cmd" &> /dev/null; then
        error "未安装 $cmd，请先安装此依赖"
        return 1
    fi
    return 0
}

# 函数: 创建示例配置文件
create_config_example() {
    if [ ! -f "$CONFIG_FILE" ]; then
        cat > "$CONFIG_FILE" <<EOF
#!/bin/bash

# Webhook服务配置
PORT=9000                      # webhook服务监听端口
WEBHOOK_SECRET=""              # webhook密钥，GitHub/GitLab中配置的Secret
LOG_FILE="webhook.log"         # 日志文件路径
REPO_PATH="$(pwd)"             # 仓库路径，通常是当前目录
USE_SOCAT=false                # 如果netcat有问题，设置为true使用socat
API_PATH="/webhook"            # API基础路径，例如 /webhook, /webhook/restart, /webhook/deploy
PID_FILE="webhook.pid"         # PID文件路径

EOF
        success "已创建配置文件模板: $CONFIG_FILE，请按需修改"
    else
        warn "配置文件已存在，未覆盖"
    fi
}

# 函数: 使用netcat启动服务
start_with_netcat() {
    log "使用netcat启动Webhook服务..."
    
    # 检查nc命令
    if ! check_dependency "nc"; then
        error "请安装netcat: 对于MacOS，请运行: brew install netcat"
        exit 1
    fi
    
    # 创建临时响应脚本
    cat > /tmp/webhook_response.sh <<'EOF'
#!/bin/bash

# 从环境变量获取配置信息
REPO_PATH="${REPO_PATH:-$(pwd)}"
API_PATH="${API_PATH:-/webhook}"
LOG_FILE="${LOG_FILE:-webhook.log}"

# 快速处理输入
read first_line
method=$(echo "$first_line" | awk '{print $1}')
path=$(echo "$first_line" | awk '{print $2}')

# 读取所有头部
while read line; do
    line=$(echo "$line" | tr -d '\r\n')
    if [ -z "$line" ]; then break; fi
    
    # 提取特定头部
    if echo "$line" | grep -qi "^X-GitHub-Event:"; then
        event=$(echo "$line" | cut -d':' -f2- | xargs)
    fi
    if echo "$line" | grep -qi "^X-Hub-Signature-256:"; then
        signature=$(echo "$line" | cut -d':' -f2- | xargs)
    fi
    if echo "$line" | grep -qi "^Content-Length:"; then
        length=$(echo "$line" | cut -d':' -f2- | xargs)
    fi
done

# 处理请求体
body=""
if [ -n "$length" ] && [ "$length" -gt 0 ]; then
    # 最多读取1MB数据
    max_size=$((1024*1024))
    if [ "$length" -gt "$max_size" ]; then
        length=$max_size
    fi
    body=$(dd bs=1 count=$length 2>/dev/null)
fi

# 处理GET请求 - 直接返回状态页面
if [ "$method" = "GET" ]; then
    echo -e "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"status\":\"running\",\"path\":\"$path\"}"
    exit 0
fi

# 处理POST请求
if [ "$method" = "POST" ]; then
    # 输出日志
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 收到Webhook请求，事件类型: ${event:-未知}, 路径: ${path:-/}" >> "$LOG_FILE"
    
    # 验证Secret (如果配置了密钥)
    if [ -n "$WEBHOOK_SECRET" ]; then
        hash="sha256=$(echo -n "$body" | openssl dgst -sha256 -hmac "$WEBHOOK_SECRET" | awk '{print $2}')"
        if [ "$hash" != "$signature" ]; then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [错误] Webhook签名验证失败: $hash" >> "$LOG_FILE"
            echo -e "HTTP/1.1 401 Unauthorized\r\nContent-Type: application/json\r\n\r\n{\"code\":401,\"msg\":\"签名验证失败\"}"
            exit 1
        fi
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] [OK] Webhook签名验证成功" >> "$LOG_FILE"
    fi
    
    # 根据API路径执行不同操作
    cd "$REPO_PATH"
    result=0
    if [ "$path" = "$API_PATH" ] || [ "$path" = "/" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 执行更新..." >> "$LOG_FILE"
        if ./cmd update; then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [OK] 代码更新成功" >> "$LOG_FILE"
        else
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [错误] 代码更新失败" >> "$LOG_FILE"
            result=1
        fi
    elif [ "$path" = "$API_PATH/restart" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 执行重启..." >> "$LOG_FILE"
        if ./cmd restart; then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [OK] 服务重启成功" >> "$LOG_FILE"
        else
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [错误] 服务重启失败" >> "$LOG_FILE"
            result=1
        fi
    elif [ "$path" = "$API_PATH/deploy" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 执行部署..." >> "$LOG_FILE"
        if ./cmd deploy; then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [OK] 服务部署成功" >> "$LOG_FILE"
        else
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [错误] 服务部署失败" >> "$LOG_FILE"
            result=1
        fi
    else
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] [警告] 未知API路径: $path" >> "$LOG_FILE"
        result=1
    fi
    
    # 返回响应
    if [ $result -eq 0 ]; then
        echo -e "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"code\":200,\"msg\":\"任务已触发\",\"path\":\"$path\"}"
    else
        echo -e "HTTP/1.1 404 Not Found\r\nContent-Type: application/json\r\n\r\n{\"code\":404,\"msg\":\"未找到匹配的API路径或执行失败\",\"path\":\"$path\"}"
    fi
    exit 0
fi

# 其他请求类型
echo -e "HTTP/1.1 405 Method Not Allowed\r\nContent-Type: application/json\r\n\r\n{\"code\":405,\"msg\":\"不支持的请求方法\"}"
EOF
    chmod +x /tmp/webhook_response.sh
    
    # 检查端口是否被占用
    if nc -z localhost $PORT 2>/dev/null; then
        error "端口 $PORT 已被占用，请选择其他端口或释放此端口"
        exit 1
    fi
    
    # 导出环境变量，供响应脚本使用
    export REPO_PATH
    export API_PATH
    export WEBHOOK_SECRET
    export LOG_FILE
    
    # 使用连续循环处理请求
    while true; do
        log "等待连接到端口 $PORT..."
        nc -l $PORT | /tmp/webhook_response.sh
        log "连接关闭，等待下一个连接..."
        sleep 1
    done
}

# 函数: 使用socat启动服务
start_with_socat() {
    log "使用socat启动Webhook服务..."
    
    # 检查socat命令
    if ! check_dependency "socat"; then
        error "请安装socat: 对于MacOS，请运行: brew install socat"
        exit 1
    fi
    
    # 检查端口是否被占用
    if nc -z localhost $PORT 2>/dev/null; then
        error "端口 $PORT 已被占用，请选择其他端口或释放此端口"
        exit 1
    fi
    
    # 确保响应脚本存在（与netcat使用同一脚本）
    if [ ! -f "/tmp/webhook_response.sh" ]; then
        error "响应脚本不存在，请先运行n"etcat配置
        exit 1
    fi
    
    # 导出环境变量，供响应脚本使用
    export REPO_PATH
    export API_PATH
    export WEBHOOK_SECRET
    export LOG_FILE
    
    # 使用socat启动服务
    while true; do
        log "使用socat等待连接到端口 $PORT..."
        socat TCP-LISTEN:$PORT,crlf,reuseaddr,fork EXEC:"/tmp/webhook_response.sh" 2>&1 | while read line; do
            log "socat: $line"
        done
        
        # 如果socat退出，重新启动
        log "socat退出，正在重启..."
        sleep 1
    done
}

# 函数: 显示使用帮助
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  start           启动webhook服务器（前台模式）"
    echo "  start-daemon    后台启动webhook服务器（守护进程模式）"
    echo "  stop            停止webhook服务器"
    echo "  status          显示webhook服务器状态"
    echo "  config          创建webhook配置文件模板"
    echo "  -h, --help      显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 start        在前台启动webhook服务器"
    echo "  $0 start-daemon 在后台启动webhook服务器"
    echo "  $0 config       创建配置文件模板"
    echo ""
    echo "支持的API路径:"
    echo "  $API_PATH       默认路径，执行更新操作"
    echo "  $API_PATH/restart  重启服务"
    echo "  $API_PATH/deploy   部署服务"
}

# 函数: 启动webhook服务器
start_webhook_server() {
    local daemon_mode=$1
    
    # 创建日志文件
    touch "$LOG_FILE"
    
    # 显示启动信息
    echo "========================================"
    echo "   Webhook服务器 (端口: $PORT)          "
    echo "========================================"
    echo "仓库路径: $REPO_PATH"
    echo "日志文件: $LOG_FILE"
    echo "API路径: $API_PATH (默认路径)"
    echo "       $API_PATH/restart (重启服务)"
    echo "       $API_PATH/deploy (部署服务)"
    if [ -n "$WEBHOOK_SECRET" ]; then
        echo "Webhook密钥: 已配置"
    else
        echo "Webhook密钥: 未配置 (不进行签名验证)"
    fi
    echo "========================================"
    
    # 检查端口是否被占用
    if nc -z localhost $PORT 2>/dev/null; then
        error "端口 $PORT 已被占用，请选择其他端口或释放此端口"
        exit 1
    fi
    
    # 守护进程模式
    if [ "$daemon_mode" = "daemon" ]; then
        echo "正在后台启动webhook服务器..."
        
        # 确保响应脚本存在
        if [ ! -f "/tmp/webhook_response.sh" ]; then
            error "响应脚本不存在，请先运行前台模式"
            exit 1
        fi
        
        # 导出环境变量
        export REPO_PATH
        export API_PATH
        export WEBHOOK_SECRET
        export LOG_FILE
        export PORT
        export USE_SOCAT
        
        # 启动后台服务，将输出重定向到日志文件
        nohup "$0" "daemon-worker" >> "$LOG_FILE" 2>&1 &
        
        # 保存PID
        PID=$!
        echo $PID > "$PID_FILE"
        
        # 等待1秒检查进程是否存在
        sleep 1
        if ps -p $PID > /dev/null; then
            success "webhook服务已在后台启动 (PID: $PID)"
            echo "日志文件: $LOG_FILE"
        else
            error "webhook服务启动失败，请查看日志文件: $LOG_FILE"
            rm -f "$PID_FILE"
            exit 1
        fi
    else
        # 前台模式
        echo "按 CTRL+C 停止服务"
        echo "========================================"
        
        # 根据配置选择启动方式
        if [ "$USE_SOCAT" = "true" ]; then
            start_with_socat
        else
            start_with_netcat
        fi
    fi
}

# 函数: 在后台工作的进程
daemon_worker() {
    # 记录启动信息
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] webhook服务在后台启动 (PID: $$)" >> "$LOG_FILE"
    
    # 根据配置选择启动方式
    if [ "$USE_SOCAT" = "true" ]; then
        start_with_socat
    else
        start_with_netcat
    fi
}

# 主程序开始
case "$1" in
    start)
        # 前台启动
        start_webhook_server "foreground"
        ;;
        
    start-daemon)
        # 后台启动
        start_webhook_server "daemon"
        ;;
        
    daemon-worker)
        # 守护进程工作模式（内部使用）
        daemon_worker
        ;;
        
    stop)
        # 查找并杀死webhook进程
        if [ -f "$PID_FILE" ]; then
            PID=$(cat "$PID_FILE")
            if ps -p $PID > /dev/null; then
                kill -9 $PID
                success "已停止webhook服务 (PID: $PID)"
                rm -f "$PID_FILE"
            else
                warn "PID文件存在，但进程不存在 (PID: $PID)"
                rm -f "$PID_FILE"
            fi
        else
            # 尝试通过ps查找进程
            PID=$(ps -ef | grep "$0" | grep -v grep | grep -v "stop\|status" | awk '{print $2}')
            if [ -n "$PID" ]; then
                kill -9 $PID
                success "已停止webhook服务 (PID: $PID)"
            else
                warn "没有找到运行中的webhook服务"
            fi
        fi
        ;;
        
    status)
        # 检查webhook服务状态
        if [ -f "$PID_FILE" ]; then
            PID=$(cat "$PID_FILE")
            if ps -p $PID > /dev/null; then
                success "webhook服务正在后台运行 (PID: $PID, 端口: $PORT)"
                echo "支持的API路径:"
                echo "  $API_PATH       默认路径，执行更新操作"
                echo "  $API_PATH/restart  重启服务"
                echo "  $API_PATH/deploy   部署服务"
            else
                warn "PID文件存在，但进程不存在 (PID: $PID)"
                rm -f "$PID_FILE"
            fi
        else
            # 尝试通过ps查找进程
            PID=$(ps -ef | grep "$0" | grep -v grep | grep -v "stop\|status" | awk '{print $2}')
            if [ -n "$PID" ]; then
                success "webhook服务正在运行 (PID: $PID, 端口: $PORT)"
                echo "支持的API路径:"
                echo "  $API_PATH       默认路径，执行更新操作"
                echo "  $API_PATH/restart  重启服务"
                echo "  $API_PATH/deploy   部署服务"
            else
                warn "webhook服务未运行"
            fi
        fi
        ;;
        
    config)
        # 创建配置文件模板
        create_config_example
        ;;
        
    -h|--help|help)
        # 显示帮助信息
        show_help
        ;;
        
    *)
        # 未知参数，显示帮助
        if [ -n "$1" ]; then
            error "未知参数: $1"
        fi
        show_help
        exit 1
        ;;
esac

exit 0 