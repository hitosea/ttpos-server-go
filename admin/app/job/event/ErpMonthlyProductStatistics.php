<?php

namespace app\job\event;

use app\common\model\erp\ErpMonthlyProductStatistics as ErpMonthlyProductStatisticsModel;

/**
 * 同步
 */
class ErpMonthlyProductStatistics
{
    /**
     * 执行函数
     */
    public function handle()
    {
        try {
            (new ErpMonthlyProductStatisticsModel([], request()->appId ?: 0))->recordStart();
            (new ErpMonthlyProductStatisticsModel([], request()->appId ?: 0))->recordEnd();
        } catch (\Throwable $e) {
            trace('月度商品记录错误：' . $e->getMessage());
        }
        return true;
    }
}
