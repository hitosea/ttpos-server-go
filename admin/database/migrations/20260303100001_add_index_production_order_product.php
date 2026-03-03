<?php
/**
 * 为生产订单商品表添加复合索引，优化 KDS 厨显查询
 *
 * 慢查询：SELECT * FROM ttpos_production_order_product WHERE status=2 AND delete_time=0 ... ORDER BY finished_time desc LIMIT 3
 * 原因：status=2 占 99.5% 行，单列 idx_status 无效 + filesort 4.2万行
 * 优化：复合索引 (status, delete_time, finished_time) 消除 filesort，配合 LIMIT 3 只扫前几行
 */

use think\migration\Migrator;

class AddIndexProductionOrderProduct extends Migrator
{
    public function change()
    {
        if ($this->hasTable('production_order_product')) {
            $table = $this->table('production_order_product');

            if (!$table->hasIndex(['status', 'delete_time', 'finished_time'])) {
                $table->addIndex(['status', 'delete_time', 'finished_time'], ['name' => 'idx_status_delete_finished']);
            }

            $table->update();
        }
    }
}
