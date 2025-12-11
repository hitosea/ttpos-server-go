<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddInvoiceNumberToSaleOrderInvoiceInfo extends Migrator
{
    /**
     * 为销售订单发票表新增发票编号字段，按日递增生成
     * Requirement: add-invoice-number-for-printing
     */
    public function change()
    {
        if ($this->hasTable('sale_order_invoice_info')) {
            $table = $this->table('sale_order_invoice_info');

            if (!$table->hasColumn('invoice_number')) {
                $table->addColumn('invoice_number', 'string', ['limit' => 64, 'default' => '', 'comment' => '发票编号', 'after' => 'sale_order_uuid']);
            }

            $table->update();
        }
    }
}


