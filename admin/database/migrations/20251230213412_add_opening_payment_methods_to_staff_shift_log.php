<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddOpeningPaymentMethodsToStaffShiftLog extends Migrator
{
    /**
     * 为 ttpos_staff_shift_log 表添加 opening_payment_methods 字段
     * 
     * 用途：存储班次开账时的支付方式UUID列表（逗号分隔）
     * 
     * 涉及表：
     * 1. ttpos_staff_shift_log - 添加 opening_payment_methods 字段
     * 
     * 版本兼容性：向前兼容，旧数据 opening_payment_methods 默认为 NULL
     */
    public function change()
    {
        $this->addOpeningPaymentMethodsField();
    }
    
    /**
     * 为班次记录表添加开账支付方式字段
     */
    private function addOpeningPaymentMethodsField()
    {
        if (!$this->hasTable('staff_shift_log')) {
            return;
        }
        $table = $this->table('staff_shift_log');
        
        if (!$table->hasColumn('opening_payment_methods')) {
            $table->addColumn('opening_payment_methods', 'string', [
                'limit' => 2000,
                'null' => false,
                'default' => '',
                'comment' => '开账时的支付方式UUID列表（逗号分隔）',
                'after' => 'erpnext_async_record_id'
            ])->update();
        }
    }
}

