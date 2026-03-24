<?php
/**
 * 给 sale_bill 表增加 is_order_first_pay_later 字段
 * 用于标记该订单是否为"先下单后付款"模式
 * 在会员端调用 submit 提交订单时写入 1
 */

use think\migration\Migrator;

class AddIsOrderFirstPayLaterToSaleBill extends Migrator
{
    public function change()
    {
        if ($this->hasTable('sale_bill')) {
            $table = $this->table('sale_bill');

            if (!$table->hasColumn('is_order_first_pay_later')) {
                $table->addColumn('is_order_first_pay_later', 'integer', [
                    'default' => 0,
                    'null'    => false,
                    'comment' => '是否先下单后付款, 0-否 1-是',
                    'after'   => 'is_split_order',
                ]);
            }

            $table->update();
        }
    }
}
