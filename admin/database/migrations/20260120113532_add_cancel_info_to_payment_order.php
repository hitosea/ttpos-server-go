<?php

use think\migration\Migrator;

/**
 * 为支付订单表添加撤销信息字段
 * 
 * 在 ttpos_payment_order 表中添加 cancel_info 字段，用于存储第三方支付返回的撤销信息（JSON格式）
 */
class AddCancelInfoToPaymentOrder extends Migrator
{
    /**
     * Change Method.
     *
     * 为支付订单表添加撤销信息字段
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('payment_order')) {
            return;
        }

        // 获取表对象
        $table = $this->table('payment_order');
        
        // 检查字段是否已存在
        if (!$table->hasColumn('cancel_info')) {
            // 添加 cancel_info 字段
            // 添加到 payment_info 字段之后
            $table->addColumn('cancel_info', 'text', [
                'null' => false,
                'default' => '',
                'comment' => '撤销信息(JSON格式,存储第三方支付返回的撤销信息)',
                'after' => 'payment_info'
            ]);
        }
        
        $table->update();
    }
}
