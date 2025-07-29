#!/bin/sh

# 检查 envsubst 是否安装，如果没有安装则自动安装
check_and_install_envsubst() {
    if ! command -v envsubst >/dev/null 2>&1; then
        echo "envsubst 未安装，正在自动安装..."
        
        # 检测操作系统类型
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            # Linux 系统
            if command -v apt-get >/dev/null 2>&1; then
                # Ubuntu/Debian
                sudo apt-get update && sudo apt-get install -y gettext-base
            elif command -v yum >/dev/null 2>&1; then
                # CentOS/RHEL
                sudo yum install -y gettext
            elif command -v dnf >/dev/null 2>&1; then
                # Fedora
                sudo dnf install -y gettext
            elif command -v apk >/dev/null 2>&1; then
                # Alpine
                sudo apk add --no-cache gettext
            else
                echo "错误：无法识别的Linux发行版，请手动安装 gettext 包"
                exit 1
            fi
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS 系统
            if command -v brew >/dev/null 2>&1; then
                brew install gettext
                # 在 macOS 上，gettext 可能不会自动添加到 PATH
                if ! command -v envsubst >/dev/null 2>&1; then
                    echo "正在配置 gettext 路径..."
                    export PATH="/usr/local/opt/gettext/bin:$PATH"
                fi
            else
                echo "错误：请先安装 Homebrew，然后运行: brew install gettext"
                exit 1
            fi
        else
            echo "错误：不支持的操作系统类型: $OSTYPE"
            exit 1
        fi
        
        # 再次检查是否安装成功
        if command -v envsubst >/dev/null 2>&1; then
            echo "envsubst 安装成功"
        else
            echo "错误：envsubst 安装失败"
            exit 1
        fi
    else
        echo "envsubst 已安装"
    fi
}

# 执行检查和安装
check_and_install_envsubst

old_dir=$OLDPWD
cd `dirname $0`
set -o allexport
#引入环境变量
. ../../.env
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