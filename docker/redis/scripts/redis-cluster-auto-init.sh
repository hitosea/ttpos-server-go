#!/bin/bash

# Redis 集群自动初始化脚本
# 适配新的端口配置 (内部统一使用6379端口)

set -e

REDIS_NODES=${REDIS_NODES:-"10.0.0.30:6379 10.0.0.31:6379 10.0.0.32:6379"}
MAX_WAIT_TIME=60
HEALTH_CHECK_INTERVAL=30

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

# 等待Redis节点启动
wait_for_redis_nodes() {
    log "等待Redis节点启动..."
    local count=0
    
    for node in $REDIS_NODES; do
        local host=$(echo $node | cut -d':' -f1)
        local port=$(echo $node | cut -d':' -f2)
        
        while [ $count -lt $MAX_WAIT_TIME ]; do
            if redis-cli -h $host -p $port ping > /dev/null 2>&1; then
                log "节点 $node 已启动"
                break
            fi
            
            log "等待节点 $node 启动... ($count/$MAX_WAIT_TIME)"
            sleep 2
            count=$((count + 1))
        done
        
        if [ $count -ge $MAX_WAIT_TIME ]; then
            log "❌ 节点 $node 启动超时"
            return 1
        fi
    done
    
    log "✅ 所有Redis节点已启动"
    return 0
}

# 检查集群是否已初始化
is_cluster_initialized() {
    local node=$(echo $REDIS_NODES | awk '{print $1}')
    local host=$(echo $node | cut -d':' -f1)
    local port=$(echo $node | cut -d':' -f2)
    
    local cluster_info=$(redis-cli -h $host -p $port cluster info 2>/dev/null | grep cluster_state || true)
    
    if [ -n "$cluster_info" ]; then
        local state=$(echo $cluster_info | cut -d':' -f2 | tr -d '\r\n')
        if [ "$state" = "ok" ]; then
            log "✅ 集群已初始化且状态正常"
            return 0
        fi
    fi
    
    return 1
}

# 检查集群健康状态
check_cluster_health() {
    local healthy_nodes=0
    local total_nodes=0
    
    for node in $REDIS_NODES; do
        local host=$(echo $node | cut -d':' -f1)
        local port=$(echo $node | cut -d':' -f2)
        total_nodes=$((total_nodes + 1))
        
        if redis-cli -h $host -p $port ping > /dev/null 2>&1; then
            local cluster_info=$(redis-cli -h $host -p $port cluster info 2>/dev/null || true)
            if echo "$cluster_info" | grep -q "cluster_state:ok"; then
                healthy_nodes=$((healthy_nodes + 1))
            fi
        fi
    done
    
    log "集群健康检查: $healthy_nodes/$total_nodes 节点正常"
    
    if [ $healthy_nodes -eq $total_nodes ]; then
        return 0
    else
        return 1
    fi
}

# 初始化集群
initialize_cluster() {
    log "🚀 开始初始化Redis集群..."
    
    # 构建集群创建命令
    local create_cmd="redis-cli --cluster create"
    for node in $REDIS_NODES; do
        create_cmd="$create_cmd $node"
    done
    create_cmd="$create_cmd --cluster-replicas 0 --cluster-yes"
    
    log "执行命令: $create_cmd"
    
    if eval $create_cmd; then
        log "✅ 集群初始化成功"
        
        # 验证集群状态
        sleep 5
        if check_cluster_health; then
            log "✅ 集群健康检查通过"
            return 0
        else
            log "⚠️ 集群初始化完成，但健康检查未通过"
            return 1
        fi
    else
        log "❌ 集群初始化失败"
        return 1
    fi
}

# 主循环 - 持续监控集群状态
maintenance_loop() {
    log "🔄 开始集群维护监控..."
    
    while true; do
        if check_cluster_health; then
            log "✅ 集群状态正常"
        else
            log "⚠️ 检测到集群异常，尝试修复..."
            
            # 尝试重新初始化
            if initialize_cluster; then
                log "✅ 集群修复成功"
            else
                log "❌ 集群修复失败，将在下次检查时重试"
            fi
        fi
        
        sleep $HEALTH_CHECK_INTERVAL
    done
}

# 主函数
main() {
    log "🚀 Redis集群自动初始化脚本启动"
    
    # 等待节点启动
    if ! wait_for_redis_nodes; then
        log "❌ Redis节点启动失败，退出"
        exit 1
    fi
    
    # 检查是否已初始化
    if is_cluster_initialized; then
        log "✅ 集群已存在，开始维护监控"
    else
        log "📦 集群未初始化，开始初始化..."
        if ! initialize_cluster; then
            log "❌ 集群初始化失败，退出"
            exit 1
        fi
    fi
    
    # 进入维护循环
    maintenance_loop
}

# 捕获退出信号
trap 'log "📋 收到退出信号，停止集群维护"; exit 0' TERM INT

# 启动主函数
main 