<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTakeoutTable extends Migrator
{
    /**
     * 创建外卖平台管理表
     * - takeout: 外卖平台状态管理表
     */
    public function change()
    {
        // 创建外卖平台管理表 takeout
        if (!$this->hasTable('takeout')) {
            $table = $this->table('takeout', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '外卖平台状态管理表',
                'id' => false,
                'primary_key' => ['id']
            ]);

            $table
                // 基础字段
                ->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识'])

                // 业务字段
                ->addColumn('platform', 'string', ['limit' => 50, 'default' => '', 'comment' => '外卖平台(grab/lineman等)'])
                ->addColumn('enabled', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'comment' => '是否开启(1:开启 0:关闭)'])
                ->addColumn('menu', 'json', ['null' => true, 'comment' => '平台菜单数据(JSON格式)'])
                ->addColumn('is_bound', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'comment' => '是否已经绑定平台(1:已绑定 0:未绑定)'])

                // 时间字段
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])

                // 索引
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['platform', 'delete_time'], ['unique' => true, 'name' => 'uk_platform'])
                ->addIndex(['platform'], ['name' => 'idx_platform'])
                ->addIndex(['enabled'], ['name' => 'idx_enabled'])
                ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])

                ->create();
        }
    }
}
