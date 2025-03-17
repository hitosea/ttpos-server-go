<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddBalanceToRechargeOrder extends Migrator
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
        if (!$table->hasColumn('balance_recharged')) {
            $table->addColumn('balance_recharged', 'decimal', ['precision' => 12, 'scale' => 2, 'after' => 'payment_time','default' => 0, 'comment' => '充值后会员余额']);
        }
        if (!$table->hasColumn('balance')) {
            $table->addColumn('balance', 'decimal', ['precision' => 12, 'scale' => 2, 'after' => 'payment_time', 'default' => 0, 'comment' => '充值前会员余额']);
        }
        $table->update();
    }
}
