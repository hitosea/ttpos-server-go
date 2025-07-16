#!/bin/bash

# ====== 配置start ======
# 项目根目录
current_dir=$(pwd)
# 翻译文件存放目录
target_path="$current_dir/src/locales"
# 默认语言
default_lang="zh"
# 需要翻译的语言列表
language_list=("sv")
# 翻译平台 URL
translate_url="https://aitrans.ttpos.com/translate"
# ====== 配置end ======

# 添加日志功能
log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# 获取需要翻译的文件
get_source_file() {
    local source_file="$target_path/$default_lang.json"
    if [ ! -f "$source_file" ]; then
        log_message "Error: 源语言文件不存在: $source_file"
        exit 1
    fi
    echo "$source_file"
}

# 发送翻译请求
translate_text() {
    local text="$1"
    local target_lang="$2"
    local max_retries=3  # 最大重试次数
    local retry_count=0
    local translation=""
    
    while [ $retry_count -lt $max_retries ]; do
        log_message "正在翻译: $text -> $target_lang (尝试 $((retry_count+1))/$max_retries)" >&2
        
        local request_data=$(jq -n --arg text "$text" --arg lang "$default_lang" --argjson trans "$(printf '%s\n' "${language_list[@]}" | jq -R . | jq -s .)" '{
            "data": [{
                "lang": $lang,
                "content": $text
            }],
            "trans": $trans
        }')
        
        # 发送请求
        local response=$(curl -sS -X POST "$translate_url" \
            -H "Content-Type: application/json" \
            -d "$request_data")
        
        log_message "翻译结果: $response" >&2
        
        # 检查是否成功获取响应
        if [ -z "$response" ] || ! echo "$response" | jq -e . >/dev/null 2>&1; then
            log_message "警告: 翻译接口返回无效响应，准备重试..." >&2
            ((retry_count++))
            sleep 1  # 添加延迟避免频繁重试
            continue
        fi
        
        # 检查是否有错误码
        local error_code=$(echo "$response" | jq -r '.code // empty')
        if [ "$error_code" != "200" ] && [ -n "$error_code" ]; then
            log_message "警告: 翻译接口返回错误码 $error_code，准备重试..." >&2
            ((retry_count++))
            sleep 1
            continue
        fi
        
        # 尝试提取目标语言
        translation=$(echo "$response" | jq -r --arg text "$text" --arg lang "$target_lang" '
            .data[] | select(.key == $text) | .[$lang] // empty
        ' 2>/dev/null)
        
        # 如果成功获取到翻译，退出重试循环
        if [ -n "$translation" ] && [ "$translation" != "null" ]; then
            break
        fi
        
        # 如果目标语言不存在，尝试英文
        if [ -z "$translation" ] || [ "$translation" = "null" ]; then
            log_message "警告: 无法提取 $target_lang 翻译，尝试英文" >&2
            translation=$(echo "$response" | jq -r --arg text "$text" '
                .data[] | select(.key == $text) | .en // empty
            ' 2>/dev/null)
            
            if [ -n "$translation" ] && [ "$translation" != "null" ]; then
                break
            fi
        fi
        
        ((retry_count++))
        if [ $retry_count -lt $max_retries ]; then
            log_message "警告: 无法提取翻译结果，准备重试 ($retry_count/$max_retries)..." >&2
            sleep 1
        fi
    done
    
    # 如果还是失败，返回原始文本
    if [ -z "$translation" ] || [ "$translation" = "null" ]; then
        log_message "警告: 达到最大重试次数，使用原始文本" >&2
        echo -n "$text"
    else
        echo -n "$translation"
    fi
}

# 处理单个语言文件
process_language() {
    local lang="$1"
    local source_file="$target_path/$default_lang.json"
    local target_file="$target_path/$lang.json"
    
    log_message "开始处理语言: $lang"
    
    # 创建临时文件
    local temp_file=$(mktemp)
    local processed_keys=0
    
    # 初始化目标 JSON 对象
    echo "{" > "$temp_file"
    
    # 使用 jq 处理 JSON 文件
    while IFS= read -r entry; do
        local key=$(echo "$entry" | jq -r '.key')
        local value=$(echo "$entry" | jq -r '.value')
        
        # 如果目标文件已存在，检查是否已有翻译
        local translated_value=""
        if [ -f "$target_file" ]; then
            translated_value=$(jq -r --arg k "$key" '.[$k] // empty' "$target_file" 2>/dev/null)
        fi
        
        # 如果没有现有翻译，则进行翻译
        if [ -z "$translated_value" ] || [ "$translated_value" = "null" ]; then
            log_message "翻译中: $key"
            translated_value=$(translate_text "$value" "$lang")
            # 转义特殊字符并移除可能的换行符
            translated_value=$(echo -n "$translated_value" | tr -d '\n' | sed 's/"/\\"/g')
        fi

        # 添加逗号分隔符（除了第一个元素）
        if [ $processed_keys -gt 0 ]; then
            echo "," >> "$temp_file"
        fi
        
        # 写入键值对
        echo -n "  \"$key\": \"$translated_value\"" >> "$temp_file"
        ((processed_keys++))
        
    done < <(jq -c 'to_entries[]' "$source_file")
    
    # 结束 JSON 对象
    echo -e "\n}" >> "$temp_file"
    
    # 格式化 JSON 并保存
    jq . "$temp_file" > "$target_file" 2>/dev/null || {
        log_message "警告: 无法格式化 JSON，保存原始内容"
        mv "$temp_file" "$target_file"
    }
    
    # 清理临时文件
    [ -f "$temp_file" ] && rm -f "$temp_file"
    
    log_message "完成处理: $lang"
}

# 主函数
main() {
    # 检查 jq 是否安装
    if ! command -v jq &> /dev/null; then
        log_message "Error: 请先安装 jq 工具 (brew install jq)"
        exit 1
    fi
    
    # 获取源文件
    source_file=$(get_source_file)
    log_message "开始翻译，源文件: $source_file"
    
    # 遍历所有目标语言
    for lang in "${language_list[@]}"; do
        if [ "$lang" != "$default_lang" ]; then
            process_language "$lang"
        fi
    done
    
    log_message "所有翻译完成"
}

# 运行主函数
main