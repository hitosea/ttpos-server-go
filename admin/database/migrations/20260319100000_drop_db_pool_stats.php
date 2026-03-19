<?php

use think\migration\Migrator;

/**
 * 清空 ttpos_db_pool_stats 表数据（连接池监控采集已关闭）
 */
class DropDbPoolStats extends Migrator
{
    const TARGET = 'main';

    public function up()
    {
        if ($this->hasTable('db_pool_stats')) {
            $this->execute("TRUNCATE TABLE `ttpos_db_pool_stats`");
        }
    }

    public function down()
    {
        // 数据清空不可逆
    }
}
