<?php
// +----------------------------------------------------------------------
// | 多语言设置
// +----------------------------------------------------------------------

use think\facade\Env;

return [
    // 默认语言
    'default_lang'    => Env::get('lang.default_lang', 'zh'),
    // 开启语言切换
    'lang_switch_on' => false,   
    // 允许的语言列表
    'allow_lang_list' => [],
    // 多语言自动侦测变量名
    'detect_var'      => 'lang',
    // 是否使用Cookie记录
    'use_cookie'      => false,
    // 多语言cookie变量
    'cookie_var'      => 'think_lang',
    // 扩展语言包
    'extend_list'     => [
        'en'    => [
            app()->getRootPath() . '/lang/en/auto.php',
        ],
        'ko'    => [
            app()->getRootPath() . 'lang/ko/auto.php',
        ],
        'th'    => [
            app()->getRootPath() . 'lang/th/auto.php',
        ],
        'tr'    => [
            app()->getRootPath() . 'lang/tr/auto.php',
        ],
        'zh'    => [
            app()->getRootPath() . 'lang/zh/auto.php',
        ],
        'zhtw'    => [
            app()->getRootPath() . 'lang/zhtw/auto.php',
        ],
        'ja'    => [
            app()->getRootPath() . 'lang/ja/auto.php',
        ],
        'my'    => [
            app()->getRootPath() . 'lang/my/auto.php',
        ]
    ],
    // Accept-Language转义为对应语言包名称
    'accept_language' => [
        'zh-hans-cn' => 'zh-cn',
    ],
    // 是否支持语言分组
    'allow_group'     => false,
];
