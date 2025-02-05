<?php

namespace app\job\event;

use app\common\model\erp\ErpMonthlyStatistics as ErpMonthlyStatisticsModel;

/**
 * 同步
 */
class ErpMonthlyStatistics
{
    /**
     * 执行函数
     */
    public function handle()
    {
        try {
            (new ErpMonthlyStatisticsModel([], request()->appId ?: 0))->record();
        } catch (\Throwable $e) {
            trace('月度记录错误：' . $e->getMessage());
        }
        return true;
    }
}
