<?php

use think\migration\Migrator;

class UpdateTakeoutOrderItemModifierPriceComment extends Migrator
{
    /**
     * 修改 ttpos_takeout_order_item_modifier 表 price 字段描述
     * - takeout_order_item_modifier.price: "价格(元,4位小数)" -> "加价-总加价"
     */
    public function change()
    {
        if ($this->hasTable('takeout_order_item_modifier')) {
            $this->table('takeout_order_item_modifier')
                ->changeColumn('price', 'decimal', [
                    'precision' => 20,
                    'scale' => 4,
                    'default' => '0.0000',
                    'comment' => '加价-单价-外卖平台价格',
                ])
            ->update();
        }
    }
}

