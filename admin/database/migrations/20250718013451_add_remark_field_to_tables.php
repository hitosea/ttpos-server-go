<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddRemarkFieldToTables extends Migrator
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
        // 会员余额日志
        $table = $this->table('member_balance_log');
        if (!$table->hasColumn('remark')) {
            $table->addColumn('remark', 'text', ['comment' => '备注', 'after' => 'related_uuid'])
                ->update();
        }
        // 会员积分日志
        $table = $this->table('member_point_log');
        if (!$table->hasColumn('remark')) {
            $table->addColumn('remark', 'text', ['comment' => '备注', 'after' => 'related_uuid'])
                ->update();
        }
    }
}
