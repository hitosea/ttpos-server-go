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
            $table->addColumn('erp_discount_amount', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0, 'comment' => '订单应收优惠金额，整单改价优惠掉的金额', 'after' => 'erp_material_invoice_name'])
                  ->update();
        }
    }
}
