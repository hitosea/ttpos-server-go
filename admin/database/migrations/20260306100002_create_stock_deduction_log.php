<?php

use think\migration\Migrator;
use think\migration\db\Column;

/**
 * 创建库存扣减日志表
 * 记录每笔订单每个 item_code 的 Stock Entry 扣减状态
 * 用于防止 Stock Entry 合并扣减时部分成功导致的重复扣减
 */
class CreateStockDeductionLog extends Migrator
{
    public function change()
    {
        if ($this->hasTable('stock_deduction_log')) {
            return;
        }

        $this->table('stock_deduction_log', [
            'id' => false,
            'primary_key' => 'id',
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_unicode_ci',
            'comment' => '库存扣减日志-Stock Entry item级别记录',
        ])
        ->addColumn('id', 'biginteger', ['identity' => true])
        ->addColumn('uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => 'UUID'])
        ->addColumn('sale_order_uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '关联订单UUID'])
        ->addColumn('erp_code', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => 'ERP物品编码(item_code)'])
        ->addColumn('qty', 'decimal', ['precision' => 14, 'scale' => 4, 'null' => false, 'default' => 0, 'comment' => '扣减数量'])
        ->addColumn('stock_entry_name', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '关联的Stock Entry单据名'])
        ->addColumn('create_time', 'integer', ['limit' => 11, 'null' => false, 'default' => 0, 'comment' => '创建时间'])
        ->addColumn('update_time', 'integer', ['limit' => 11, 'null' => false, 'default' => 0, 'comment' => '更新时间'])
        ->addColumn('delete_time', 'integer', ['limit' => 11, 'null' => false, 'default' => 0, 'comment' => '删除时间'])
        ->addIndex(['uuid'], ['unique' => true, 'name' => 'idx_uuid'])
        ->addIndex(['sale_order_uuid', 'erp_code'], ['name' => 'idx_order_erp_code'])
        ->addIndex(['erp_code'], ['name' => 'idx_erp_code'])
        ->create();
    }
}
