<?php

use think\migration\Migrator;
use think\migration\worker\Incr;
use Phinx\Db\Adapter\MysqlAdapter;

class CreateOrderItemRemarkTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-change-method
     *
     * @return void
     */
    public function change()
    {
        // 检查表是否已经存在
        if ($this->hasTable('order_item_remark')) {
            return;
        }

        $table = $this->table('order_item_remark', [
            'comment' => '单品备注原因表',
            'engine' => 'InnoDB',
            'encoding' => 'utf8mb4',
            'collation' => 'utf8mb4_unicode_ci',
            'id' => false,
            'primary_key' => ['id']
        ]);

        // 定义表字段
        $table->addColumn('id', 'integer', ['identity' => true, 'comment' => '主键ID', 'signed' => false]);
        $table->addColumn('uuid', 'biginteger', ['comment' => '唯一标识', 'default' => 0, 'signed' => false]);
        $table->addColumn('name', 'string', ['limit' => 255, 'comment' => '名称', 'default' => '']);
        $table->addColumn('multi_language_name_uuid', 'biginteger', ['comment' => '多语言名称ID', 'default' => 0, 'signed' => false]);
        $table->addColumn('create_time', 'integer', ['comment' => '创建时间(时间戳)', 'default' => 0, 'signed' => false]);
        $table->addColumn('update_time', 'integer', ['comment' => '更新时间(时间戳)', 'default' => 0, 'signed' => false]);
        $table->addColumn('delete_time', 'integer', ['comment' => '删除时间(时间戳)', 'default' => 0, 'signed' => false]);

        // 创建索引
        $table->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid']);
        $table->addIndex(['delete_time'], ['name' => 'idx_delete_time']);

        $table->create();
    }
    
}

