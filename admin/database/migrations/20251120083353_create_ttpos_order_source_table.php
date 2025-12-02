<?php
use think\migration\Migrator;
use think\migration\db\Column;

class CreateTtposOrderSourceTable extends Migrator
{
    /**
     * 创建外卖来源配置表
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('order_source')) {
            $table = $this->table('order_source', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '外卖来源配置表'
            ]);

            $table->addColumn('id', 'biginteger', ['identity' => true, 'signed' => false, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'null' => false, 'comment' => '唯一标识'])
                ->addColumn('multi_language_name_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'null' => false, 'comment' => '多语言名称UUID'])
                ->addColumn('sort', 'integer', ['default' => 0, 'null' => false, 'comment' => '排序'])
                ->addColumn('status', 'integer', ['limit' => 3, 'default' => 1, 'null' => false, 'comment' => '状态：1-启用；0-禁用'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'null' => false, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'null' => false, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'null' => false, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['multi_language_name_uuid'], ['name' => 'idx_multi_language_name_uuid'])
                ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])
                ->addIndex(['status'], ['name' => 'idx_status'])
                ->create();
        }
    }
}

