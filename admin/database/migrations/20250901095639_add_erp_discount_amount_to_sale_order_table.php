<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpDiscountAmountToSaleOrderTable extends Migrator
{
    /**
     * 添加 erp_discount_amount 字段到 ttpos_sale_order 表
     */
    public function change()
    {
        $table = $this->table('sale_order');
        
        // 检查字段是否已存在
        if (!$table->hasColumn('erp_discount_amount')) {
            $table->addColumn('erp_discount_amount', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0, 'comment' => '订单应收优惠金额,整单改价、订单抹零、结账抹零、优惠券抵扣、积分抵扣，以上总共的优惠金额（正常是负数，有时候是正数）', 'after' => 'erp_material_invoice_name'])
                  ->update();
        }
    }
}
