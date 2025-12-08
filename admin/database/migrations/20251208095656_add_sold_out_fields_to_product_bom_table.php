<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSoldOutFieldsToProductBomTable extends Migrator
{
    /**
     * 添加沽清相关字段到 ttpos_product_bom 表
     */
    public function change()
    {
        $table = $this->table('product_bom');

         // 添加 sellable_quantity 字段
         if (!$table->hasColumn('sellable_quantity')) {
            $table->addColumn('sellable_quantity', 'decimal', [
                'precision' => 22,
                'scale' => 4,
                'null' => false,
                'default' => 0.0000,
                'comment' => '可售数量',
                'after' => 'stock_num'
            ])->update();
        }

        // 添加 use_bom_card_stock 字段
        if (!$table->hasColumn('use_bom_card_stock')) {
            $table->addColumn('use_bom_card_stock', 'integer', [
                'limit' => 1,
                'null' => false,
                'default' => 1,
                'comment' => '是否使用成本卡库存，0-否 1-是',
                'after' => 'sellable_quantity'
            ])->update();
        }
    }
}

