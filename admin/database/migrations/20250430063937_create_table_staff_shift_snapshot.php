<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTableStaffShiftSnapshot extends Migrator
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
        if (!$this->hasTable('staff_shift_snapshot')) {
            $table = $this->table('staff_shift_snapshot', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '员工交班快照表'
            ]);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '交班快照ID'])
                ->addColumn('shift_log_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '交班记录ID'])
                ->addColumn('content', 'text', ['comment' => '快照json'])
                ->addColumn('create_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '删除时间'])
                ->create();
        }
    }
}
