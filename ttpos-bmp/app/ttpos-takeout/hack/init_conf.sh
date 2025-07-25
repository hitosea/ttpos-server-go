#!/bin/sh
old_dir=$OLDPWD
cd `dirname $0`
set -o allexport
#引入环境变量
. ../../../../.env
app_dir=`dirname $(pwd)`
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
            echo "成功复制配置文件：$tpl_file -> $target_file"
        else
            echo "警告：模板文件不存在，跳过目录 $app_dir(路径：$tpl_file)"
        fi
    fi
set +o allexport
cd $old_dir