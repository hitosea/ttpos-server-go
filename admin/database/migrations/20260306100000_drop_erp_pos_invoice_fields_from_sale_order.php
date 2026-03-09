<?php

use think\migration\Migrator;

/**
 * 删除 ttpos_sale_order 表中已废弃的 POS Invoice 字段：
 * - erp_products_invoice_name: 商品发票名称（已迁移到 SI 模式）
 * - erp_material_invoice_name: 原材料发票名称（已迁移到 SI 模式）
 */
class DropErpPosInvoiceFieldsFromSaleOrder extends Migrator
{
    public function up()
    {
        if (!$this->hasTable('sale_order')) {
            return;
        }

        $table = $this->table('sale_order');

        if ($table->hasColumn('erp_products_invoice_name')) {
            $table->removeColumn('erp_products_invoice_name')->update();
        }

        if ($table->hasColumn('erp_material_invoice_name')) {
            $table->removeColumn('erp_material_invoice_name')->update();
        }
    }

    public function down()
    {
        if (!$this->hasTable('sale_order')) {
            return;
        }

        $table = $this->table('sale_order');

        if (!$table->hasColumn('erp_products_invoice_name')) {
            $table->addColumn('erp_products_invoice_name', 'string', [
                'limit' => 255,
                'null' => false,
                'default' => '',
                'comment' => '商品发票名称',
                'after' => 'cashier_name',
            ])->update();
        }

        if (!$table->hasColumn('erp_material_invoice_name')) {
            $table->addColumn('erp_material_invoice_name', 'string', [
                'limit' => 255,
                'null' => false,
                'default' => '',
                'comment' => '原材料发票名称',
                'after' => 'erp_products_invoice_name',
            ])->update();
        }
    }
}
