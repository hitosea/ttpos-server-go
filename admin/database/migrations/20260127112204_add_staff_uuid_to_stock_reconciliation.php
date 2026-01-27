<?php

use think\migration\Migrator;

/**
 * 为盘点单表添加发起人和提交人字段
 *
 * 在 ttpos_stock_reconciliation 表中添加:
 * - creator_staff_uuid: 发起人员工UUID
 * - submitter_staff_uuid: 提交人员工UUID
 */
class AddStaffUuidToStockReconciliation extends Migrator
{
    /**
     * Change Method.
     *
     * 为盘点单表添加发起人和提交人字段
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('stock_reconciliation')) {
            return;
        }

        // 获取表对象
        $table = $this->table('stock_reconciliation');

        // 添加 submitter_staff_uuid 字段（提交人）
        if (!$table->hasColumn('submitter_staff_uuid')) {
            $table->addColumn('submitter_staff_uuid', 'biginteger', [
                'null' => false,
                'default' => 0,
                'comment' => '提交人员工UUID',
                'after' => 'submit_time'
            ]);
        }

        $table->update();
    }
}
