<?php
// 事件定义文件
return [
    'listen' => [
        'AppInit' => [],
        'HttpRun' => [],
        'HttpEnd' => [],
        'LogLevel' => [],
        'LogWrite' => [],
        'CashierPaySuccess' => [
            \app\cashier\event\PaySuccess::class
        ],
        /*用户等级*/
        'UserGrade' => [
            \app\job\event\UserGrade::class
        ],
        /*任务调度*/
        'JobScheduler' => [
            \app\job\event\JobScheduler::class
        ],
        /*订单事件*/
        'Order' => [
            \app\job\event\Order::class
        ],
        /*交班打印*/
        'ShopPrint' => [
            \app\job\event\ShopPrint::class
        ],
        /*收集请求日志*/
        'RecordShopLog' => [
            \app\job\event\RecordShopLog::class
        ],
        /*月度库存统计*/
        'ErpMonthlyStatistics' => [
            \app\job\event\ErpMonthlyStatistics::class
        ],
        /*月度商品库存统计*/
        'ErpMonthlyProductStatistics' => [
            \app\job\event\ErpMonthlyProductStatistics::class
        ],
        /*同步*/
        'Synchronization' => [
            \app\job\event\Sync::class
        ],
    ],

    'subscribe' => [],
];
