<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSupplierFieldsToPurchaseReceiptOrderTable extends Migrator
{
   
    public function change()
    {
        $table = $this->table('purchase_receipt_order');

        // 检查字段是否已存在
        if (!$table->hasColumn('supplier_name')) {
            $table->addColumn('supplier_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '供应商名称', 'after' => 'purchase_order_no'])
                  ->update();
        }

        if (!$table->hasColumn('supplier_erp_code')) {
            $table->addColumn('supplier_erp_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '供应商ERP编码', 'after' => 'supplier_name'])
                  ->update();
        }
    }
}
