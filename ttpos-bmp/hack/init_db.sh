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

export EXEC="mysql -h $DB_HOST -u$DB_ROOT_USERNAME -p$DB_ROOT_PASSWORD -P $DB_PORT_OPEN"

docker run --rm mariadb:10.11.6 $EXEC -e "CREATE DATABASE IF NOT EXISTS erp DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
docker run --rm mariadb:10.11.6 $EXEC -e "CREATE DATABASE IF NOT EXISTS takeout DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
docker run --rm mariadb:10.11.6 $EXEC -e "CREATE DATABASE IF NOT EXISTS message DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
docker run --rm mariadb:10.11.6 $EXEC -e "GRANT ALL PRIVILEGES ON erp.* TO '$DB_USERNAME'@'%'  WITH GRANT OPTION;"
docker run --rm mariadb:10.11.6 $EXEC -e "GRANT ALL PRIVILEGES ON takeout.* TO '$DB_USERNAME'@'%'  WITH GRANT OPTION;"
docker run --rm mariadb:10.11.6 $EXEC -e "GRANT ALL PRIVILEGES ON message.* TO '$DB_USERNAME'@'%'  WITH GRANT OPTION;"

set +o allexport
cd $old_dir
