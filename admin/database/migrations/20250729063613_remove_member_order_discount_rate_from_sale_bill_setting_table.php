<?php

use Phinx\Migration\AbstractMigration;

class RemoveMemberOrderDiscountRateFromSaleBillSettingTable extends AbstractMigration
{
    /**
     * 删除 ttpos_sale_bill_setting 表的 member_order_discount_rate 字段
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('sale_bill_setting')) {
            return;
        }

        $table = $this->table('sale_bill_setting');
        
        // 检查字段是否存在
        if (!$table->hasColumn('member_order_discount_rate')) {
            return;
        }

        // 删除 member_order_discount_rate 字段
        $table->removeColumn('member_order_discount_rate');

        $table->save();
    }
} 