<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddBalanceAmountToPaymentOrder extends Migrator
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
        $table = $this->table('payment_order');
        if (!$table->hasColumn('balance_amount')) {
            $table->addColumn('balance_amount', 'decimal', [
                'default' => 0,
                'comment' => '主账户金额,用于反结账时退款',
                'null' => false,
                'after' => 'status',
                'precision' => 12,
                'scale' => 2,
                'signed' => true,
            ]);
        }
        if (!$table->hasColumn('gift_balance_amount')) {
            $table->addColumn('gift_balance_amount', 'decimal', [
                'default' => 0,
                'comment' => '赠送帐户金额,用于反结账时退款',
                'null' => false,
                'after' => 'balance_amount',
                'precision' => 12,
                'scale' => 2,
                'signed' => true,
            ]);
        }
        $table->update();
    }
}
