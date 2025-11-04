<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTransferOrderItemTables extends Migrator
{
    /**
     * 创建调拨单相关表
     */
    public function change()
    {
        
        // 创建调拨单明细表
        $this->createTransferOrderItemTable();
        
        // 创建调拨单明细单位表
        $this->createTransferOrderItemUnitTable();
    }

    /**
     * 创建调拨单明细表
     */
    private function createTransferOrderItemTable()
    {
        if (!$this->hasTable('transfer_order_item')) {
            $table = $this->table('transfer_order_item', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '调拨单明细表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '主键UUID'])
                ->addColumn('transfer_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '调拨单UUID'])
                ->addColumn('company_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '所属公司UUID'])
                ->addColumn('headquarter_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '总部UUID'])
                ->addColumn('material_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '物品UUID'])
                ->addColumn('material_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '物品编码'])
                ->addColumn('material_name', 'text', ['comment' => '物品名称JSON'])
                ->addColumn('material_internal_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '物品内部编码'])
                ->addColumn('valuation', 'decimal', ['precision' => 20, 'scale' => 8, 'default' => 0.00000000, 'comment' => '估值单价（基准单位）'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['transfer_order_uuid'], ['name' => 'idx_transfer_order_uuid'])
                ->addIndex(['material_uuid'], ['name' => 'idx_material_uuid'])
                ->addIndex(['company_uuid'], ['name' => 'idx_company_uuid'])
                ->create();
        }
    }

    /**
     * 创建调拨单明细单位表
     */
    private function createTransferOrderItemUnitTable()
    {
        if (!$this->hasTable('transfer_order_item_unit')) {
            $table = $this->table('transfer_order_item_unit', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '调拨单明细单位表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '主键UUID'])
                ->addColumn('item_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '调拨单明细UUID'])
                ->addColumn('transfer_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '调拨单UUID'])
                ->addColumn('unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '单位UUID'])
                ->addColumn('unit_name', 'text', ['comment' => '单位名称JSON'])
                ->addColumn('unit_conversion_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1.0000, 'comment' => '单位转换率'])
                ->addColumn('num', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0.0000, 'comment' => '调拨数量'])
                ->addColumn('erpnext_uom', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERP单位'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['item_uuid'], ['name' => 'idx_item_uuid'])
                ->addIndex(['transfer_order_uuid'], ['name' => 'idx_transfer_order_uuid'])
                ->create();
        }
    }

}

