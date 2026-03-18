<?php

use think\migration\Migrator;

/**
 * 为 ttpos_member_recharge_order 添加 erp_payment_entry_names 字段
 * 用于存储 Sales Invoice 模式下关联的 Payment Entry 名称（JSON 数组）
 * 取消 SI 时需要同时取消关联的 PE，避免 ERP 中产生孤立 Payment Entry
 */
class AddErpPaymentEntryNamesToMemberRechargeOrder extends Migrator
{
    public function up()
    {
        if (!$this->hasTable('member_recharge_order')) {
            return;
        }

        $table = $this->table('member_recharge_order');

        if (!$table->hasColumn('erp_payment_entry_names')) {
            $table->addColumn('erp_payment_entry_names', 'text', [
                'null' => true,
                'default' => null,
                'comment' => 'Payment Entry名称(JSON)',
                'after' => 'erp_products_invoice_name',
            ])->update();
        }
    }

    public function down()
    {
        if (!$this->hasTable('member_recharge_order')) {
            return;
        }

        $table = $this->table('member_recharge_order');

        if ($table->hasColumn('erp_payment_entry_names')) {
            $table->removeColumn('erp_payment_entry_names')->update();
        }
    }
}
