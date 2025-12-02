<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddAddPriceToSaleOrderProductTable extends Migrator
{
    /**
     * 添加加价金额字段到 ttpos_sale_order_product 表
     */
    public function change()
    {
        $table = $this->table('sale_order_product');

        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('add_price')) {
            $table->addColumn('add_price', 'decimal', [
                'precision' => 22,
                'scale' => 4,
                'null' => false,
                'default' => 0.00,
                'comment' => '加价金额。子商品记录单商品加价金额；套餐主商品记录所有子商品加价总和',
                'after' => 'sauce_price'
            ])->update();
        }
    }
}

