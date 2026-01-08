#!/bin/sh

# 获取默认路由的网关地址
DEFAULT_GATEWAY=$(ip r | grep "^default" | awk '{print $3}' | head -n 1)

if [ -n "$DEFAULT_GATEWAY" ]; then
    # 提取默认路由的前3段（例如：192.168.1.1 -> 192.168.1）
    NETWORK_PREFIX=$(echo $DEFAULT_GATEWAY | cut -d '.' -f 1-3)
    
    # 通过 ip a 查找该网段的 IP 地址，排除 docker0 和虚拟网卡
    LOCAL_IP=$(ip a | grep -v "docker0\|veth\|br-" | grep "inet " | grep "$NETWORK_PREFIX" | awk '{print $2}' | cut -d '/' -f 1 | head -n 1)
fi

# 如果没有找到，尝试使用 ifconfig 兼容逻辑（排除 docker0）
if [ -z "$LOCAL_IP" ]; then
    LOCAL_IP=$(ifconfig 2>/dev/null | grep -v "docker0" | grep "inet " | grep -E "192\.|172\.1" | awk '{print $2}' | head -n 1)
fi

# 最后才尝试 docker0 作为兼容
if [ -z "$LOCAL_IP" ]; then
    LOCAL_IP=$(ip addr show docker0 2>/dev/null | grep "inet " | awk '{print $2}' | cut -d '/' -f 1 | head -n 1)
fi

echo $LOCAL_IP