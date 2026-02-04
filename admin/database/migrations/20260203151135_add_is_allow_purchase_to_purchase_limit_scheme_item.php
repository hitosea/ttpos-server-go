<?php

use think\migration\Migrator;

/**
 * 为限购方案物品配置表添加是否允许采购字段
 *
 * 在 ttpos_purchase_limit_scheme_item 表中添加：
 * - is_allow_purchase: 是否允许采购 yes/no，默认 yes
 */
class AddIsAllowPurchaseToPurchaseLimitSchemeItem extends Migrator
{
    /**
     * Change Method.
     *
     * 为限购方案物品配置表添加是否允许采购字段
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('purchase_limit_scheme_item')) {
            return;
        }

        // 获取表对象
        $table = $this->table('purchase_limit_scheme_item');

        // 添加 is_allow_purchase 字段 - 是否允许采购
        if (!$table->hasColumn('is_allow_purchase')) {
            $table->addColumn('is_allow_purchase', 'string', [
                'limit' => 10,
                'null' => false,
                'default' => 'yes',
                'comment' => '是否允许采购 yes/no',
                'after' => 'quota_limit'
            ]);
        }

        $table->update();
    }
}
