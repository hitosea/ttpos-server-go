<?php

use think\migration\Migrator;

class AddMemberSaleOrderUuidToLlPaymentOrderTable extends Migrator
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
        // 检查表是否存在
        if ($this->hasTable('ll_payment_order')) {
            $table = $this->table('ll_payment_order');
            if (!$table->hasColumn('member_sale_order_uuid')) {
                $table->addColumn('member_sale_order_uuid', 'biginteger', [
                    'null' => false,
                    'default' => 0,
                    'comment' => '会员销售订单UUID',
                    'after' => 'pay_time'
                ]);
                $table->update();
            }
        }
    }
} 