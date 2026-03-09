<?php
/**
 * 为出库单表添加索引，优化慢查询
 *
 * 慢查询：SELECT * FROM ttpos_warehouse_out_form WHERE associated_order_uuid = ? AND delete_time = 0
 * 原因：124万行全表扫描，无 associated_order_uuid 索引
 *
 * 为出库单明细表添加 takeout_order_uuid 索引
 * 慢查询：SELECT * FROM ttpos_warehouse_out_form_item WHERE takeout_order_uuid = ? AND revoke...
 * 原因：48万行全表扫描，无 takeout_order_uuid 索引
 */

use think\migration\Migrator;

class AddIndexWarehouseOutForm extends Migrator
{
    public function change()
    {
        // 出库单表：添加 associated_order_uuid 索引
        if ($this->hasTable('warehouse_out_form')) {
            $table = $this->table('warehouse_out_form');

            if (!$table->hasIndex('associated_order_uuid')) {
                $table->addIndex('associated_order_uuid', ['name' => 'idx_associated_order_uuid']);
            }

            $table->update();
        }

        // 出库单明细表：添加 takeout_order_uuid 索引
        if ($this->hasTable('warehouse_out_form_item')) {
            $table = $this->table('warehouse_out_form_item');

            if (!$table->hasIndex('takeout_order_uuid')) {
                $table->addIndex('takeout_order_uuid', ['name' => 'idx_takeout_order_uuid']);
            }

            $table->update();
        }
    }
}
