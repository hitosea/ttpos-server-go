<?php

use think\migration\Migrator;

class UpdateTtposPaymentOrderUuidTable extends Migrator
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
        if ($table->hasColumn('uuid')) {
            $table->changeColumn('uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '支付订单ID'])
                  ->update();
        }
        if ($table->hasColumn('payment_type_uuid')) {
            $table->changeColumn('payment_type_uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '支付类型ID'])
                  ->update();
        }
        if ($table->hasColumn('sale_order_uuid')) {
            $table->renameColumn('sale_order_uuid', 'related_uuid')
                  ->changeColumn('related_uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '充值订单、销售订单ID'])
                  ->update();
        }
        if ($table->hasColumn('amount')) {
            $table->changeColumn('amount', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '金额：支付金额+支付手续费'])
                ->update();
        }

        $table = $this->table('member_recharge_order');
        if ($table->hasColumn('payment_amount')) {
            $table->changeColumn('payment_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '交易金额，确认充值后设置'])
                ->update();
        }
        if ($table->hasColumn('gift_point')) {
            $table->changeColumn('gift_point', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '赠送积分'])
                ->update();
        }
    }
}