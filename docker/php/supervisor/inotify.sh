#!/bin/bash

# 设置 umask 为新文件/目录的默认权限
umask 0002

chown -R www-data:www-data "./app"
chown -R www-data:www-data "./public"
chown -R www-data:www-data "./runtime"
chown -R www-data:www-data "./extend"
chown -R www-data:www-data "./vendor"
chown -R www-data:www-data "./route"
chown -R www-data:www-data "./bin"
chown -R www-data:www-data "think"

# 启动 inotifywait 来监听目录
inotifywait -m -r -e create --format "%w%f" "./runtime/" |
while read -r FILE
do
  if [ -f "\$FILE" ] || [ -d "\$FILE" ]; then
    # 新创建的是文件或目录
    chown -R www-data:www-data "\$FILE"
    chmod 766 "\$FILE"  # 设置文件或目录的权限，根据需要修改
  fi
done