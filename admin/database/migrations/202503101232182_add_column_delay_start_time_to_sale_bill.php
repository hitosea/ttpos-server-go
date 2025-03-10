<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnDelayStartTimeToSaleBill extends Migrator
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
        $table = $this->table('sale_bill');
        if (!$table->hasColumn('delay_start_time')) {
            $table->addColumn('delay_duration', 'integer', ['default' => 0, 'comment' => '总延迟时长(秒)', 'after' => 'buffet_start_time']);
            $table->addColumn('delay_start_time', 'integer', ['default' => 0, 'comment' => '总延迟时长开始时间(秒)', 'after' => 'delay_duration']);
            $table->update();
        }
    }
}
	