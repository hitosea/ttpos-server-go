FROM php:8.2.7-fpm

ENV php_conf /usr/local/etc/php-fpm.conf
ENV fpm_conf /usr/local/etc/php-fpm.d/www.conf
ENV php_vars /usr/local/etc/php/conf.d/docker-vars.ini

# 设置工作目录
WORKDIR /var/www

# 设置时区
# ENV TZ=Asia/Shanghai
# RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 更新安装依赖包和PHP核心拓展
RUN apt-get update && apt-get install -y --no-install-recommends libfreetype6-dev libjpeg62-turbo-dev libpng-dev libgmp-dev curl supervisor cron vim \
        && apt-get install -y sshpass git \
        && apt-get install -y zip unzip \
        && apt-get install -y net-tools \
        && apt-get install -y procps bash \
        && rm -r /var/lib/apt/lists/* \
        && docker-php-ext-configure gd --enable-gd --with-freetype --with-jpeg \
        && docker-php-ext-install -j$(nproc) gd opcache pdo_mysql gettext sockets bcmath pcntl gmp mysqli \
        && pecl install -o -f redis \ 
        && echo "extension=redis.so" > /usr/local/etc/php/conf.d/redis.ini

RUN apt-get update && apt-get install -y libmagickwand-dev \
        && docker-php-ext-install -j$(nproc) intl \
        && pecl install imagick \ 
        && echo "extension=imagick.so" > /usr/local/etc/php/conf.d/imagick.ini

RUN apt-get update && apt-get install -y libzip-dev \
        && pecl install zip \ 
        && echo "extension=zip.so" > /usr/local/etc/php/conf.d/zip.ini

RUN apt-get update && apt-get install -y librabbitmq-dev \
        && pecl install amqp \ 
        && echo "extension=amqp.so" > /usr/local/etc/php/conf.d/amqp.ini

RUN docker-php-ext-install -j$(nproc) ffi

RUN apt-get update && apt-get install -y aapt

RUN apt-get update && apt-get install -y inotify-tools

# 配置 crontab
RUN mkdir -p /var/spool/cron/crontabs
RUN echo '* * * * * bash /var/www/docker/crontab/crontab.sh &> /dev/null' >> /var/spool/cron/crontabs/cron
RUN crontab /var/spool/cron/crontabs/cron 

# tweak php-fpm config
RUN echo "cgi.fix_pathinfo=0" > ${php_vars} &&\
    echo "upload_max_filesize = 100M"  >> ${php_vars} &&\
    echo "post_max_size = 100M"  >> ${php_vars} &&\
    echo "variables_order = \"EGPCS\""  >> ${php_vars} && \
    echo "memory_limit = 128M"  >> ${php_vars} && \
    sed -i \
        -e "s/;catch_workers_output\s*=\s*yes/catch_workers_output = yes/g" \
        -e "s/pm.max_children = 5/pm.max_children = 4/g" \
        -e "s/pm.start_servers = 2/pm.start_servers = 3/g" \
        -e "s/pm.min_spare_servers = 1/pm.min_spare_servers = 2/g" \
        -e "s/pm.max_spare_servers = 3/pm.max_spare_servers = 4/g" \
        -e "s/;pm.max_requests = 500/pm.max_requests = 200/g" \
        -e "s/;listen.mode = 0660/listen.mode = 0666/g" \
        -e "s/^;clear_env = no$/clear_env = no/" \
        ${fpm_conf}

# 配置 php.ini
RUN cp /usr/local/etc/php/php.ini-development /usr/local/etc/php/php.ini && \
	sed -i \
	    -e "s/;opcache/opcache/g" \
	    -e "s/;zend_extension=opcache/zend_extension=opcache/g" \
	    -e "s/;extension=zip/extension=zip/g" \
        /usr/local/etc/php/php.ini

# 安装 Composer
ENV COMPOSER_HOME /root/composer
RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer
ENV PATH $COMPOSER_HOME/vendor/bin:$PATH

# 启动 supervisord 和 php-fpm
CMD supervisord -c /etc/supervisor/supervisord.conf && php-fpm