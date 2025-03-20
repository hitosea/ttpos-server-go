<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddCommissionFeeToLlPaymentOrderTable extends Migrator
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
        $table = $this->table('ll_payment_order');
        if (!$table->hasColumn('commission_fee')) {
            $table->addColumn('commission_fee', 'decimal', [
                'null' => false,
                'default' => 0,
                'comment' => '支付手续费',
                'after' => 'order_amount',
            ]);
        }
        // 更新表
        $table->update();
    }
}
