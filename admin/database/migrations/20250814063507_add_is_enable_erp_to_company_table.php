<?php

use think\migration\Migrator;

class AddIsEnableErpToCompanyTable extends Migrator
{
    // 迁移目标
    const TARGET = 'all';

    /**
     * 变更方法：为 company 表新增 is_enable_erp 字段
     */
    public function change()
    {
        $table = $this->table('company');

        // 新增是否启用erp字段（若不存在）
        if (!$table->hasColumn('is_enable_erp')) {
            $table->addColumn('is_enable_erp', 'integer', ['signed' => false, 'default' => 0, 'comment' => '是否启用ERP: 0不启用, 1启用', 'after' => 'old_company_id']);
        }

        $table->update();
    }
}
