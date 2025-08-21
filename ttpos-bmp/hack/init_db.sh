#!/bin/bash
set -o allexport

ENV_FILE=.env

if [ ! -f "$ENV_FILE" ] ;then
  echo "环境变量文件不存在，试用上级目录"
  ENV_FILE=../.env
fi

. "$ENV_FILE"

export EXEC="mysql -h $DB_HOST -uroot -p$DB_ROOT_PASSWORD -P $DB_PORT_OPEN"
docker run --rm mariadb:10.11.6 $EXEC -e "CREATE DATABASE IF NOT EXISTS erp DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"
docker run --rm mariadb:10.11.6 $EXEC -e "CREATE DATABASE IF NOT EXISTS takeout DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"
docker run --rm mariadb:10.11.6 $EXEC -e "GRANT ALL PRIVILEGES ON erp.* TO '$DB_USERNAME'@'%'  WITH GRANT OPTION;"
docker run --rm mariadb:10.11.6 $EXEC -e "GRANT ALL PRIVILEGES ON takeout.* TO '$DB_USERNAME'@'%'  WITH GRANT OPTION;"

set +o allexport
