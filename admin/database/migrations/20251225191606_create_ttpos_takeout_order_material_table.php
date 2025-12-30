<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTtposTakeoutOrderMaterialTable extends Migrator
{
    /**
     * 创建外卖订单原料表
     * 用于记录外卖订单使用的原料，供日终汇总出库统计使用
     */
    public function change()
    {
        // 检查表是否已存在
        if (!$this->hasTable('takeout_order_material')) {
            $table = $this->table('takeout_order_material', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '外卖订单原料表',
                'id' => false,
                'primary_key' => ['id']
            ]);

            $table
                // 基础字段
                ->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识'])
                
                // 关联字段
                ->addColumn('takeout_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '外卖订单ID'])
                ->addColumn('takeout_order_item_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '外卖订单商品UUID(关联ttpos_takeout_order_item.uuid)'])
                ->addColumn('takeout_order_item_modifier_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '外卖订单商品修饰符UUID(关联ttpos_takeout_order_item_modifier.uuid)'])
                ->addColumn('material_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '原料ID'])
                ->addColumn('erp_code', 'string', ['limit' => 50, 'default' => '', 'comment' => 'ERP编码(来自Material.Code)'])
                ->addColumn('warehouse_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '仓库ID'])
                
                // 数量字段
                ->addColumn('num', 'decimal', ['precision' => 20, 'scale' => 4, 'default' => '0.0000', 'comment' => '数量(原料的实际使用数量)'])
                
                // 统计字段
                ->addColumn('is_summarized', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'comment' => '是否已经统计: 0=未统计,1=已统计'])
                
                // 标准字段
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
                
                // 索引
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['takeout_order_uuid'], ['name' => 'idx_takeout_order_uuid'])
                ->addIndex(['takeout_order_item_uuid'], ['name' => 'idx_takeout_order_item_uuid'])
                ->addIndex(['takeout_order_item_modifier_uuid'], ['name' => 'idx_takeout_order_item_modifier_uuid'])
                ->addIndex(['material_uuid'], ['name' => 'idx_material_uuid'])
                ->addIndex(['warehouse_uuid'], ['name' => 'idx_warehouse_uuid'])
                ->addIndex(['is_summarized', 'create_time'], ['name' => 'idx_is_summarized_create_time'])
                ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])
                
                ->create();
        }
    }
}

