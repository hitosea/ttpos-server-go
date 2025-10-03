<?php

use think\migration\Migrator;
use think\migration\db\Column;

class ModifyWarehouseItemDecimalFields extends Migrator
{
    /**
     * 修改 warehouse_item 表的 decimal 字段类型为 decimal(20,8)
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('warehouse_item')) {
            $table = $this->table('warehouse_item');
            
            // 检查字段是否存在并修改
            if ($table->hasColumn('stock')) {
                $table->changeColumn('stock', 'decimal', ['precision' => 22, 'scale' => 8, 'default' => 0, 'comment' => '库存数量']);
            }
            
            if ($table->hasColumn('reserved_stock')) {
                $table->changeColumn('reserved_stock', 'decimal', ['precision' => 22, 'scale' => 8, 'default' => 0, 'comment' => '预留库存数量']);
            }
            
            $table->save();
        }
    }
}
