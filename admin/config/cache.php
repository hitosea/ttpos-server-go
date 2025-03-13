<?php
use think\facade\Env;

// +----------------------------------------------------------------------
// | 缓存设置
// +----------------------------------------------------------------------
// 根目录
$rootPath = dirname(__DIR__);

return [
    // 默认缓存驱动
    'default' => Env::get('CACHE_DRIVER', 'redis'),

    // 缓存连接方式配置
    'stores'  => [
        'file' => [
            // 驱动方式
            'type'       => 'File',
            // 缓存保存目录
            'path'       => "{$rootPath}/runtime/cache/",
            // 缓存前缀
            'prefix'     => '',
            // 缓存有效期 0表示永久缓存
            'expire'     => 0,
            // 缓存标签前缀
            'tag_prefix' => 'tag:',
            // 序列化机制 例如 ['serialize', 'unserialize']
            'serialize'  => [],
        ],
        'redis' => [
            'type'       => 'redis',
            'host'       => Env::get('REDIS_HOST', '127.0.0.1'),
            'port'       => Env::get('REDIS_PORT', 6379),
            'password'   => Env::get('REDIS_PASSWORD', ''),
            'select'     => 0,
            'timeout'    => 0,
            'persistent' => false,
            'serialize' => ['json_encode', function($val) { return json_decode($val, true); }],
        ],
        'redis-write' => [
            'type'       => 'redis',
            'host'       => Env::get('REDIS_HOST_WRITE', Env::get('REDIS_HOST', '127.0.0.1')),
            'port'       => Env::get('REDIS_PORT_WRITE', Env::get('REDIS_PORT', 6379)),
            'password'   => Env::get('REDIS_PASSWORD_WRITE', Env::get('REDIS_PASSWORD', '')),
            'select'     => 0,
            'timeout'    => 0,
            'persistent' => false,
            'serialize' => ['json_encode', function($val) { return json_decode($val, true); }],
        ],
        // 更多的缓存连接
    ],
];
