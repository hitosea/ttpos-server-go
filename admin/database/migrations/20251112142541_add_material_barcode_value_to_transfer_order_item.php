<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddMaterialBarcodeValueToTransferOrderItem extends Migrator
{
    /**
     * 在调拨单明细表中添加物品条码值字段
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('transfer_order_item')) {
            $table = $this->table('transfer_order_item');
            
            // 检查字段是否已存在，不存在则添加
            if (!$table->hasColumn('material_barcode_value')) {
                $table->addColumn('material_barcode_value', 'string', ['limit' => 255, 'default' => '', 'comment' => '物品条码值', 'after' => 'material_internal_code'])
                    ->update();
            }
        }
    }
}

