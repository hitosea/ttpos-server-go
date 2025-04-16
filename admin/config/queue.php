<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK IT ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006-2016 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: yunwuxin <448901948@qq.com>
// +----------------------------------------------------------------------

return [
    // 默认驱动
    'default'     => 'redis',
    // 队列连接方式配置
    'connections' => [
        // Redis 驱动
        'redis'   => [
            // 驱动类型
            'type'       => 'redis',
            // 队列名称（支持商户隔离）
            'queue'      => 'default', 
            // 连接配置
            'host'       => 'golang',
            'port'       => 6739,
            'password'   => '',
            'select'     => 0, 
            'timeout'    => 0,
            'persistent' => false,
        ],
    ],
    'failed'      => [
        'type'  => 'none',
        'table' => 'failed_jobs',
    ]
];
