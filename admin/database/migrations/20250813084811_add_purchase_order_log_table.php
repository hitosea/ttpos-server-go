<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddPurchaseOrderLogTable extends Migrator
{
    /**
     * 迁移
     */
    public function change()
    {
        // 创建采购订单操作日志表
        $this->createPurchaseOrderLogTable();
    }

    /**
     * 创建采购订单操作日志表
     */
    private function createPurchaseOrderLogTable()
    {
        if (!$this->hasTable('purchase_order_log')) {
            $table = $this->table('purchase_order_log', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '采购订单操作日志表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '操作日志ID'])
                ->addColumn('purchase_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '采购订单ID'])
                ->addColumn('operator_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '操作人ID'])
                ->addColumn('operator_name', 'string', ['limit' => 100, 'default' => '', 'comment' => '操作人姓名'])
                ->addColumn('action', 'string', ['limit' => 50, 'default' => '', 'comment' => '操作动作'])
                ->addColumn('action_desc', 'string', ['limit' => 255, 'default' => '', 'comment' => '操作描述'])
                ->addColumn('old_status', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '操作前状态'])
                ->addColumn('new_status', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '操作后状态'])
                ->addColumn('content', 'text', ['comment' => '操作内容详情'])
                ->addColumn('remark', 'text', ['comment' => '备注'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['purchase_order_uuid'], ['name' => 'idx_purchase_order_uuid'])
                ->addIndex(['operator_uuid'], ['name' => 'idx_operator_uuid'])
                ->addIndex(['action'], ['name' => 'idx_action'])
                ->create();
        }
    }
}
