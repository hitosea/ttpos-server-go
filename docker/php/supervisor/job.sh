#!/bin/bash

# Delay for 3 seconds
sleep 3

# Start the actual program
php think job start

# 重启，确保正常启动
php think job stop
php think job start