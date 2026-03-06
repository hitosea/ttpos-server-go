<?php

use think\migration\Migrator;
use think\migration\db\Column;
use Phinx\Db\Adapter\MysqlAdapter;

/**
 * 创建盘点快照表
 * 记录盘点时未通过 Stock Entry 扣减的订单商品快照
 * 用于盘点差异计算：预期库存 = ERP账面 - 快照扣减量
 */
class CreateStocktakeSnapshot extends Migrator
{
    public function change()
    {
        if ($this->hasTable('stocktake_snapshot')) {
            return;
        }

        $this->table('stocktake_snapshot', [
            'id' => false,
            'primary_key' => 'id',
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_unicode_ci',
            'comment' => '盘点快照-未出库订单快照',
        ])
        ->addColumn('id', 'biginteger', ['identity' => true])
        ->addColumn('uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => 'UUID'])
        ->addColumn('stock_reconciliation_uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '关联盘点单UUID'])
        ->addColumn('item_code', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => 'ERP物品编码'])
        ->addColumn('warehouse_erp_code', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => 'ERP仓库编码'])
        ->addColumn('pending_qty', 'decimal', ['precision' => 12, 'scale' => 3, 'null' => false, 'default' => 0, 'comment' => '待扣减数量'])
        ->addColumn('order_count', 'integer', ['null' => false, 'default' => 0, 'comment' => '涉及订单数'])
        ->addColumn('create_time', 'integer', ['limit' => 11, 'null' => false, 'default' => 0, 'comment' => '创建时间'])
        ->addColumn('update_time', 'integer', ['limit' => 11, 'null' => false, 'default' => 0, 'comment' => '更新时间'])
        ->addColumn('delete_time', 'integer', ['limit' => 11, 'null' => false, 'default' => 0, 'comment' => '删除时间'])
        ->addIndex(['uuid'], ['unique' => true, 'name' => 'idx_uuid'])
        ->addIndex(['stock_reconciliation_uuid'], ['name' => 'idx_stock_reconciliation_uuid'])
        ->addIndex(['item_code', 'warehouse_erp_code'], ['name' => 'idx_item_warehouse'])
        ->create();
    }
}
