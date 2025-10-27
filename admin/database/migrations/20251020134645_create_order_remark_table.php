<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class CreateOrderRemarkTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     *
     * The following commands can be used in this method and Phinx will
     * automatically reverse them when rolling back:
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
        // 整单备注表
        if (!$this->hasTable('order_remark')) {
            $table = $this->table('order_remark', ['comment' => '整单备注表']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '整单备注ID'])
                ->addColumn('name', 'string', ['limit' => 255, 'default' => '', 'comment' => '名称'])
                ->addColumn('multi_language_name_uuid', 'biginteger', ['default' => 0, 'comment' => '多语言名称ID'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true])
                ->create();
        }
    }
}
