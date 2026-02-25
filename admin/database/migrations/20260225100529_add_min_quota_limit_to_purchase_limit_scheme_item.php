<?php

use think\migration\Migrator;
use think\migration\db\Column;

/**
 * 任务 #39636: 限购方案增加最小采购数量
 *
 * 为 ttpos_purchase_limit_scheme_item 表添加 min_quota_limit 字段
 * 用于设置物品的最小采购数量限制
 */
class AddMinQuotaLimitToPurchaseLimitSchemeItem extends Migrator
{
    /**
     * Change Method.
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('purchase_limit_scheme_item')) {
            return;
        }

        $table = $this->table('purchase_limit_scheme_item');

        // 添加 min_quota_limit 字段 - 最小采购数量
        if (!$table->hasColumn('min_quota_limit')) {
            $table->addColumn('min_quota_limit', 'decimal', [
                'precision' => 20,
                'scale' => 8,
                'null' => false,
                'default' => '0',
                'comment' => '最小采购数量（0=不限制）',
                'after' => 'quota_limit',
            ])->update();
        }
    }
}
