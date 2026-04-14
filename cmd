#!/bin/bash

#fonts color
Green="\033[32m"
Red="\033[31m"
GreenBG="\033[42;37m"
RedBG="\033[41;37m"
Font="\033[0m"

#notification information
OK="${Green}[OK]${Font}"
Error="${Red}[错误]${Font}"

cur_path="$(pwd)"
COMPOSE="docker-compose -p ttpos-server-go -f docker-compose.production.yml"

judge() {
    if [[ 0 -eq $? ]]; then
        success "$1 完成"
        sleep 1
    else
        error "$1 失败"
        exit 1
    fi
}

success() {
    echo -e "${OK} ${GreenBG}$1${Font}"
}

warning() {
    echo -e "${Warn} ${YellowBG}$1${Font}"
}

error() {
    echo -e "${Error} ${RedBG}$1${Font}"
}

docker_name() {
    echo `$COMPOSE ps | awk '{print $1}' | grep "\-$1\-"`
}

env_init(){
    # 初始化env文件 
    if [ ! -f ".env" ]; then 
        cp .env.example .env; 
        success "Created .env file from .env.example"; 
        sed -i 's/^APP_ID=.*/APP_ID='$(openssl rand -hex 3)'/' .env 
    fi 
    if [ -z $(env_get DB_REDIS_TYPE) ]; then
        read -p "是否使用项目中的数据库，请输入 y 或 n: " input
        # Check the input and echo the result
        if [ "$input" = "y" ]; then
            sed -i 's/^DB_REDIS_TYPE=.*/DB_REDIS_TYPE=local/' .env 
            sed -i 's/^DB_HOST=.*/DB_HOST=db/' .env 
            sed -i 's/^DB_PORT=.*/DB_PORT=3306/' .env 
            sed -i 's/^REDIS_HOST=.*/REDIS_HOST=redis/' .env 
            sed -i 's/^REDIS_PORT=.*/REDIS_PORT=6379/' .env
            sed -i 's/^DB_PASSWORD=.*/DB_PASSWORD='$(openssl rand -hex 8)'/' .env 
            sed -i 's/^DB_ROOT_PASSWORD=.*/DB_ROOT_PASSWORD='$(openssl rand -hex 8)'/' .env 
        elif [ "$input" = "n" ]; then
            sed -i 's/^DB_REDIS_TYPE=.*/DB_REDIS_TYPE=remote/' .env 
            success "请自行修改.env文件中的数据库、reids连接信息和修改docker/mysql-proxy/conf.d/stream.conf文件，然后重新运行: bash cmd install"
            exit 1
        else
            error "输入无效，请输入 y 或 n"
            exit 1
        fi
    elif [ $(env_get DB_REDIS_TYPE) = "local" ]; then
        sed -i 's/^DB_HOST=.*/DB_HOST=db/' .env 
        sed -i 's/^DB_PORT=.*/DB_PORT=3306/' .env 
        sed -i 's/^REDIS_HOST=.*/REDIS_HOST=redis/' .env 
        sed -i 's/^DB_PASSWORD=.*/DB_PASSWORD='$(openssl rand -hex 8)'/' .env 
        sed -i 's/^DB_ROOT_PASSWORD=.*/DB_ROOT_PASSWORD='$(openssl rand -hex 8)'/' .env 
    elif [ $(env_get DB_REDIS_TYPE) = "remote" ]; then
        read -p "请确认是否已修改.env文件中的数据库、reids连接信息和修改docker/mysql-proxy/conf.d/stream.conf文件，是请输入 y; 否请输入 n: " input
        if [ "$input" = "y" ]; then
            success "数据库和reids连接配置信息已修改"
        elif [ "$input" = "n" ]; then
            success "请自行修改.env文件中的数据库、reids连接信息和修改docker/mysql-proxy/conf.d/stream.conf文件，然后重新运行: bash cmd install"
            exit 1
        else
            error "输入无效，请输入 y 或 n"
            exit 1
        fi
    fi
}

env_get() {
    local key=$1
    local value=`cat ${cur_path}/.env | grep "^$key=" | awk -F '=' '{print $2}'`
    echo "$value"
}

env_set() {
    local key=$1
    local val=$2
    local exist=`cat ${cur_path}/.env | grep "^$key="`
    if [ -z "$exist" ]; then
        echo "$key=$val" >> $cur_path/.env
    else
        if [[ `uname` == 'Linux' ]]; then
            sed -i "/^${key}=/c\\${key}=${val}" ${cur_path}/.env
        else
            docker run -it --rm -v ${cur_path}:/www alpine sh -c "sed -i "/^${key}=/c\\${key}=${val}" /www/.env"
        fi
        if [ $? -ne  0 ]; then
            error "设置env参数失败！"
            exit 1
        fi
    fi
}

run_exec() {
    local container=$1
    local cmd=$2
    local name=`docker_name $container`
    if [ -z "$name" ]; then
        echo -e "${Error} ${RedBG} 没有找到 $container 容器! ${Font}"
        exit 1
    fi
    if [ "$container" = "mariadb" ] || [ "$container" = "nginx" ] || [ "$container" = "redis" ]; then
        docker exec -t "$name" /bin/sh -c "$cmd"
    else
        docker exec -t "$name" /bin/bash -c "$cmd"
    fi
}

#
run_mysql() {
    username="root"
    password=$(env_get DB_ROOT_PASSWORD)
    database=$(env_get DB_DATABASE)
    prefix=$(env_get DB_PREFIX)
    host=$(env_get DB_HOST)
    port=$(env_get DB_PORT)
    
    # 开启端口
    if [ "$1" = "open" ]; then
        container_name=`docker_name db`
        echo "$container_name";
        if [ -z "$container_name" ]; then
            error "没有找到 mariadb 容器!"
            exit 1
        fi
        mkdir -p ${cur_path}/docker/mysql/tmp
        cat > ${cur_path}/docker/mysql/tmp/${container_name}.conf <<EOF
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log notice;
pid /var/run/nginx.pid;
events {
    worker_connections 1024;
}
stream {
    upstream mysql {
        server ${container_name}:3306 max_fails=1 fail_timeout=30s;
    }
    server {
        listen 3306;
        proxy_pass mysql;
        proxy_connect_timeout 5s;
    }
}
EOF
        default_value="$(env_get DB_PORT_OPEN)"
        if [ -z "$default_value" ]; then
            read_tip="请输入代理端口 (3300-65500): "
            read -rp "$read_tip" inputport
        fi
        inputport=${inputport:-$default_value}
        if [ $inputport -lt 3300 ] || [ $inputport -gt 65500 ]; then
            error "端口范围不正确！"
            exit 1
        fi
        env_set DB_PORT $inputport
        env_set DB_PORT_OPEN $inputport
        run_mysql rm-port
        container_network=$(docker inspect --format='{{range .NetworkSettings.Networks}}{{.NetworkID}}{{end}}' ${container_name})
        docker run --name ${container_name}-port \
            --network ${container_network} \
            -p ${inputport}:3306 \
            -v ${cur_path}/docker/mysql/tmp/${container_name}.conf:/etc/nginx/nginx.conf \
            -d nginx:alpine > /dev/null
        judge "开启代理"

    # 关闭端口
    elif [ "$1" = "close" ]; then
        container_name=`docker_name db`
        if [ -z "$container_name" ]; then
            error "没有找到 mariadb 容器!"
            exit 1
        fi
        docker stop ${container_name}-port > /dev/null
        docker rm ${container_name}-port > /dev/null
        judge "关闭代理"

    # 删除端口
    elif [ "$1" = "rm-port" ]; then
        docker rm -f $(docker_name db)-port &> /dev/null
    fi
}

arg_get() {
    local find="n"
    local value=""
    for var in $cur_arg; do
        if [[ "$find" == "y" ]]; then
            if [[ ! $var =~ "--" ]]; then
                value=$var
            fi
            break
        fi
        if [[ "--$1" == "$var" ]] || [[ "-$1" == "$var" ]]; then
            find="y"
            value="yes"
        fi
    done
    echo $value
}

compare_env_vars_only_check() {
    local example_file=".env.example"
    local env_file=".env"
    local missing_keys=()
    if [ ! -f "$example_file" ] || [ ! -f "$env_file" ]; then
        echo -e "${Error} ${RedBG} .env 或 .env.example 文件不存在，无法对比！${Font}"
        exit 1
    fi
    while IFS= read -r line; do
        # 跳过注释和空行
        [[ "$line" =~ ^#.*$ || -z "$line" ]] && continue
        key=$(echo "$line" | cut -d'=' -f1)
        [[ -z "$key" ]] && continue
        if ! grep -q "^$key=" "$env_file"; then
            missing_keys+=("$key")
        fi
    done < "$example_file"
    if [ ${#missing_keys[@]} -ne 0 ]; then
        echo -e "${Error} ${RedBG} 检查到以下环境变量在 .env 中缺失，请补充后再执行 update：${Font}"
        for key in "${missing_keys[@]}"; do
            echo -e "  - $key"
        done
        exit 1
    else
        echo -e "${OK} ${GreenBG} .env 文件已包含所有 .env.example 的环境变量，无需补充 ${Font}"
    fi
}

####################################################################################
####################################################################################
####################################################################################

if [ $# -gt 0 ]; then
    if [[ "$1" == "init" ]] || [[ "$1" == "install" ]]; then
        shift 1
        #
        env_init
        $COMPOSE up -d 
        run_exec php "composer install --ignore-platform-reqs"
        echo -e "${OK} ${GreenBG} 初始化数据库 ${Font}"
        # 
        sleep 10
        create=`run_exec db "sh /etc/mysql/create_saas.sh"`
        echo -e "$create"
        run_exec php "php think migrate:run"
        #
        run_exec php "chown -R www-data:www-data ./app"
        run_exec php "chown -R www-data:www-data ./public"
        run_exec php "chown -R www-data:www-data ./runtime"
        run_exec php "chown -R www-data:www-data ./extend"
        run_exec php "chown -R www-data:www-data ./vendor"
        run_exec php "chown -R www-data:www-data ./route"
        run_exec php "chown -R www-data:www-data ./bin"
        run_exec php "chown -R www-data:www-data think"
        run_exec php "chmod +x ./bin/license.so"
        run_exec php "chmod +x ./bin/license_arm.so"
        # 
        $COMPOSE restart
        #
        echo -e "${OK} ${GreenBG} 安装完成 ${Font}"
        echo -e "地址: ${GreenBG}http://127.0.0.1:$(env_get NGINX_PORT)${Font}"
        # 设置云部署初始化密码
        res=`run_exec db "sh /etc/mysql/repassword.sh"`
        echo -e "$res"
    elif [[ "$1" == "update" ]]; then
        shift 1
        if [[ "$@" != "nobackup" ]]; then
            run_mysql backup
        fi
            echo "当前分支: $(git branch | sed -n -e 's/^\* \(.*\)/\1/p')"
            if [[ -z "$(arg_get local)" ]]; then
                git fetch --all
                git reset --hard origin/$(git branch | sed -n -e 's/^\* \(.*\)/\1/p')
            if git pull; then
                echo -e "${OK} ${GreenBG} Git pull 成功 ${Font}"
                compare_env_vars_only_check
                run_exec php "composer update --ignore-platform-reqs"
                run_exec php "php think migrate:run"
                echo -e "${OK} ${GreenBG} 更新完成 ${Font}"
                # echo -e "${OK} ${GreenBG} 更新外送服务 ${Font}"
                # cd takeout && make conf && make db_up.docker
                # echo -e "${OK} ${GreenBG} 更新外送服务完成 ${Font}"
                # cd ..
                $COMPOSE down
                $COMPOSE up -d --pull always
            else
                echo -e "${Error} ${RedBG} Git pull 失败，请检查网络或远程仓库状态 ${Font}"
                exit 1
            fi
        fi
        #
        run_exec php "chown -R www-data:www-data ./app"
        run_exec php "chown -R www-data:www-data ./public"
        run_exec php "chown -R www-data:www-data ./runtime"
        run_exec php "chown -R www-data:www-data ./extend"
        run_exec php "chown -R www-data:www-data ./vendor"
        run_exec php "chown -R www-data:www-data ./route"
        run_exec php "chown -R www-data:www-data ./bin"
        run_exec php "chown -R www-data:www-data think"
        run_exec php "chmod +x ./bin/license.so"
        run_exec php "chmod +x ./bin/license_arm.so"
        #
    elif [[ "$1" == "uninstall" ]]; then
        shift 1
        $COMPOSE down
        echo -e "${OK} ${GreenBG} 卸载完成 ${Font}"
    elif [[ "$1" == "mysql" ]]; then
        shift 1
        if [[ "$1" == "agent" ]] || [[ "$1" == "open" ]]; then
            run_mysql open
        elif [[ "$1" == "unagent" ]] || [[ "$1" == "close" ]]; then
            run_mysql close
        else
            e="mysql $@" && run_exec db "$e"
        fi
    elif [[ "$1" == "go" ]]; then
        shift 1
        e="./main $@" && run_exec golang "$e"
    elif [[ "$1" == "golang" ]]; then
        shift 1
        e="go $@" && run_exec golang "$e"
    elif [[ "$1" == "websocket" ]]; then
        shift 1
        e="go $@" && run_exec websocket "$e"
    elif [[ "$1" == "think" ]]; then
        shift 1
        e="php think $@" && run_exec php "$e"
    elif [[ "$1" == "composer" ]]; then
        shift 1
        e="composer $@" && run_exec php "$e"
    elif [[ "$1" == "restart" ]]; then
        shift 1
        $COMPOSE stop "$@"
        $COMPOSE start "$@"
    elif [[ "$1" == "release" ]]; then
        shift 1
            OCKER_BUILDKIT=1 docker buildx build --push -t hitosea2020/php-fpm:1.0.6 --platform linux/amd64,linux/arm64 .
    elif [[ "$1" == "repassword" ]]; then
        shift 1
        run_exec db "sh /etc/mysql/repassword.sh \"$@\""
    elif [[ "$1" == "tail-log" ]]; then
        shift 1
        # 找到最后一个目录，目录名必须是整数格式
        last_dir=$(ls -d "./admin/runtime/logs/"*/ | grep -E '/[0-9]+/$' | sort -n | tail -n 1)
        # 如果目录存在，查找最后一个以整数格式命名的.log结尾的文件
        if [ -n "$last_dir" ]; then
            last_log_file=$(ls -t "$last_dir"*.log 2>/dev/null | grep -E '/[0-9]+\.log$' | head -n 1)
            if [ -n "$last_log_file" ]; then
                echo "最后一个.log文件是: $last_log_file"
                head -n 50 "$last_log_file" && tail -f "$last_log_file"
            else
                echo "在目录 $last_dir 中没有找到 .log 文件"
            fi
        else
            echo "没有找到任何目录"
        fi
    elif [[ "$1" == "toggle_brand" ]]; then
        shift 1
        echo "$1";
    elif [[ "$1" == "build" ]]; then
        shift 1
        ./build "$@"
    else
        $COMPOSE "$@"
    fi
else
    $COMPOSE ps
fi
