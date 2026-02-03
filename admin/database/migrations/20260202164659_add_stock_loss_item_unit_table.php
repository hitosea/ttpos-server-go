<?php

use think\migration\Migrator;
use think\migration\db\Column;

/**
 * 添加报损单物品单位明细表
 */
class AddStockLossItemUnitTable extends Migrator
{
    public function change()
    {
        $this->createStockLossItemUnitTable();
        $this->modifyStockLossItemTable();
    }

    /**
     * 创建报损单物品单位明细表
     */
    protected function createStockLossItemUnitTable()
    {
        $table = $this->table('stock_loss_item_unit', [
            'id' => false,
            'primary_key' => 'id',
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_unicode_ci',
            'comment' => '报损单物品单位明细表',
        ]);

        $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
            ->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '报损单物品单位明细UUID'])
            ->addColumn('stock_loss_item_uuid', 'biginteger', ['default' => 0, 'comment' => '报损单物品明细UUID'])
            ->addColumn('material_unit_uuid', 'biginteger', ['default' => 0, 'comment' => '物料单位UUID'])
            ->addColumn('material_unit_name', 'text', ['null' => true, 'comment' => '物料单位名称（多语言JSON备份）'])
            ->addColumn('quantity', 'decimal', ['precision' => 22, 'scale' => 4, 'null' => true, 'comment' => '单位数量'])
            ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
            ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
            ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
            ->addIndex(['stock_loss_item_uuid'], ['name' => 'idx_stock_loss_item_uuid'])
            ->addIndex(['material_unit_uuid'], ['name' => 'idx_material_unit_uuid'])
            ->create();
    }

    /**
     * 修改报损单物品明细表，添加基准单位数量字段
     */
    protected function modifyStockLossItemTable()
    {
        $table = $this->table('stock_loss_item');

        // 添加基准单位数量字段（用于存储转换后的总量）
        $table->addColumn('base_quantity', 'decimal', [
            'precision' => 14,
            'scale' => 4,
            'default' => '0.0000',
            'comment' => '基准单位数量（所有单位转换后的总量）',
            'after' => 'quantity'
        ])->update();
    }
}
