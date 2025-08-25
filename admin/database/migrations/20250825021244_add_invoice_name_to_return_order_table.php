<?php

use think\migration\Migrator;

class AddInvoiceNameToReturnOrderTable extends Migrator
{
    /**
     * 变更方法：为 product_sauce 表新增 erp_code 字段
     */
    public function change()
    {
        $table = $this->table('return_order');

        if (!$table->hasColumn('erp_invoice_name')) {
            $table->addColumn('erp_invoice_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '发票名称', 'after' => 'duty_no']);
        }

        $table->update();

        $table = $this->table('refund_order');

        if (!$table->hasColumn('erp_invoice_name')) {
            $table->addColumn('erp_invoice_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '发票名称', 'after' => 'status']);
        }

        $table->update();

        $table = $this->table('sale_order');

        if (!$table->hasColumn('erp_products_invoice_name')) {
            $table->addColumn('erp_products_invoice_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '商品发票名称', 'after' => 'cashier_name']);
        }
        if (!$table->hasColumn('erp_material_invoice_name')) {
            $table->addColumn('erp_material_invoice_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '原材料发票名称', 'after' => 'erp_products_invoice_name']);
        }

        $table->update();
    }
}
