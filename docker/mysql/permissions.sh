#!/bin/bash

# 等待 MariaDB 启动
sleep 3

# 使用 root 用户连接并授予权限
mysql -uroot -p"${MARIADB_ROOT_PASSWORD}" <<EOF
-- 授予用户完整的 root 权限
GRANT ALL PRIVILEGES ON *.* TO '${MARIADB_USER}'@'%' WITH GRANT OPTION;
GRANT ALL PRIVILEGES ON *.* TO '${MARIADB_USER}'@'localhost' WITH GRANT OPTION;
GRANT ALL PRIVILEGES ON *.* TO '${MARIADB_USER}'@'*' WITH GRANT OPTION;

-- 授予特殊权限
GRANT SUPER ON *.* TO '${MARIADB_USER}'@'%';
GRANT SUPER ON *.* TO '${MARIADB_USER}'@'localhost';
GRANT SUPER ON *.* TO '${MARIADB_USER}'@'*';

-- 授予创建用户权限
GRANT CREATE USER ON *.* TO '${MARIADB_USER}'@'%';
GRANT CREATE USER ON *.* TO '${MARIADB_USER}'@'localhost';
GRANT CREATE USER ON *.* TO '${MARIADB_USER}'@'*';

-- 授予重载权限
GRANT RELOAD ON *.* TO '${MARIADB_USER}'@'%';
GRANT RELOAD ON *.* TO '${MARIADB_USER}'@'localhost';
GRANT RELOAD ON *.* TO '${MARIADB_USER}'@'*';

-- 刷新权限表
FLUSH PRIVILEGES;
EOF

echo "已成功授予用户 ${MARIADB_USER} 完整的 root 权限"