<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddStockReconciliationTables extends Migrator
{
    /**
     * 创建盘点相关表
     */
    public function change()
    {
        // 创建盘点单表
        if (!$this->hasTable('stock_reconciliation')) {
            $table = $this->table('stock_reconciliation', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '盘点单表',
            ]);

            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '盘点单ID'])
                ->addColumn('order_no', 'string', ['limit' => 255, 'default' => '', 'comment' => '单据编号'])
                ->addColumn('erp_code', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERP盘点单号'])
                ->addColumn('type', 'integer', ['signed' => true, 'default' => 1, 'comment' => '盘点类型 1-指定物品盘点 2-全部物品盘点'])
                ->addColumn('warehouse_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '仓库ID'])
                ->addColumn('purpose', 'integer', ['signed' => true, 'default' => 1, 'comment' => '盘点目的 1-库存盘点 2-期初盘点'])
                ->addColumn('status', 'integer', ['signed' => true, 'default' => 0, 'comment' => '状态 0-已保存 1-已提交 2-已审核 3-已驳回'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['warehouse_uuid'], ['name' => 'idx_warehouse_uuid'])
                ->addIndex(['status'], ['name' => 'idx_status'])
                ->addIndex(['order_no'], ['name' => 'idx_order_no'])
                ->create();
        }

        // 创建盘点单物品明细表
        if (!$this->hasTable('stock_reconciliation_item')) {
            $table = $this->table('stock_reconciliation_item', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '盘点单物品明细表',
            ]);

            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '盘点单物品明细ID'])
                ->addColumn('stock_reconciliation_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '盘点单ID'])
                ->addColumn('material_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '物品ID'])
                ->addColumn('material_name', 'text', ['comment' => '物品名称，用于备份多语言'])
                ->addColumn('booked_quantity', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0.0000, 'comment' => '账面库存数量，基准单位后的数量'])
                ->addColumn('counted_quantity', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0.0000, 'comment' => '实盘库存数量，物品所有单位换算成基准单位后的数量'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['stock_reconciliation_uuid'], ['name' => 'idx_stock_reconciliation_uuid'])
                ->addIndex(['material_uuid'], ['name' => 'idx_material_uuid'])
                ->create();
        }

        // 创建盘点单物品单位明细表
        if (!$this->hasTable('stock_reconciliation_item_unit')) {
            $table = $this->table('stock_reconciliation_item_unit', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '盘点单物品单位明细表',
            ]);

            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '盘点单物品单位明细ID'])
                ->addColumn('stock_reconciliation_item_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '盘点单物品明细ID'])
                ->addColumn('material_unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '单位ID'])
                ->addColumn('material_unit_name', 'text', ['comment' => '物品单位名称，用于备份多语言'])
                ->addColumn('quantity', 'decimal', ['precision' => 22, 'scale' => 4, 'null' => true, 'comment' => '单位数量'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['stock_reconciliation_item_uuid'], ['name' => 'idx_stock_reconciliation_item_uuid'])
                ->addIndex(['material_unit_uuid'], ['name' => 'idx_material_unit_uuid'])
                ->create();
        }
    }
}

