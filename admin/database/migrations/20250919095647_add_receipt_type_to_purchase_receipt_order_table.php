<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddReceiptTypeToPurchaseReceiptOrderTable extends Migrator
{
    /**
     * 添加收货类型字段到采购收货订单表
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('purchase_receipt_order')) {
            return;
        }

        $table = $this->table('purchase_receipt_order');
        
        // 检查字段是否已存在
        if (!$table->hasColumn('receipt_type')) {
            // 添加收货类型字段
            $table->addColumn('receipt_type', 'integer', ['default' => 1, 'comment' => '收货类型 1-外部收货 2-内部收货']);
            $table->update();
        }
    }
}
