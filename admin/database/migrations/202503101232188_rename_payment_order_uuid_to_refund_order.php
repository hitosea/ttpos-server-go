<?php

use think\migration\Migrator;
use think\migration\db\Column;

class RenamePaymentOrderUuidToRefundOrder extends Migrator
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
        $table = $this->table('refund_order');
        if ($table->hasColumn('payment_bill_uuid')) {
            $table->renameColumn('payment_bill_uuid', 'payment_order_uuid')
                ->update();
        }
        if ($table->hasColumn('refund_amount')) {
            $table->renameColumn('refund_amount', 'amount')
                ->update();
        }
        if ($table->hasColumn('refund_reason')) {
            $table->renameColumn('refund_reason', 'reason')
                ->update();
        }
        if ($table->hasColumn('refund_status')) {
            $table->renameColumn('refund_status', 'status')
                ->update();
        }
        if ($table->hasColumn('refund_type')) {
            $table->changeColumn('refund_type', 'tinyinteger', ['default' => 1, 'comment' => '退款类型,1-反结账,2-取消付款', 'after' => 'payment_order_uuid'])
                ->update();
        }
    }
}
	