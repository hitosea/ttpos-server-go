#!/bin/bash

# Delay for 3 seconds
sleep 3

PID_DIR="/var/www/runtime/workerman"
mkdir -p $PID_DIR
chown -R www-data:www-data $PID_DIR

# Start the actual program
php think job start

# 重启，确保正常启动
php think job stop
php think job start