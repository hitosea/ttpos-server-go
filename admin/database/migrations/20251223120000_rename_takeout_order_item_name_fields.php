<?php

use think\migration\Migrator;

class RenameTakeoutOrderItemNameFields extends Migrator
{
    /**
     * 重命名外卖订单商品和修饰符的名称字段
     * - takeout_order_item.platform_item_name → item_name (varchar → text)
     * - takeout_order_item_modifier.platform_modifier_name → modifier_name (varchar → text)
     */
    public function change()
    {
        // 1. 处理 takeout_order_item 表
        $itemTable = $this->table('takeout_order_item');
        
        if ($itemTable->hasColumn('platform_item_name')) {
            // 先删除旧列
            $itemTable->removeColumn('platform_item_name')->update();
            
            // 添加新列
            $itemTable->addColumn('item_name', 'text', ['null' => true, 'comment' => '商品名称', 'after' => 'platform_item_id'])->update();
        }

        // 2. 处理 takeout_order_item_modifier 表
        $modifierTable = $this->table('takeout_order_item_modifier');
        
        if ($modifierTable->hasColumn('platform_modifier_name')) {
            // 先删除旧列
            $modifierTable->removeColumn('platform_modifier_name')->update();
            
            // 添加新列
            $modifierTable->addColumn('modifier_name', 'text', ['null' => true, 'comment' => '修饰符名称', 'after' => 'platform_modifier_id'])->update();
        }
    }
}

