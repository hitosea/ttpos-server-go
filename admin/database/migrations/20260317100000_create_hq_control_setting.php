<?php
/**
 * 创建总部控制设置 KV 表
 *
 * 存储总部对子店商品/物品的控制模式（统一控制/分开控制）
 * 仅总部可写，子店通过查询总部数据库读取
 * 无记录时代码层返回默认值（负库存=统一控制，其他=分开控制）
 */

use think\migration\Migrator;

class CreateHqControlSetting extends Migrator
{
    // 不设置 TARGET: 应用到所有商户数据库
    public function change()
    {
        if (!$this->hasTable('hq_control_setting')) {
            $table = $this->table('hq_control_setting', [
                'id' => false,
                'primary_key' => 'id',
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '总部控制设置（KV表）',
            ]);

            $table->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'UUID'])
                ->addColumn('field_type', 'string', ['limit' => 50, 'null' => false, 'comment' => '字段类型: dine_shelf/takeout_shelf/takeout_price/negative_stock'])
                ->addColumn('control_mode', 'integer', ['limit' => 3, 'default' => 0, 'null' => false, 'comment' => '控制模式: 0-分开控制 1-统一控制'])
                ->addColumn('create_time', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
                ->addIndex(['field_type'], ['unique' => true, 'name' => 'uk_field_type'])
                ->create();
        }
    }
}
