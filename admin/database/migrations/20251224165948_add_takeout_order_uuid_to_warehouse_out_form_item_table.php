<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddTakeoutOrderUuidToWarehouseOutFormItemTable extends Migrator
{
    /**
     * 添加外卖订单uuid字段到 ttpos_warehouse_out_form_item 表
     */
    public function change()
    {
        $table = $this->table('warehouse_out_form_item');

        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('takeout_order_uuid')) {
            $table->addColumn('takeout_order_uuid', 'biginteger', ['limit' => 20, 'null' => false, 'default' => 0, 'comment' => '外卖订单uuid,用于记录外卖订单的出库记录', 'after' => 'staff_shift_log_uuid'])->update();
        }
    }
}

