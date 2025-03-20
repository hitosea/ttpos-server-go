<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddRefundStatusToReturnOrderAmountTable extends Migrator
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
        $table = $this->table('return_order');
        if (!$table->hasColumn('bank_code')) {
            $table->addColumn('bank_code', 'string', ['null' => false, 'default' => '', 'comment' => '银行编码 - 当存在QR PromptPay的时候需要传', 'after' => 'refund_status']);
            $table->addColumn('account_no', 'string', ['null' => false, 'default' => '', 'comment' => '账号 - 当存在QR PromptPay的时候需要传', 'after' => 'bank_code']);
            $table->addColumn('account_name', 'string', ['null' => false, 'default' => '', 'comment' => '账户名称 - 当存在QR PromptPay的时候需要传', 'after' => 'account_no']);
        }
        $table->update();
        // 2. 添加退款状态和商户退款单号字段
        $table = $this->table('return_order_amount');
        if (!$table->hasColumn('refund_status')) {
            $table->addColumn('refund_status', 'integer', ['signed' => false, 'null' => false, 'default' => 1, 'comment' => '退款状态 0-退款中 1-退款成功 2-退款失败', 'after' => 'payment_method_uuid']);
            $table->addColumn('merchant_refund_order_no', 'string', ['null' => false, 'default' => '', 'comment' => '商户退款单号', 'after' => 'refund_status']);
        }
        $table->update();
    }
}
