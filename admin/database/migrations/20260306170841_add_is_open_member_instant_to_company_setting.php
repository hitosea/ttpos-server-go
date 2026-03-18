<?php
/**
 * 为商家设置表添加会员端即时点餐功能开关字段
 *
 * 功能: 扫码点餐到店自取
 */

use think\migration\Migrator;

class AddIsOpenMemberInstantToCompanySetting extends Migrator
{
    // 迁移目标：所有商户数据库
    const TARGET = 'all';

    /**
     * Change Method.
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('company_setting')) {
            $table = $this->table('company_setting');

            // 检查 is_open_member_instant 字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('is_open_member_instant')) {
                $table->addColumn('is_open_member_instant', 'integer', [
                    'limit' => 10,
                    'default' => 0,
                    'null' => false,
                    'comment' => '是否开启会员端即时点餐功能（扫码点餐到店自取）：0-否；1-是',
                    'after' => 'enable_kiosk'
                ]);
            }

            $table->update();
        }
    }
}
