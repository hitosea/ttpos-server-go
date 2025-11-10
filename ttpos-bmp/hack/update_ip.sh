#!/bin/bash

LOCAL_IP=`ifconfig | grep "inet " | grep "192" | awk '{print $2}' | head -n 1`

if [ -z "$LOCAL_IP" ] ;then
  # 兼容 coder 环境
  LOCAL_IP=`ifconfig | grep "inet " | grep "172.1" | awk '{print $2}' | head -n 1`
fi

ENV_FILE=../.env

if [ -f "$ENV_FILE" ] ;then
  echo "替换中台.env 本地IP"
  #替换本地IP
  sed -i.bak "s/^DB_HOST=.*/DB_HOST=${LOCAL_IP}/" .env && rm .env.bak
  sed -i.bak "s/^GRPC_ENDPOINTS=.*/GRPC_ENDPOINTS=${LOCAL_IP}/" .env && rm .env.bak
  sed -i.bak "s/^REDIS_HOST=.*/REDIS_HOST=${LOCAL_IP}/" .env && rm .env.bak
  if grep -q '^REDIS_CLUSTER_ANNOUNCE_IP=' .env; then
    sed -i.bak "s/^REDIS_CLUSTER_ANNOUNCE_IP=.*/REDIS_CLUSTER_ANNOUNCE_IP=${LOCAL_IP}/" .env && rm .env.bak
  else
    echo "REDIS_CLUSTER_ANNOUNCE_IP=${LOCAL_IP}" >> .env
  fi
  sed -i.bak "s/^NACOS_SERVER_IP=.*/NACOS_SERVER_IP=${LOCAL_IP}/" .env && rm .env.bak
   echo "替换中台.env 本地IP成功 $LOCAL_IP"
fi