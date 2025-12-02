#!/bin/sh


# 判断操作系统类型
OS_TYPE=$(uname -s)

# 从 eth 开头的网卡获取 IP（仅 Linux 系统）
LOCAL_IP=""
if [ "$OS_TYPE" = "Linux" ]; then
  # 遍历所有 eth 开头的网卡
  for iface in $(ls /sys/class/net/ 2>/dev/null | grep '^eth'); do
    IP=$(ifconfig $iface 2>/dev/null | grep "inet " | awk '{print $2}' | head -n 1)
    if [ -n "$IP" ]; then
      LOCAL_IP=$IP
      break
    fi
  done
fi

# 如果 eth 网卡没有找到 IP，使用原来的逻辑作为兼容
if [ -z "$LOCAL_IP" ] ;then
  LOCAL_IP=`ifconfig | grep "inet " | grep "192" | awk '{print $2}' | head -n 1`
fi

if [ -z "$LOCAL_IP" ] ;then
  # 兼容 coder 环境
  LOCAL_IP=`ifconfig | grep "inet " | grep "172.1" | awk '{print $2}' | head -n 1`
fi

echo $LOCAL_IP