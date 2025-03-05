<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdatePercentPaymentOrder extends Migrator
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
        if ($table->hasColumn('payment_fee_percent')) {
            $table->changeColumn('payment_fee_percent', 'decimal', ['precision' => 5, 'scale' => 4, 'default' => 0, 'comment' => '支付手续费百分比,取值范围0-1'])->update();
        }
    }
}
