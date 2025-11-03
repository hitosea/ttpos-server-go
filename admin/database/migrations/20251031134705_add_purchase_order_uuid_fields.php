<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddPurchaseOrderUuidFields extends Migrator
{
    /**
     * 为采购相关表添加订单UUID字段
     */
    public function change()
    {
        // 为采购申请物品单位表添加 purchase_order_uuid 字段
        if ($this->hasTable('purchase_order_item_unit')) {
            $table = $this->table('purchase_order_item_unit');
            if (!$table->hasColumn('purchase_order_uuid')) {
                $table->addColumn('purchase_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '采购申请UUID', 'after' => 'item_uuid'])
                    ->addIndex(['purchase_order_uuid'], ['name' => 'idx_purchase_order_uuid'])
                    ->update();
            }
        }

        // 为收货单物品单位表添加 purchase_receipt_order_uuid 字段
        if ($this->hasTable('purchase_receipt_order_item_unit')) {
            $table = $this->table('purchase_receipt_order_item_unit');
            if (!$table->hasColumn('purchase_receipt_order_uuid')) {
                $table->addColumn('purchase_receipt_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '收货单UUID', 'after' => 'item_uuid'])
                    ->addIndex(['purchase_receipt_order_uuid'], ['name' => 'idx_purchase_receipt_order_uuid'])
                    ->update();
            }
        }
    }
}

