<?php

use think\migration\Migrator;

class AddTtposProductUuidToTakeoutOrderItemModifier extends Migrator
{
    /**
     * 为外卖订单商品修饰符表添加 ttpos_product_uuid 字段
     * - takeout_order_item_modifier.ttpos_product_uuid (bigint unsigned) - TTPOS 商品 UUID
     * 
     * 用于存储修饰符所属的商品 UUID，便于追溯修饰符来源
     */
    public function change()
    {
        if ($this->hasTable('takeout_order_item_modifier')) {
            $table = $this->table('takeout_order_item_modifier');
            if (!$table->hasColumn('ttpos_product_uuid')) {
                $table->addColumn('ttpos_product_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'TTPOS商品UUID(关联ttpos_product_package.uuid)', 'after' => 'ttpos_modifier_name'])->update();
            }
        }
    }
}

