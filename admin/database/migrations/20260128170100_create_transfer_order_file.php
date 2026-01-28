<?php

use think\migration\Migrator;

/**
 * 创建调拨单附件表
 *
 * 用于存储调拨单收货时上传的附件关联信息，包括：
 * - 签收单、送货单等收货凭证
 * - 支持最多10个附件，单个文件≤20MB
 */
class CreateTransferOrderFile extends Migrator
{
    /**
     * Change Method.
     */
    public function change()
    {
        // 检查表是否已存在
        if ($this->hasTable('transfer_order_file')) {
            return;
        }

        // 创建调拨单附件表
        $table = $this->table('transfer_order_file', ['comment' => '调拨单附件表']);
        $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '附件关联ID'])
            ->addColumn('transfer_order_uuid', 'biginteger', ['default' => 0, 'comment' => '调拨单UUID'])
            ->addColumn('file_uuid', 'biginteger', ['default' => 0, 'comment' => '文件UUID'])
            ->addColumn('sort_order', 'integer', ['default' => 0, 'comment' => '排序顺序'])
            ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
            ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
            ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
            ->addIndex(['uuid'], ['unique' => true])
            ->addIndex(['transfer_order_uuid'])
            ->addIndex(['file_uuid'])
            ->create();
    }
}
