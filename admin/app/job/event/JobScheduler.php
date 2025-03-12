<?php

namespace app\job\event;

/**
 * 订单事件管理
 */
class JobScheduler
{

    /**
     * 执行函数
     */
    public function handle()
    {
        // 记录日志
        // event('RecordShopLog');

        // 月度报表库存记录
        event('ErpMonthlyStatistics');
        // 月度报表库存记录
        event('ErpMonthlyProductStatistics');
        // 设置 Redis 缓存，过期时间为60秒
        
        return true;
    }
}
