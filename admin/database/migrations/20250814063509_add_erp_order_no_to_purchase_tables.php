<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpOrderNoToPurchaseTables extends Migrator
{
    /**
     * 迁移
     */
    public function change()
    {
        // 为采购申请表添加ERP单号字段
        $this->addErpOrderNoToPurchaseOrder();
        
        // 为收货单表添加ERP单号字段
        $this->addErpOrderNoToPurchaseReceiptOrder();
    }

    /**
     * 为采购申请表添加ERP单号字段
     */
    private function addErpOrderNoToPurchaseOrder()
    {
        $table = $this->table('purchase_order');
        
        // 检查字段是否已存在，避免重复添加
        if (!$table->hasColumn('erp_order_no')) {
            $table->addColumn('erp_order_no', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERP采购单号', 'after' => 'order_no']);
        }
        
        $table->save();
    }

    /**
     * 为收货单表添加ERP单号字段
     */
    private function addErpOrderNoToPurchaseReceiptOrder()
    {
        $table = $this->table('purchase_receipt_order');
        
        // 检查字段是否已存在，避免重复添加
        if (!$table->hasColumn('erp_order_no')) {
            $table->addColumn('erp_order_no', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERP收货单号', 'after' => 'order_no']);
        }
        
        $table->save();
    }
}
