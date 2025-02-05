<?php

use think\facade\Env;
// +----------------------------------------------------------------------
// | abbitmq设置
// +----------------------------------------------------------------------

return [
    'host' => Env::get('RABBITMQ_HOST', 'rabbitmq'),
    'port' => Env::get('RABBITMQ_PORT', 5672),
    'user' => Env::get('RABBITMQ_USER', 'guest') == '${APP_ID}' ? Env::get('APP_ID', 'guest') : Env::get('RABBITMQ_USER', 'guest'),
    'password' => Env::get('RABBITMQ_PASSWORD', 'guest') == '${DB_ROOT_PASSWORD}' ? Env::get('DB_ROOT_PASSWORD', 'guest') : Env::get('RABBITMQ_PASSWORD', 'guest'),
];
