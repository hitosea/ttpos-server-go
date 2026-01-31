<?php

use think\migration\Migrator;

/**
 * 创建盘点单批注表
 *
 * 用于记录盘点单审核过程中的批注信息，包括：
 * - 重新发起：发起人重新提交审核时的说明
 * - 驳回：审核人驳回时的说明
 * - 通过：审核人通过时的说明
 */
class CreateStockReconciliationAnnotation extends Migrator
{
    /**
     * Change Method.
     */
    public function change()
    {
        // 检查表是否已存在
        if ($this->hasTable('stock_reconciliation_annotation')) {
            return;
        }

        // 创建盘点单批注表
        $table = $this->table('stock_reconciliation_annotation', ['comment' => '盘点单批注表']);
        $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
            ->addColumn('stock_reconciliation_uuid', 'biginteger', ['default' => 0, 'comment' => '盘点单UUID'])
            ->addColumn('annotation_type', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '批注类型 1-重新发起 2-驳回 3-通过'])
            ->addColumn('content', 'text', ['null' => true, 'comment' => '批注内容'])
            ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
            ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
            ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
            ->addIndex(['uuid'], ['unique' => true])
            ->addIndex(['stock_reconciliation_uuid'])
            ->create();
    }
}
