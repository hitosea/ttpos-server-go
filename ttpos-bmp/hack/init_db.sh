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

# 判断DB_ROOT_USERNAME是否为空，如果为空则设置为root
if [ -z "$DB_ROOT_USERNAME" ]; then
  DB_ROOT_USERNAME="root"
fi

# 选择访问方式：
#   DB_DOCKER_NETWORK 非空 -> 工具容器加入该 docker network，使用内部端口 DB_PORT
#   DB_DOCKER_NETWORK 为空 -> 直接通过宿主机映射端口 DB_PORT_OPEN 访问
if [ -n "$DB_DOCKER_NETWORK" ]; then
    NETWORK_OPT="--network $DB_DOCKER_NETWORK"
    DB_PORT_EXEC=${DB_PORT:-3306}
else
    NETWORK_OPT=""
    DB_PORT_EXEC=$DB_PORT_OPEN
fi

export EXEC="mysql -h $DB_HOST -u$DB_ROOT_USERNAME -p$DB_ROOT_PASSWORD -P $DB_PORT_EXEC"

docker run --rm $NETWORK_OPT mysql:8.0 $EXEC -e "CREATE DATABASE IF NOT EXISTS erp DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
docker run --rm $NETWORK_OPT mysql:8.0 $EXEC -e "CREATE DATABASE IF NOT EXISTS takeout DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
docker run --rm $NETWORK_OPT mysql:8.0 $EXEC -e "CREATE DATABASE IF NOT EXISTS message DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
docker run --rm $NETWORK_OPT mysql:8.0 $EXEC -e "CREATE DATABASE IF NOT EXISTS websocket DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
# MySQL 8 不再允许 GRANT 隐式建用户，先 CREATE USER IF NOT EXISTS（不覆盖已存在用户的密码）
docker run --rm $NETWORK_OPT mysql:8.0 $EXEC -e "CREATE USER IF NOT EXISTS '$DB_USERNAME'@'%' IDENTIFIED BY '$DB_PASSWORD';"
docker run --rm $NETWORK_OPT mysql:8.0 $EXEC -e "GRANT ALL PRIVILEGES ON erp.* TO '$DB_USERNAME'@'%'  WITH GRANT OPTION;"
docker run --rm $NETWORK_OPT mysql:8.0 $EXEC -e "GRANT ALL PRIVILEGES ON takeout.* TO '$DB_USERNAME'@'%'  WITH GRANT OPTION;"
docker run --rm $NETWORK_OPT mysql:8.0 $EXEC -e "GRANT ALL PRIVILEGES ON message.* TO '$DB_USERNAME'@'%'  WITH GRANT OPTION;"
docker run --rm $NETWORK_OPT mysql:8.0 $EXEC -e "GRANT ALL PRIVILEGES ON websocket.* TO '$DB_USERNAME'@'%'  WITH GRANT OPTION;"

set +o allexport
cd $old_dir
