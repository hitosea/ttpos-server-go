<?php

use think\migration\Migrator;

/**
 * 添加 ERP 发票模式字段到 company_setting 表
 * 0 = POS Invoice 模式（默认，兼容旧流程）
 * 1 = Sales Invoice 模式（新流程）
 */
class AddErpInvoiceModeToCompanySetting extends Migrator
{
    const TARGET = 'all';

    public function change()
    {
        if (!$this->hasTable('company_setting')) {
            return;
        }

        $table = $this->table('company_setting');

        if (!$table->hasColumn('erp_invoice_mode')) {
            $table->addColumn('erp_invoice_mode', 'integer', [
                'limit' => 11,
                'null' => false,
                'default' => 1,
                'comment' => 'ERP发票模式: 0=POS Invoice 1=Sales Invoice',
                'after' => 'enable_lineman_delivery',
            ])->update();
        }
    }
}
