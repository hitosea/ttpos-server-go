<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddTakeoutOrderUuidToProductionOrderTable extends Migrator
{
    /**
     * 添加外卖订单uuid字段到 ttpos_production_order 表
     */
    public function change()
    {
        if ($this->hasTable('production_order')) {
            $table = $this->table('production_order');
            // 检查字段是否已存在，如果不存在则添加
            if (!$table->hasColumn('takeout_order_uuid')) {
                $table->addColumn('takeout_order_uuid', 'biginteger', ['limit' => 20, 'null' => false, 'default' => 0, 'comment' => '外卖订单UUID（关联 ttpos_takeout_order.uuid）', 'after' => 'sale_bill_uuid'])->update();
            }
            // 检查索引是否已存在，如果不存在则添加
            if (!$table->hasIndexByName('idx_takeout_order_uuid')) {
                $table->addIndex(['takeout_order_uuid'], ['name' => 'idx_takeout_order_uuid'])->update();
            }
        }

        if ($this->hasTable('production_order_product')) {
            $table = $this->table('production_order_product');
            // 检查字段是否已存在，如果不存在则添加
            if (!$table->hasColumn('takeout_order_item_uuid')) {
                $table->addColumn('takeout_order_item_uuid', 'biginteger', ['limit' => 20, 'null' => false, 'default' => 0, 'comment' => '外卖订单UUID（关联 ttpos_takeout_order.uuid）', 'after' => 'production_order_uuid'])->update();
            }
            // 检查字段是否已存在，如果不存在则添加
            if (!$table->hasColumn('takeout_order_uuid')) {
                $table->addColumn('takeout_order_uuid', 'biginteger', ['limit' => 20, 'null' => false, 'default' => 0, 'comment' => '外卖订单UUID（关联 ttpos_takeout_order.uuid）', 'after' => 'production_order_uuid'])->update();
            }
            // 检查索引是否已存在，如果不存在则添加
            if (!$table->hasIndexByName('idx_takeout_order_item_uuid')) {
                $table->addIndex(['takeout_order_item_uuid'], ['name' => 'idx_takeout_order_item_uuid'])->update();
            }
            if (!$table->hasIndexByName('idx_takeout_order_uuid')) {
                $table->addIndex(['takeout_order_uuid'], ['name' => 'idx_takeout_order_uuid'])->update();
            }
        }

        if ($this->hasTable('production_order_material')) {
            $table = $this->table('production_order_material');
            // 检查字段是否已存在，如果不存在则添加
            if (!$table->hasColumn('takeout_order_item_uuid')) {
                $table->addColumn('takeout_order_item_uuid', 'biginteger', ['limit' => 20, 'null' => false, 'default' => 0, 'comment' => '外卖订单商品UUID（关联 ttpos_takeout_order_item.uuid）', 'after' => 'sale_order_product_uuid'])->update();
            }
            // 检查索引是否已存在，如果不存在则添加
            if (!$table->hasIndexByName('idx_takeout_order_item_uuid')) {
                $table->addIndex(['takeout_order_item_uuid'], ['name' => 'idx_takeout_order_item_uuid'])->update();
            }
        }

        // 修改 ttpos_takeout 表的 enabled 字段默认值为 0
        if ($this->hasTable('takeout')) {
            $table = $this->table('takeout');
            // 检查字段是否存在
            if ($table->hasColumn('enabled')) {
                $table->changeColumn('enabled', 'integer', [
                    'limit' => 11,
                    'null' => false,
                    'default' => 0,
                    'comment' => '是否启用,0-未启用 1-已启用'
                ])->update();
            }
        }
    }
}

