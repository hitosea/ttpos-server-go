<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddWarehouseInOutLogTable extends Migrator
{
    /**
     * 创建 ttpos_warehouse_in_out_log 表
     */
    public function change()
    {
        
        // 检查表是否已存在
        if (!$this->hasTable('warehouse_in_out_log')) {
            $table = $this->table('warehouse_in_out_log');
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '出入库记录ID'])
                  ->addColumn('log_type', 'integer', ['default' => 0, 'comment' => '日志类型,0-入库 1-出库'])
                  ->addColumn('scene', 'integer', ['default' => 0, 'comment' => '场景,0-采购入库 1-销售出库 2-发货出库'])
                  ->addColumn('warehouse_uuid', 'biginteger', ['default' => 0, 'comment' => '仓库ID'])
                  ->addColumn('material_uuid', 'biginteger', ['default' => 0, 'comment' => '物品ID'])
                  ->addColumn('material_name', 'text', ['comment' => '物品名称JSON,记录当时物品名称'])
                  ->addColumn('material_base_unit_uuid', 'biginteger', ['default' => 0, 'comment' => '物品基准单位ID'])
                  ->addColumn('material_base_unit_name', 'text', ['comment' => '物品基准单位名称'])
                  ->addColumn('num', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0, 'comment' => '数量'])
                  ->addColumn('price', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0, 'comment' => '单价，物品基准单位单价'])
                  ->addColumn('amount', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0, 'comment' => '金额,单价*数量'])
                  ->addColumn('supplier_uuid', 'biginteger', ['default' => 0, 'comment' => '供应商ID'])
                  ->addColumn('order_no', 'string', ['limit' => 255, 'default' => '', 'comment' => '单据编号'])
                  ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                  ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                  ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                  ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                  ->create();
        }
    }
}
