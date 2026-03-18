<?php

use think\migration\Migrator;

/**
 * 品采自动收货：为收货单增加 is_auto_receipt 标识字段
 */
class AddIsAutoReceiptToPurchaseReceiptOrder extends Migrator
{
    public function change()
    {
        if (!$this->hasTable('purchase_receipt_order')) {
            return;
        }

        $table = $this->table('purchase_receipt_order');

        if (!$table->hasColumn('is_auto_receipt')) {
            $table->addColumn('is_auto_receipt', 'integer', [
                'limit' => 1,
                'null' => false,
                'default' => 0,
                'comment' => '是否自动收货:0-否 1-是',
                'after' => 'is_from_delivery_note',
            ])->update();
        }
    }
}
