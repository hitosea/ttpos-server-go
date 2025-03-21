<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddLlRefundOrderIdToReturnOrderAmountTable extends Migrator
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
        // 1. 添加银行编码、账号、账户名称字段
        $table = $this->table('return_order_amount');
        if (!$table->hasColumn('ll_return_order_id')) {
            $table->addColumn('ll_return_order_id', 'string', ['null' => false, 'default' => '', 'comment' => '连连退款订单ID, 用来重新发起退款', 'after' => 'refund_status']);
        }
        $table->update();
    }
}
