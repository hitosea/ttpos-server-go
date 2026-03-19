<?php

use think\migration\Migrator;

/**
 * 清理 ttpos_order_operation_duration 表历史数据，仅保留最近7天
 */
class CleanOrderOperationDurationData extends Migrator
{
    const TARGET = 'main';

    public function up()
    {
        if (!$this->hasTable('order_operation_duration')) {
            return;
        }

        $sevenDaysAgo = time() - 7 * 86400;
        $batchSize = 10000;

        // 分批删除，避免锁表
        do {
            $this->execute("DELETE FROM `ttpos_order_operation_duration` WHERE `create_time` > 0 AND `create_time` < {$sevenDaysAgo} LIMIT {$batchSize}");
            $result = $this->fetchRow("SELECT ROW_COUNT() AS affected");
            $affected = $result['affected'] ?? 0;
        } while ($affected > 0);
    }

    public function down()
    {
        // 数据删除不可逆
    }
}
