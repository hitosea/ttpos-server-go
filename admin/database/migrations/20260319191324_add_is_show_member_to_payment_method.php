<?php

use think\migration\Migrator;
use think\migration\db\Column;

/**
 * 任务 #40544: 会员端增加QR PromptPay支付方式
 *
 * 为 ttpos_payment_method 表添加 is_show_member 字段
 * 用于控制支付方式在会员端（堂食订单、外送订单）的显示
 */
class AddIsShowMemberToPaymentMethod extends Migrator
{
    /**
     * Change Method.
     */
    public function change()
    {
        if (!$this->hasTable('payment_method')) {
            return;
        }

        $table = $this->table('payment_method');

        if (!$table->hasColumn('is_show_member')) {
            $table->addColumn('is_show_member', 'integer', [
                'limit' => \Phinx\Db\Adapter\MysqlAdapter::INT_TINY,
                'null' => false,
                'default' => 0,
                'comment' => '0-不显示 1-会员端显示（堂食、外送订单）',
                'after' => 'is_show_member_recharge',
            ])->update();
        }
    }
}
