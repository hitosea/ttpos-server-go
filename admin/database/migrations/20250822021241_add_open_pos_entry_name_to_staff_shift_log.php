<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddOpenPosEntryNameToStaffShiftLog extends Migrator
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
        // 新增开账名称字段（若不存在）
        if (!$table->hasColumn('erpnext_open_pos_entry_name')) {
            $table->addColumn('erpnext_open_pos_entry_name', 'string', ['limit' => 255, 'default' => '', 'comment' => 'erpnext开账名称', 'after' => 'shift_start_time']);
        }
        $table->update();
    }
}
