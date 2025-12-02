#!/bin/bash
old_dir=$OLDPWD
cd `dirname $0`
set -o allexport
ENV_FILE=../.env

if [ ! -f "$ENV_FILE" ] ;then
  echo "环境变量文件不存在，试用上级目录"
  ENV_FILE=../../.env
fi

. "$ENV_FILE"

# 处理 RocketMQ NameServer 地址列表（支持集群配置）
# 如果 ROCKETMQ_NAME_SRV_ADDR 包含逗号或分号，则格式化为 YAML 列表项
if [ -n "$ROCKETMQ_NAME_SRV_ADDR" ]; then
    # 将分号替换为逗号，统一分隔符
    rocketmq_addrs=$(echo "$ROCKETMQ_NAME_SRV_ADDR" | sed 's/;/,/g')
    # 检查是否包含逗号（即多个地址）
    if echo "$rocketmq_addrs" | grep -q ','; then
        # 多地址：分割并格式化为 "addr1", "addr2" 格式
        formatted_addrs=""
        IFS=',' read -ra addrs <<< "$rocketmq_addrs"
        for addr in "${addrs[@]}"; do
            # 清理地址前后空格
            addr=$(echo "$addr" | xargs)
            if [ -n "$addr" ]; then
                if [ -n "$formatted_addrs" ]; then
                    formatted_addrs="$formatted_addrs, \"$addr\""
                else
                    formatted_addrs="\"$addr\""
                fi
            fi
        done
        export ROCKETMQ_NAME_SRV_ADDR="$formatted_addrs"
    else
        # 单地址：添加引号
        addr=$(echo "$rocketmq_addrs" | xargs)
        export ROCKETMQ_NAME_SRV_ADDR="\"$addr\""
    fi
    # echo "ROCKETMQ_NAME_SRV_ADDR=$ROCKETMQ_NAME_SRV_ADDR"
fi

# 遍历app目录下的所有子目录（假设init_conf.sh位于hack目录，app目录在上级目录）
for app_dir in ../app/*; do
    # 仅处理目录类型
    if [ -d "$app_dir" ]; then
        # 定义模板文件路径（根据用户提供的config.tpl.yaml路径结构）
        tpl_file="$app_dir/manifest/config/config.tpl.yaml"
        # 目标文件路径
        target_file="$app_dir/manifest/config/config.yaml"

        # 检查模板文件是否存在
        if [ -f "$tpl_file" ]; then
            # 执行复制操作
            cp "$tpl_file" "$target_file"
             # 处理REDIS地址组合逻辑（如果ENV_FILE中未配置REDIS_ADDRESSES则自动生成）
            if [ -z "$REDIS_ADDRESSES" ] && [ -n "$REDIS_CLUSTER_ANNOUNCE_IP" ] && [ -n "$REDIS_PORT" ]; then
                # 处理REDIS_CLUSTER_ANNOUNCE_IP为auto的情况
                if [ "$REDIS_CLUSTER_ANNOUNCE_IP" = "auto" ]; then
                    local_ip=$(../ttpos-scripts/get_ip.sh)
                    # 设置为获取到的IP
                    export REDIS_CLUSTER_ANNOUNCE_IP="$local_ip"
                fi
                # 分割端口列表
                IFS=',' read -ra ports <<< "$REDIS_PORT"
                addresses=()
                for port in "${ports[@]}"; do
                    # 清理端口前后空格
                    port=$(echo "$port" | xargs)
                    # 拼接IP和端口
                    if [ -n "$port" ]; then
                        addresses+=("$REDIS_CLUSTER_ANNOUNCE_IP:$port")
                    fi
                done
                # 导出组合后的地址列表（供envsubst使用）
                export REDIS_ADDRESSES=$(IFS=','; echo "${addresses[*]}")
                # echo "REDIS_ADDRESSES=$REDIS_ADDRESSES"
            fi

            # 替换文件中$开头的环境变量为实际值
            envsubst < "$target_file" > "$target_file.tmp" && mv "$target_file.tmp" "$target_file"
            echo "成功复制配置文件：$tpl_file -> $target_file"
        else
            echo "警告：模板文件不存在，跳过目录 $app_dir(路径：$tpl_file)"
        fi

        tpl_file="$app_dir/hack/config.tpl.yaml"
        # 目标文件路径
        target_file="$app_dir/hack/config.yaml"

        # 检查模板文件是否存在
        if [ -f "$tpl_file" ]; then
            # 执行复制操作
            cp "$tpl_file" "$target_file"
            envsubst < "$target_file" > "$target_file.tmp" && mv "$target_file.tmp" "$target_file"
            echo "成功复制gen配置文件：$tpl_file -> $target_file"
        else
            echo "警告：模板文件不存在，跳过目录 $app_dir(路径：$tpl_file)"
        fi
    fi
done
set +o allexport
cd $old_dir
