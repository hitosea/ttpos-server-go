<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddFieldsToStatisticsSale extends Migrator
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
        $table = $this->table('statistics_sale');
        if (!$table->hasColumn('refund_payment_balance')) {
            $table->addColumn('refund_payment_balance', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '退款支付余额', 'after' => 'refund_amount'])
                ->update();
        }
    }
}
