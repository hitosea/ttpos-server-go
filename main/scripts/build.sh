#!/bin/bash

# 设置输出二进制文件名
OUTPUT_NAME="server"

# 确保bin目录存在
mkdir -p bin

# 检查是否需要包含swagger
if [ "$1" == "swagger" ]; then
    echo "构建包含 swagger 的版本..."
    # 生成Swagger文档
    swag init
    # 添加 -v 参数来查看详细的构建信息
    go build -ldflags="-s -w -X ttpos-server-go/version.Version=$(grep '^VERSION=' ../.env | cut -d'=' -f2) -X ttpos-server-go/version.CommitSHA=$(git rev-parse HEAD) -X ttpos-server-go/version.BuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" -v -tags swagger -o "bin/${OUTPUT_NAME}" main.go
    if [ $? -ne 0 ]; then
        echo "构建失败！"
        exit 1
    fi
else
    echo "构建不包含 swagger 的版本..."
    go build -ldflags="-s -w -X ttpos-server-go/version.Version=$(grep '^VERSION=' ../.env | cut -d'=' -f2) -X ttpos-server-go/version.CommitSHA=$(git rev-parse HEAD) -X ttpos-server-go/version.BuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" -v -tags "!swagger" -o "bin/${OUTPUT_NAME}" main.go
    if [ $? -ne 0 ]; then
        echo "构建失败！"
        exit 1
    fi
fi

echo "构建完成: bin/${OUTPUT_NAME}" 