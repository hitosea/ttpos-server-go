#!/bin/bash

# Redis集群管理脚本
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m's
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
COMPOSE_FILE="docker-compose.dev.yml -f docker-compose.dev.redis.yml"
APP_ID=${APP_ID:-dev}

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Docker服务
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker服务未运行，请启动Docker"
        exit 1
    fi
}

# 启动Redis集群
start_cluster() {
    log_info "启动Redis集群（自动初始化模式）..."
    
    # 创建数据目录
    mkdir -p docker/redis/cluster/data-{1,2,3}
    
    # 启动所有Redis服务（包括自动初始化服务）
    docker-compose -f $COMPOSE_FILE up -d redis-node-1 redis-node-2 redis-node-3 redis-cluster-init
    
    log_info "Redis集群正在启动和自动初始化中..."
    log_info "请稍等片刻，自动初始化服务会处理集群配置"
    log_info "使用 '$0 status' 检查集群状态"
    log_info "使用 'docker-compose -f $COMPOSE_FILE logs -f redis-cluster-init' 查看初始化日志"
    
    log_success "Redis集群服务已启动"
}

# 停止Redis集群
stop_cluster() {
    log_info "停止Redis集群..."
    
    docker-compose -f $COMPOSE_FILE stop redis-cluster-init redis-node-1 redis-node-2 redis-node-3
    
    log_success "Redis集群已停止"
}

# 重启Redis集群
restart_cluster() {
    log_info "重启Redis集群..."
    stop_cluster
    sleep 5
    start_cluster
}

# 检查集群状态
status_cluster() {
    log_info "检查Redis集群状态..."
    
    echo ""
    echo "=== Docker容器状态 ==="
    docker-compose -f $COMPOSE_FILE ps redis-node-1 redis-node-2 redis-node-3 redis-cluster-init
    
    echo ""
    echo "=== 集群节点信息 ==="
    docker exec -i saas-redis-node-1-${APP_ID} redis-cli -p 7001 cluster nodes 2>/dev/null || log_warning "无法获取集群信息，集群可能正在初始化中"
    
    echo ""
    echo "=== 集群状态 ==="
    docker exec -i saas-redis-node-1-${APP_ID} redis-cli -p 7001 cluster info 2>/dev/null || log_warning "无法获取集群状态"
    
    echo ""
    echo "=== 自动初始化服务状态 ==="
    if docker ps --filter "name=saas-redis-cluster-init-${APP_ID}" --format "table {{.Names}}\t{{.Status}}" | grep -q "Up"; then
        log_success "自动初始化服务正在运行"
    else
        log_warning "自动初始化服务未运行"
    fi
}

# 重新初始化集群
init_cluster() {
    log_info "触发集群重新初始化..."
    
    # 重启自动初始化服务来触发重新初始化
    log_info "重启自动初始化服务..."
    docker-compose -f $COMPOSE_FILE restart redis-cluster-init
    
    log_info "自动初始化服务已重启，集群将自动检查和初始化"
    log_info "使用 'docker-compose -f $COMPOSE_FILE logs -f redis-cluster-init' 查看进度"
    
    log_success "重新初始化已触发"
}

# 连接到集群
connect_cluster() {
    local node=${1:-1}
    local port
    
    case $node in
        1) port=7001 ;;
        2) port=7002 ;;
        3) port=7003 ;;
        *) log_error "无效的节点编号，请使用 1、2 或 3"; exit 1 ;;
    esac
    
    log_info "连接到Redis集群节点 $node (端口: $port)..."
    docker exec -it saas-redis-node-${node}-${APP_ID} redis-cli -p $port -c
}

# 查看集群日志
logs_cluster() {
    local node=${1:-all}
    
    if [[ "$node" == "all" ]]; then
        log_info "查看所有Redis服务日志..."
        docker-compose -f $COMPOSE_FILE logs -f redis-node-1 redis-node-2 redis-node-3 redis-cluster-init
    elif [[ "$node" == "init" ]]; then
        log_info "查看集群自动初始化日志..."
        docker-compose -f $COMPOSE_FILE logs -f redis-cluster-init
    else
        if [[ "$node" =~ ^[1-3]$ ]]; then
            log_info "查看Redis节点 $node 日志..."
            docker-compose -f $COMPOSE_FILE logs -f redis-node-$node
        else
            log_error "无效的节点编号，请使用 1、2、3、init 或 all"
            exit 1
        fi
    fi
}

# 集群健康检查
health_check() {
    log_info "执行Redis集群健康检查..."
    
    local healthy_nodes=0
    local total_nodes=3
    
    for i in {1..3}; do
        local port=$((7000 + i))
        if docker exec -i saas-redis-node-${i}-${APP_ID} redis-cli -p $port ping >/dev/null 2>&1; then
            log_success "节点 $i (端口 $port) 状态正常"
            healthy_nodes=$((healthy_nodes + 1))
        else
            log_error "节点 $i (端口 $port) 状态异常"
        fi
    done
    
    echo ""
    if [[ $healthy_nodes -eq $total_nodes ]]; then
        log_success "集群健康状态：$healthy_nodes/$total_nodes 节点正常"
    else
        log_warning "集群健康状态：$healthy_nodes/$total_nodes 节点正常"
    fi
    
    # 检查集群状态
    if docker exec -i saas-redis-node-1-${APP_ID} redis-cli -p 7001 cluster info 2>/dev/null | grep "cluster_state:ok" >/dev/null; then
        log_success "集群状态：OK"
    else
        log_warning "集群状态：异常"
    fi
}

# 显示帮助信息
show_help() {
    echo "Redis集群管理脚本 (自动初始化版本)"
    echo ""
    echo "用法: $0 [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  start       启动Redis集群（自动初始化）"
    echo "  stop        停止Redis集群"
    echo "  restart     重启Redis集群"
    echo "  status      查看集群状态"
    echo "  init        触发集群重新初始化"
    echo "  connect     连接到集群节点"
    echo "  logs        查看集群日志"
    echo "  health      健康检查"
    echo "  help        显示帮助信息"
    echo ""
    echo "日志选项:"
    echo "  logs all    查看所有服务日志"
    echo "  logs init   查看自动初始化服务日志"
    echo "  logs 1      查看节点1日志"
    echo "  logs 2      查看节点2日志"
    echo "  logs 3      查看节点3日志"
    echo ""
    echo "示例:"
    echo "  $0 start                 # 启动集群（自动初始化）"
    echo "  $0 status                # 查看集群状态"
    echo "  $0 logs init             # 查看初始化进度"
    echo "  $0 connect 1             # 连接到节点1"
    echo "  $0 health                # 健康检查"
    echo ""
    echo "注意:"
    echo "  - 集群启动后会自动初始化，无需手动操作"
    echo "  - 自动初始化服务会持续监控集群健康状态"
    echo "  - 如果集群出现问题，会自动尝试修复"
    echo ""
}

# 主函数
main() {
    check_docker
    
    case "${1:-help}" in
        start)
            start_cluster
            ;;
        stop)
            stop_cluster
            ;;
        restart)
            restart_cluster
            ;;
        status)
            status_cluster
            ;;
        init)
            init_cluster
            ;;
        connect)
            connect_cluster $2
            ;;
        logs)
            logs_cluster $2
            ;;
        health)
            health_check
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@" 