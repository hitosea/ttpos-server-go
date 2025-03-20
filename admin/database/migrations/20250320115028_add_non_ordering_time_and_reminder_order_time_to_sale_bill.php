<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddNonOrderingTimeAndReminderOrderTimeToSaleBill extends Migrator
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
        if (!$table->hasColumn('reminder_order_time')) {
            $table->addColumn('reminder_order_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '自助餐结束前x分钟时提醒不可下单，用于助手端、平板端和h5', 'after' => 'buffet_duration'])->update();
        }
        if (!$table->hasColumn('non_ordering_time')) {
            $table->addColumn('non_ordering_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '自助餐结束前x分钟时不可下单，用于助手端、平板端和h5', 'after' => 'buffet_duration'])->update();
        }
    }
}
