<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnDutyNoAndRefundMoneyToMemberRechargeOrder extends Migrator
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
        $table = $this->table('member_recharge_order');
        if (!$table->hasColumn('duty_no')) {
            $table->addColumn('duty_no', 'string', ['limit' => 64, 'default' => '','null' => false, 'comment' => '收银员当班编号', 'after' => 'order_no']);
        }
        if (!$table->hasColumn('refund_money')) {
            $table->addColumn('refund_money', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '退款金额', 'after' => 'amount']);
        }
        $table->update();
    }
}
