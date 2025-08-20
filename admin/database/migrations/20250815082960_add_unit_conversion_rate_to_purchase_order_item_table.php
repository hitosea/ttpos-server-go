<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddUnitConversionRateToPurchaseOrderItemTable extends Migrator
{
    /**
     * 迁移
     */
    public function change()
    {
        $this->addUnitConversionRateToPurchaseOrderItem();
    }

    /**
     * 为采购申请物品表添加单位转换率字段
     */
    private function addUnitConversionRateToPurchaseOrderItem()
    {
        // 检查表是否存在
        if ($this->hasTable('purchase_order_item')) {
            $table = $this->table('purchase_order_item');
            
            // 检查字段是否已存在，避免重复添加
            if (!$table->hasColumn('unit_conversion_rate')) {
                $table->addColumn('unit_conversion_rate', 'decimal', [
                    'precision' => 12,
                    'scale' => 4,
                    'default' => 1.0000,
                    'comment' => '单位转换率。申请数量*转换率=基准单位申请数量',
                    'after' => 'unit_name'
                ]);
                $table->update();
            }
        }
    }
}
