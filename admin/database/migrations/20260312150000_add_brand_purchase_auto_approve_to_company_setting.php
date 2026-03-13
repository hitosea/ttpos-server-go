<?php
/**
 * 为商家设置表添加品牌采购自动审批开关字段
 *
 * 任务: #40322 #40317 品牌采购自动审核开关
 */

use think\migration\Migrator;

class AddBrandPurchaseAutoApproveToCompanySetting extends Migrator
{
    // 迁移目标：所有商户数据库和saas主库
    const TARGET = 'all';

    public function change()
    {
        if ($this->hasTable('company_setting')) {
            $table = $this->table('company_setting');

            if (!$table->hasColumn('brand_purchase_auto_approve')) {
                $table->addColumn('brand_purchase_auto_approve', 'integer', [
                    'limit' => 3,
                    'default' => 0,
                    'null' => false,
                    'comment' => '品牌采购自动审批: 0-关闭 1-开启',
                    'after' => 'enable_lineman_delivery'
                ]);
            }

            $table->update();
        }
    }
}
