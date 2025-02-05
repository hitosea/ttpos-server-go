<?php

namespace app\job\event;

use think\facade\Cache;

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

        // // 获取当前时间戳的分钟部分
        // $currentMinute = date('Y-m-d H:i');
        // // 检查 Redis 中是否存在该分钟的缓存
        // if (!Cache::get('job_scheduler_last_run:' . $currentMinute)) {
        // 月度报表库存记录
        event('ErpMonthlyStatistics');
        // 月度报表库存记录
        event('ErpMonthlyProductStatistics');
        //     // 设置 Redis 缓存，过期时间为60秒
        //     Cache::set('job_scheduler_last_run:' . $currentMinute, 60, '1');
        // }
        //
        return true;
    }
}
