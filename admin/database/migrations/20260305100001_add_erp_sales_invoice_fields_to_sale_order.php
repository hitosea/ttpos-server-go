<?php

use think\migration\Migrator;
use think\migration\db\Column;

/**
 * ERP Reform Spec 2: Sales Invoice Pipeline
 *
 * 为 ttpos_sale_order 表添加 Sales Invoice 相关字段：
 * - erp_sales_invoice_name: SI 单据名称
 * - erp_payment_entry_names: PE 单据名称列表(JSON)
 * - erp_sync_status: ERP 同步状态
 * - erp_reverse_count: 反结账次数
 * - erp_stock_deducted: 库存是否已通过 Stock Entry 扣减
 */
class AddErpSalesInvoiceFieldsToSaleOrder extends Migrator
{
    /**
     * Change Method.
     */
    public function change()
    {
        if (!$this->hasTable('sale_order')) {
            return;
        }

        $table = $this->table('sale_order');

        if (!$table->hasColumn('erp_sales_invoice_name')) {
            $table->addColumn('erp_sales_invoice_name', 'string', [
                'limit' => 255,
                'null' => false,
                'default' => '',
                'comment' => 'ERP Sales Invoice名称',
                'after' => 'erp_discount_amount',
            ])->update();
        }

        if (!$table->hasColumn('erp_payment_entry_names')) {
            $table->addColumn('erp_payment_entry_names', 'text', [
                'null' => true,
                'comment' => 'ERP Payment Entry名称(JSON)',
                'after' => 'erp_sales_invoice_name',
            ])->update();
        }

        if (!$table->hasColumn('erp_sync_status')) {
            $table->addColumn('erp_sync_status', 'integer', [
                'limit' => \Phinx\Db\Adapter\MysqlAdapter::INT_TINY,
                'null' => false,
                'default' => 0,
                'comment' => 'ERP同步状态: 0=未同步 1=已入队 2=进行中 3=成功 4=失败',
                'after' => 'erp_payment_entry_names',
            ])->update();
        }

        if (!$table->hasColumn('erp_reverse_count')) {
            $table->addColumn('erp_reverse_count', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '反结账次数',
                'after' => 'erp_sync_status',
            ])->update();
        }

        if (!$table->hasColumn('erp_stock_deducted')) {
            $table->addColumn('erp_stock_deducted', 'integer', [
                'limit' => \Phinx\Db\Adapter\MysqlAdapter::INT_TINY,
                'null' => false,
                'default' => 0,
                'comment' => '库存是否已通过StockEntry扣减',
                'after' => 'erp_reverse_count',
            ])->update();
        }
    }
}
