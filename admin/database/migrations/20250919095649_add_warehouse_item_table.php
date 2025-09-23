<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddWarehouseItemTable extends Migrator
{
    /**
     * 创建 ttpos_warehouse_item 表
     */
    public function change()
    {
        // 检查表是否已存在
        if (!$this->hasTable('warehouse_item')) {
            $table = $this->table('warehouse_item');
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => 'UUID'])
                  ->addColumn('warehouse_uuid', 'biginteger', ['default' => 0, 'comment' => '仓库UUID'])
                  ->addColumn('material_uuid', 'biginteger', ['default' => 0, 'comment' => '商品UUID'])
                  ->addColumn('material_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '商品编码'])
                  ->addColumn('stock', 'decimal', ['precision' => 14, 'scale' => 2, 'default' => 0, 'comment' => '库存数量'])
                  ->addColumn('reserved_stock', 'decimal', ['precision' => 14, 'scale' => 2, 'default' => 0, 'comment' => '预留库存数量'])
                  ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                  ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                  ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                  ->addIndex(['material_uuid'], ['name' => 'idx_material_uuid'])
                  ->addIndex(['material_code'], ['name' => 'idx_material_code'])
                  ->addIndex(['warehouse_uuid'], ['name' => 'idx_warehouse_uuid'])
                  ->create();
        }
    }
}
