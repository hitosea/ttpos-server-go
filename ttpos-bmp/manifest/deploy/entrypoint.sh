#!/bin/bash

# 判断APP_NAME环境变量是否存在
if [ -z "$APP_NAME" ]; then
  echo "错误：APP_NAME环境变量不存在" >&2
  exit 1
fi

# APP_NAME存在时执行应用
echo "启动应用：$APP_NAME"

# 执行应用启动命令
./${APP_NAME} "$@"