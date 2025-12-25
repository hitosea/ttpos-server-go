<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddOrderTypeToWarehouseOutFormTable extends Migrator
{
    /**
     * 添加订单类型字段到 ttpos_warehouse_out_form 表
     */
    public function change()
    {
        $table = $this->table('warehouse_out_form');

        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('order_type')) {
            $table->addColumn('order_type', 'integer', ['limit' => 1, 'null' => false, 'default' => 0, 'comment' => '订单类型,0-堂食订单 1-外卖订单', 'after' => 'scene'])->update();
        }
    }
}

