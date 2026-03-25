#!/bin/bash

# RocketMQ Topic 更新脚本
# 读取 .env 文件中的配置，比较现有 topic 与 manifest/topics.txt 中的差异，并创建缺失的 topic

set -e

# 脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 读取 .env 文件
ENV_FILE="$PROJECT_ROOT/.env"
if [ ! -f "$ENV_FILE" ]; then
    echo "错误: .env 文件不存在: $ENV_FILE"
    echo "请创建 .env 文件并配置以下变量:"
    echo "ROCKETMQ_NAME_SRV_ADDR=127.0.0.1:9876"
    echo "ROCKETMQ_BROKER_ADDR=127.0.0.1:10911"
    exit 1
fi

# 加载环境变量
source "$ENV_FILE"

# 检查必需的环境变量
if [ -z "$ROCKETMQ_NAME_SRV_ADDR" ]; then
    echo "错误: ROCKETMQ_NAME_SRV_ADDR 环境变量未设置"
    exit 1
fi

# 如果 ROCKETMQ_NAME_SRV_ADDR 包含多个逗号分隔的地址，只取第一个
ROCKETMQ_NAME_SRV_ADDR="${ROCKETMQ_NAME_SRV_ADDR%%,*}"

if [ -z "$ROCKETMQ_BROKER_ADDR" ]; then
    echo "错误: ROCKETMQ_BROKER_ADDR 环境变量未设置"
    exit 1
fi

# topics.txt 文件路径
TOPICS_FILE="$PROJECT_ROOT/manifest/topics.txt"
if [ ! -f "$TOPICS_FILE" ]; then
    echo "错误: topics.txt 文件不存在: $TOPICS_FILE"
    exit 1
fi

echo "开始检查和更新 RocketMQ Topics..."
echo "NameServer 地址: $ROCKETMQ_NAME_SRV_ADDR"
echo "Broker 地址: $ROCKETMQ_BROKER_ADDR"

NETWORK_OPTS=""
# 检查 docker 网络中是否存在 ttpos-bmp-mid_bmp-mid-network，如果有则设置 NETWORK_OPTS
if docker network ls | awk '{print $2}' | grep -qw "ttpos-bmp-mid_bmp-mid-network"; then
    NETWORK_OPTS="--network ttpos-bmp-mid_bmp-mid-network"
fi



# 获取现有的 topic 列表
echo "正在获取现有 topic 列表..."
EXISTING_TOPICS=$(timeout 15 docker run $NETWORK_OPTS --rm apache/rocketmq:5.3.4 ./mqadmin topicList -n "$ROCKETMQ_NAME_SRV_ADDR" 2>/dev/null | grep -v "^#" | grep -v "^$" | awk '{print $1}' | sort) || true

if [ $? -ne 0 ]; then
    echo "错误: 无法获取现有 topic 列表，请检查 RocketMQ 服务是否正常运行"
    exit 1
fi

echo "现有 topics:"
echo "$EXISTING_TOPICS"
echo ""

# 读取需要的 topic 列表
echo "需要的 topics (来自 $TOPICS_FILE):"
REQUIRED_TOPICS=$(cat "$TOPICS_FILE" | grep -v "^#" | grep -v "^$" | sort)
echo "$REQUIRED_TOPICS"
echo ""

# 比较并创建缺失的 topic
MISSING_TOPICS=""
while IFS= read -r topic; do
    if [ -n "$topic" ]; then
        if ! echo "$EXISTING_TOPICS" | grep -q "^$topic$"; then
            MISSING_TOPICS="$MISSING_TOPICS$topic\n"
        fi
    fi
done <<< "$REQUIRED_TOPICS"

if [ -z "$MISSING_TOPICS" ]; then
    echo "所有 topics 都已存在，无需创建新的 topic"
else
    echo "发现缺失的 topics，开始创建:"
    echo -e "$MISSING_TOPICS"
    
    while IFS= read -r topic; do
        if [ -n "$topic" ]; then
            echo "正在创建 topic: $topic"
            OUTPUT=$(timeout 30 docker run $NETWORK_OPTS --rm apache/rocketmq:5.3.4 ./mqadmin updateTopic \
                -n "$ROCKETMQ_NAME_SRV_ADDR" \
                -t "$topic" \
                -p 6 \
                -r 4 \
                -w 4 \
                -c DefaultCluster 2>&1) || true
            echo "$OUTPUT"

            if echo "$OUTPUT" | grep -q "success"; then
                echo "✓ 成功创建 topic: $topic"
            else
                echo "✗ 创建 topic 失败: $topic"
            fi
        fi
    done <<< "$(echo -e "$MISSING_TOPICS")"
fi

echo ""
echo "Topic 更新完成!"

