#!/bin/sh
 
GreenBG="\033[42;37m"
Font="\033[0m"


salt_hash() {
 password=$1
 hashed_password=$(echo -n "$(echo -n "$password" | md5sum | awk '{print $1}')"jjjshop_salt_2020 | md5sum | awk '{print $1}')
 echo $hashed_password
}

if [ -z "$1" ]; then
 new_password=$(date +%s%N | md5sum | awk '{print $1}' | cut -c 1-16)
else
 new_password=$1
fi

md5_password=$(salt_hash $new_password)

content=$(echo "select \`admin_user_id\` from ${MARIADB_PREFIX}admin_user order by \`admin_user_id\` limit 1;" | mysql -u$MARIADB_USER -p$MARIADB_PASSWORD $MARIADB_DATABASE -h$MARIADB_HOST)
userid=$(echo "$content" | sed -n '2p')

if [ -z "$userid" ]; then
 echo "错误：账号不存在！"
 exit 1
fi

mysql -u$MARIADB_USER -p$MARIADB_PASSWORD $MARIADB_DATABASE -h$MARIADB_HOST <<EOF
update ${MARIADB_PREFIX}admin_user set \`username\`='admin',\`password\`='${md5_password}' where \`admin_user_id\` = '${userid}';
EOF
account="admin"
echo "账号: ${GreenBG}${account}${Font}"
echo "密码: ${GreenBG}${new_password}${Font}"