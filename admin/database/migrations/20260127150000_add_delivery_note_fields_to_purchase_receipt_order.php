<?php

use think\migration\Migrator;

/**
 * 为收货单表添加 Delivery Note 相关字段
 *
 * 在 ttpos_purchase_receipt_order 表中添加：
 * - delivery_note_no: ERP Delivery Note号
 * - is_from_delivery_note: 是否来自DN单
 */
class AddDeliveryNoteFieldsToPurchaseReceiptOrder extends Migrator
{
    /**
     * Change Method.
     *
     * 为收货单表添加 Delivery Note 相关字段
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('purchase_receipt_order')) {
            return;
        }

        // 获取表对象
        $table = $this->table('purchase_receipt_order');

        // 添加 delivery_note_no 字段 - ERP Delivery Note号
        if (!$table->hasColumn('delivery_note_no')) {
            $table->addColumn('delivery_note_no', 'string', [
                'limit' => 255,
                'null' => false,
                'default' => '',
                'comment' => 'ERP Delivery Note号',
                'after' => 'erp_order_no'
            ]);
        }

        // 添加 is_from_delivery_note 字段 - 是否来自DN单
        if (!$table->hasColumn('is_from_delivery_note')) {
            $table->addColumn('is_from_delivery_note', 'integer', [
                'limit' => 10,
                'null' => false,
                'default' => 0,
                'comment' => '是否来自DN单:0-否 1-是',
                'after' => 'delivery_note_no'
            ]);
        }

        // 添加索引
        if (!$table->hasIndex('delivery_note_no')) {
            $table->addIndex(['delivery_note_no'], [
                'name' => 'idx_delivery_note_no'
            ]);
        }

        $table->update();
    }
}
