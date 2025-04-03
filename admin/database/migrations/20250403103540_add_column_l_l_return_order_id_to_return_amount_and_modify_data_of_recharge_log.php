<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnLLReturnOrderIdToReturnAmountAndModifyDataOfRechargeLog extends Migrator
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
        $table = $this->table('member_recharge_order_operation_log');
        if ($table->hasColumn('data')) {
            $table->changeColumn('data', 'text', ['comment' => '数据']);
            $table->update();
        }
        $table = $this->table('return_order_amount');
        if (!$table->hasColumn('ll_return_order_id')) {
            $table->addColumn('ll_return_order_id', 'string', ['default' => '', 'comment' => '连连退款订单ID, 用来重新发起退款', 'after' => 'refund_status']);
            $table->update();
        }
    }
}
