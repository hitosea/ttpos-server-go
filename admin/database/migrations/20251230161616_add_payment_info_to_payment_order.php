<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddPaymentInfoToPaymentOrder extends Migrator
{
    /**
     * 为 ttpos_payment_order 表添加 payment_info 字段
     * 
     * 用途：存储第三方支付返回的支付信息（JSON 格式）
     * 
     * 涉及表：
     * 1. ttpos_payment_order - 添加 payment_info 字段
     * 
     * 版本兼容性：向前兼容，旧数据 payment_info 默认为空字符串
     */
    public function change()
    {
        $this->addPaymentInfoField();
    }
    
    /**
     * 为支付订单表添加支付信息字段
     */
    private function addPaymentInfoField()
    {
        if (!$this->hasTable('payment_order')) {
            return;
        }
        $table = $this->table('payment_order');
        
        if (!$table->hasColumn('payment_info')) {
            $table->addColumn('payment_info', 'text', [
                'null' => false,
                'default' => '',
                'comment' => '支付信息(JSON格式,存储第三方支付返回的详细信息)',
                'after' => 'status_reason'
            ])->update();
        }
    }
}

