<?php
/**
 * 为出库单明细表添加索引，优化慢查询
 *
 * 慢查询1：WHERE warehouse_uuid=? AND scene=0 AND staff_shift_log_uuid IN (?) ...
 * 慢查询2：WHERE sale_order_uuid=? AND revoke_time=0
 * 原因：40-47万行全表扫描，无可用索引
 */

use think\migration\Migrator;

class AddIndexWarehouseOutFormItem extends Migrator
{
    public function change()
    {
        if ($this->hasTable('warehouse_out_form_item')) {
            $table = $this->table('warehouse_out_form_item');

            if (!$table->hasIndex('staff_shift_log_uuid')) {
                $table->addIndex('staff_shift_log_uuid', ['name' => 'idx_staff_shift_log_uuid']);
            }

            if (!$table->hasIndex('sale_order_uuid')) {
                $table->addIndex('sale_order_uuid', ['name' => 'idx_sale_order_uuid']);
            }

            $table->update();
        }
    }
}
