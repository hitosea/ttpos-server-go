<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateDataManageTable extends Migrator
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
        // 检查表是否已经存在
        if (!$this->hasTable('data_manage')) {
            $table = $this->table('data_manage', ['comment' => '数据管理']);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一ID'])
                ->addColumn('type', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '数据类型 0订单'])
                ->addColumn('data_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '数据UUID'])
                ->addColumn('staff_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '员工UUID'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
                ->create();
            $table->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid']);
            $table->addIndex(['type', 'data_uuid'], ['name' => 'idx_type_data_uuid']);
            $table->addIndex(['staff_uuid'], ['name' => 'idx_staff_uuid']);
        }
    }
}
