<?php

use think\migration\Migrator;

/**
 * 修改自动收货规则表 name 字段类型由 VARCHAR(1000) 调整为 TEXT
 *
 * 目标表：ttpos_auto_receipt_rule（迁移中使用不带 ttpos_ 前缀的表名）
 * 变更内容：
 * - name: VARCHAR(1000) -> TEXT
 */
class AlterAutoReceiptRuleNameToText extends Migrator
{
    const TARGET = 'main';

    /**
     * Change Method.
     *
     * 修改 ttpos_auto_receipt_rule.name 字段为 TEXT
     */
    public function change()
    {
        if (!$this->hasTable('auto_receipt_rule')) {
            return;
        }

        $table = $this->table('auto_receipt_rule');

        if ($table->hasColumn('name')) {
            $table->changeColumn('name', 'text', [
                'null' => true,
                'comment' => '规则名称（多语言JSON）',
            ])->update();
        }
    }
}
