<?php
/**
 * 创建门店区域表
 *
 * 任务: #40880 门店管理功能-能给门店分区域
 * 存储总部对门店的区域分组，支持多语言名称
 */

use think\migration\Migrator;

class CreateCompanyArea extends Migrator
{
    // 不设置 TARGET: 应用到所有商户数据库
    public function change()
    {
        if (!$this->hasTable('company_area')) {
            $table = $this->table('company_area', [
                'id' => false,
                'primary_key' => 'id',
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '门店区域表',
            ]);

            $table->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'UUID'])
                ->addColumn('name', 'string', ['limit' => 1000, 'default' => '', 'comment' => '区域名称（JSON多语言）'])
                ->addColumn('multi_language_name_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '多语言名称UUID'])
                ->addColumn('sort', 'integer', ['signed' => false, 'default' => 0, 'comment' => '排序'])
                ->addColumn('create_time', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['multi_language_name_uuid'], ['name' => 'idx_multi_language_name_uuid'])
                ->create();
        }
    }
}
