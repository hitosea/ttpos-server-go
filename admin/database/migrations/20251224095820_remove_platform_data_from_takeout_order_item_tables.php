<?php

use Phinx\Db\Adapter\MysqlAdapter;
use think\migration\Migrator;

class RemovePlatformDataFromTakeoutOrderItemTables extends Migrator
{
    /**
     * 删除外卖订单相关表的 platform_data 字段
     * - takeout_order
     * - takeout_order_item
     * - takeout_order_item_modifier
     * - takeout_order_promo
     */
    public function change()
    {
        // 删除 takeout_order 表的 platform_data 字段
        if ($this->hasTable('takeout_order')) {
            $table = $this->table('takeout_order');
            
            // 检查字段是否存在
            if ($table->hasColumn('platform_data')) {
                $table->removeColumn('platform_data')->update();
            }

            // 检查字段是否存在
            if ($table->hasColumn('total_amount')) {
                $table->removeColumn('total_amount')->update();
            }
            
        }

        // 删除 takeout_order_item 表的 platform_data 字段
        if ($this->hasTable('takeout_order_item')) {
            $table = $this->table('takeout_order_item');
            
            // 检查字段是否存在
            if ($table->hasColumn('platform_data')) {
                $table->removeColumn('platform_data')
                    ->update();
            }

            // 检查字段是否存在
            if ($table->hasColumn('ttpos_sku_uuid')) {
                $table->removeColumn('ttpos_sku_uuid')->update();
            }
        }

        // 删除 takeout_order_item_modifier 表的 platform_data 字段
        if ($this->hasTable('takeout_order_item_modifier')) {
            $table = $this->table('takeout_order_item_modifier');
            
            // 检查字段是否存在
            if ($table->hasColumn('platform_data')) {
                $table->removeColumn('platform_data')->update();
            }
        }

        // 删除 takeout_order_promo 表的 platform_data 字段
        if ($this->hasTable('takeout_order_promo')) {
            $table = $this->table('takeout_order_promo');
            
            // 检查字段是否存在
            if ($table->hasColumn('platform_data')) {
                $table->removeColumn('platform_data')->update();
            }
        }
    }
}

