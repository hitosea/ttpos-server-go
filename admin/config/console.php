<?php
// +----------------------------------------------------------------------
// | 控制台配置
// +----------------------------------------------------------------------
return [
    // 指令定义
    'commands' => [
        // 数据库迁移
        'migrate:run' => app\common\command\migration\MigrateRun::class,
        'migrate:rollback' => app\common\command\migration\MigrateRollback::class,
        // 定时任务
        'job' => \app\job\command\Job::class,
        // 多语言
        'lang' => app\common\command\Lang::class,
        // 
        'clear-cache' => app\common\command\ClearCache::class,
        //
        'reload-order' => app\common\command\ReloadOrder::class,
        // 
        'get-mac-addr' => app\common\command\GetMacAddr::class,
        //
        'renewal:info' => app\common\command\RenewalInfo::class,
        //
        'test' => app\common\command\Test::class,
        // 
        'tmpimport' => app\common\command\TmpImport::class,
        // 
        'upload-file-to-google' => app\common\command\UploadFileToGoogle::class,
        // 
        'fix-order-pay-type' => app\common\command\FixOrderPayType::class,
        // 
        'redis:subscribe' => app\job\command\RedisSubscribe::class,
    ],
];
