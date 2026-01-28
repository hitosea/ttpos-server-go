<?php

use think\migration\Migrator;

/**
 * 为调拨单表添加批注字段
 *
 * 在 ttpos_transfer_order 表中添加:
 * - annotations: 批注列表JSON，用于存储审核批注历史
 */
class AddAnnotationsToTransferOrder extends Migrator
{
    const TARGET = 'all';

    /**
     * Change Method.
     *
     * 为调拨单表添加批注字段
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('transfer_order')) {
            return;
        }

        // 获取表对象
        $table = $this->table('transfer_order');

        // 添加 annotations 字段
        if (!$table->hasColumn('annotations')) {
            $table->addColumn('annotations', 'text', [
                'null' => true,
                'comment' => '批注列表JSON',
                'after' => 'remark'
            ]);
        }

        $table->update();
    }
}
