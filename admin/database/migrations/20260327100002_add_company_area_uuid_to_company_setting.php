<?php
/**
 * 为商家设置表添加门店区域UUID字段
 *
 * 任务: #40880 门店管理功能-能给门店分区域
 */

use think\migration\Migrator;

class AddCompanyAreaUuidToCompanySetting extends Migrator
{
    // 迁移目标：所有商户数据库和saas主库
    const TARGET = 'all';

    public function change()
    {
        if ($this->hasTable('company_setting')) {
            $table = $this->table('company_setting');

            if (!$table->hasColumn('company_area_uuid')) {
                $table->addColumn('company_area_uuid', 'biginteger', [
                    'signed' => false,
                    'default' => 0,
                    'null' => false,
                    'comment' => '所属区域UUID: 0-未分配',
                    'after' => 'brand_purchase_auto_approve'
                ]);
            }

            $table->update();
        }
    }
}
