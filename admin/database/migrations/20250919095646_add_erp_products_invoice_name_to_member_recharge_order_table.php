<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpProductsInvoiceNameToMemberRechargeOrderTable extends Migrator
{
    /**
     * 添加 erp_products_invoice_name 字段到 ttpos_member_recharge_order 表
     */
    public function change()
    {
        $table = $this->table('member_recharge_order');

        // 检查字段是否已存在
        if (!$table->hasColumn('erp_products_invoice_name')) {
            $table->addColumn('erp_products_invoice_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '商品发票名称', 'after' => 'balance_recharged'])
                  ->update();
        }
    }
}
