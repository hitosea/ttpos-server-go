#!/bin/sh

# 设置错误处理
set -e

# 创建数据库
mysql -u$MARIADB_USER -p$MARIADB_PASSWORD -h$MARIADB_HOST -P$MARIADB_PORT <<EOF
CREATE DATABASE IF NOT EXISTS saas CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
EOF

# 检查数据库是否创建成功
if mysql -u$MARIADB_USER -p$MARIADB_PASSWORD -h$MARIADB_HOST -P$MARIADB_PORT -e "USE saas;" 2>/dev/null; then
    echo "数据库 saas 创建成功"
else
    echo "错误：数据库 saas 创建失败"
    exit 1
fi

# 检查字符集是否正确设置
CHARSET=$(mysql -u$MARIADB_USER -p$MARIADB_PASSWORD -h$MARIADB_HOST -P$MARIADB_PORT -e "SELECT DEFAULT_CHARACTER_SET_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = 'saas';" | tail -n 1)
COLLATION=$(mysql -u$MARIADB_USER -p$MARIADB_PASSWORD -h$MARIADB_HOST -P$MARIADB_PORT -e "SELECT DEFAULT_COLLATION_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = 'saas';" | tail -n 1)

if [ "$CHARSET" = "utf8mb4" ] && [ "$COLLATION" = "utf8mb4_unicode_ci" ]; then
    echo "字符集和排序规则设置正确"
else
    echo "警告：字符集或排序规则设置不正确"
    echo "当前字符集: $CHARSET"
    echo "当前排序规则: $COLLATION"
fi