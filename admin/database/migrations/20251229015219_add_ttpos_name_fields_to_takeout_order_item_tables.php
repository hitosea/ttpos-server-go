<?php

use think\migration\Migrator;

class AddTtposNameFieldsToTakeoutOrderItemTables extends Migrator
{
    /**
     * 为外卖订单商品和修饰符表添加 TTPOS 名称字段
     * - takeout_order_item.ttpos_item_name (text) - TTPOS 商品名称
     * - takeout_order_item_modifier.ttpos_modifier_name (text) - TTPOS 修饰符名称
     * - takeout_order_item_modifier.ttpos_flavor_uuid (bigint) - TTPOS 规格UUID
     * - takeout_order_item_modifier.ttpos_flavor_name (text) - TTPOS 规格名称
     */
    public function change()
    {
        // 1. 处理 takeout_order_item 表
        if ($this->hasTable('takeout_order_item')) {
            $itemTable = $this->table('takeout_order_item');
            if (!$itemTable->hasColumn('ttpos_item_name')) {
                $itemTable->addColumn('ttpos_item_name', 'text', ['null' => true, 'comment' => 'TTPOS商品名称(来自ttpos_product_package)', 'after' => 'ttpos_product_type'])->update();
            }
        }
      
        // 2. 处理 takeout_order_item_modifier 表
        if ($this->hasTable('takeout_order_item_modifier')) {
            $modifierTable = $this->table('takeout_order_item_modifier');
            if (!$modifierTable->hasColumn('ttpos_modifier_name')) {
                $modifierTable->addColumn('ttpos_modifier_name', 'text', ['null' => true, 'comment' => 'TTPOS修饰符名称', 'after' => 'ttpos_modifier_type'])->update();
            }
            if (!$modifierTable->hasColumn('ttpos_flavor_product_bom_uuid')) {
                $modifierTable->addColumn('ttpos_flavor_product_bom_uuid', 'biginteger', [
                    'signed' => false,
                    'default' => 0,
                    'comment' => 'TTPOS规格商品物料UUID(commodity类型对应product_package_group_item.product_bom_uuid)',
                    'after' => 'ttpos_modifier_name',
                ])->update();
            }
            if (!$modifierTable->hasColumn('ttpos_flavor_name')) {
                $modifierTable->addColumn('ttpos_flavor_name', 'text', [
                    'null' => true,
                    'comment' => 'TTPOS规格名称(commodity类型使用)',
                    'after' => 'ttpos_flavor_product_bom_uuid',
                ])->update();
            }
        }
    }
}

