<?php

use think\migration\Migrator;

class ChangeTakeoutOrderItemPriceToDecimal extends Migrator
{
    /**
     * 修改外卖订单商品表 price 字段类型为元（4位小数）
     * 
     * 变更说明：
     * - ttpos_takeout_order_item.price: bigint (分) -> decimal(10,4) (元)
     * - ttpos_takeout_order_item.tax: bigint (分) -> decimal(10,4) (元)
     * - ttpos_takeout_order_item_modifier.price: bigint (分) -> decimal(10,4) (元)
     * - ttpos_takeout_order_item_modifier.tax: bigint (分) -> decimal(10,4) (元)
     */
    public function change()
    {
        // 修改 takeout_order_item 表
        if ($this->hasTable('takeout_order_item')) { 
            $this->table('takeout_order_item')
                ->changeColumn('price', 'decimal', [
                    'precision' => 20,
                    'scale' => 4,
                    'default' => '0.0000',
                    'comment' => '单价(元,4位小数)-外卖平台价格',
                ])
                ->changeColumn('tax', 'decimal', [
                    'precision' => 20,
                    'scale' => 4,
                    'default' => '0.0000',
                    'comment' => '税费(元,4位小数)',
                ])
            ->update();
        }

        // 修改 takeout_order_item_modifier 表
        if ($this->hasTable('takeout_order_item_modifier')) {
            $this->table('takeout_order_item_modifier')
                ->changeColumn('price', 'decimal', [
                    'precision' => 20,
                    'scale' => 4,
                    'default' => '0.0000',
                    'comment' => '价格(元,4位小数)',
                ])
                ->changeColumn('tax', 'decimal', [
                    'precision' => 20,
                    'scale' => 4,
                    'default' => '0.0000',
                    'comment' => '税费(元,4位小数)',
                ])
            ->update();
        }
    }
}

