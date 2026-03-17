<?php
/**
 * 为销售账单表添加提交支付时间字段
 *
 * 功能: 记录会员端堂食订单提交支付的时间，用于订单列表过滤（仅显示已提交支付的订单）
 */

use think\migration\Migrator;

class AddSubmitPayTimeToSaleBill extends Migrator
{
    /**
     * Change Method.
     */
    public function change()
    {
        if ($this->hasTable('sale_bill')) {
            $table = $this->table('sale_bill');

            if (!$table->hasColumn('submit_pay_time')) {
                $table->addColumn('submit_pay_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'null' => false,
                    'comment' => '提交支付时间（时间戳）',
                    'after' => 'production_time'
                ]);
            }

            $table->update();
        }
    }
}
