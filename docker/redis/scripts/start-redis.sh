#!/bin/bash

# Redis动态启动脚本 - 支持动态IP配置
set -e

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Redis启动: $*"
}

# 获取宿主机IP地址（如果环境变量未设置）
get_host_ip() {
    if [ -z "$CLUSTER_ANNOUNCE_IP" ] || [ "$CLUSTER_ANNOUNCE_IP" = "auto" ]; then
        # 尝试多种方式获取宿主机IP
        local host_ip=""
        
        # 方式1: 通过路由表获取
        host_ip=$(ip route get 8.8.8.8 | grep -oP 'src \K\S+' 2>/dev/null || true)
        
        # 方式2: 通过网卡接口获取
        if [ -z "$host_ip" ]; then
            host_ip=$(hostname -I | awk '{print $1}' 2>/dev/null || true)
        fi
        
        # 方式3: 通过容器网关获取宿主机IP
        if [ -z "$host_ip" ]; then
            host_ip=$(getent hosts host.docker.internal | awk '{print $1}' 2>/dev/null || true)
        fi
        
        # 方式4: 默认使用网关IP
        if [ -z "$host_ip" ]; then
            host_ip=$(ip route | grep default | awk '{print $3}' | head -1)
        fi
        
        echo "$host_ip"
    else
        echo "$CLUSTER_ANNOUNCE_IP"
    fi
}

# 主函数
main() {
    log "开始启动Redis节点..."
    
    # 获取动态IP
    local announce_ip=$(get_host_ip)
    local announce_port=${CLUSTER_ANNOUNCE_PORT:-7001}
    local announce_bus_port=${CLUSTER_ANNOUNCE_BUS_PORT:-17001}
    
    log "使用配置: IP=$announce_ip, Port=$announce_port, BusPort=$announce_bus_port"
    
    # 创建临时配置文件
    local temp_conf="/tmp/redis.conf"
    cp /etc/redis/redis.conf "$temp_conf"
    
    # 确保文件末尾有换行符
    echo "" >> "$temp_conf"
    
    # 动态替换或添加配置项
    if grep -q "^cluster-announce-ip" "$temp_conf"; then
        sed -i "s/^cluster-announce-ip .*/cluster-announce-ip $announce_ip/" "$temp_conf"
    else
        echo "cluster-announce-ip $announce_ip" >> "$temp_conf"
    fi
    
    if grep -q "^cluster-announce-port" "$temp_conf"; then
        sed -i "s/^cluster-announce-port .*/cluster-announce-port $announce_port/" "$temp_conf"
    else
        echo "cluster-announce-port $announce_port" >> "$temp_conf"
    fi
    
    if grep -q "^cluster-announce-bus-port" "$temp_conf"; then
        sed -i "s/^cluster-announce-bus-port .*/cluster-announce-bus-port $announce_bus_port/" "$temp_conf"
    else
        echo "cluster-announce-bus-port $announce_bus_port" >> "$temp_conf"
    fi
    
    log "✅ 配置文件已更新"
    
    # 启动Redis服务
    log "🚀 启动Redis服务..."
    exec redis-server "$temp_conf"
}

# 执行主函数
main "$@" 