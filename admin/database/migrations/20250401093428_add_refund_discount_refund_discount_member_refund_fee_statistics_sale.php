<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddRefundDiscountRefundDiscountMemberRefundFeeStatisticsSale extends Migrator
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
        if (!$table->hasColumn('refund_discount') && !$table->hasColumn('refund_discount_member') && !$table->hasColumn('refund_fee')) {
            $table->addColumn('refund_discount', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '退款优惠折扣', 'after' => 'refund_service_fee']);
            $table->addColumn('refund_discount_member', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '退款会员折扣', 'after' => 'refund_discount']);
            $table->addColumn('refund_fee', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '退款支付手续费', 'after' => 'refund_discount_member']);
            $table->update();
        }
    }
}
