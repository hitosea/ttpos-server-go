<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddPurchaseTimeToPurchaseReceiptOrderTable extends Migrator
{
    /**
     * 添加采购时间字段到采购收货订单表
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('purchase_receipt_order')) {
            return;
        }

        $table = $this->table('purchase_receipt_order');
        
        // 检查字段是否已存在
        if (!$table->hasColumn('purchase_time')) {
            // 添加采购时间字段
            $table->addColumn('purchase_time', 'integer', ['default' => 0, 'comment' => '采购时间']);
            $table->update();
        }
    }
}
