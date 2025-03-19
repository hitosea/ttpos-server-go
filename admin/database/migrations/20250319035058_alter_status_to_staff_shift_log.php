<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AlterStatusToStaffShiftLog extends Migrator
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
        $table = $this->table('staff_shift_log');
        if ($table->hasColumn('status')) {
            $table->changeColumn('status', 'tinyinteger', ['limit' => 1, 'null' => false, 'default' => 0, 'comment' => '交班状态 0-未交班 1-已交班', 'after' => 'shift_no']);
        }
        $table->update();
    }
}
