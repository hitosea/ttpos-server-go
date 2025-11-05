<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreatePurchaseReceiptFileTable extends Migrator
{
    /**
     * 创建收货单附件表
     */
    public function change()
    {
        // 检查表是否已存在
        if (!$this->hasTable('purchase_receipt_file')) {
            $table = $this->table('purchase_receipt_file', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '收货单附件表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '附件关联ID'])
                ->addColumn('receipt_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '收货单UUID'])
                ->addColumn('file_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '文件UUID'])
                ->addColumn('sort_order', 'integer', ['default' => 0, 'comment' => '排序顺序'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'idx_uuid'])
                ->addIndex(['receipt_order_uuid'], ['name' => 'idx_receipt_order_uuid'])
                ->addIndex(['file_uuid'], ['name' => 'idx_file_uuid'])
                ->create();
        }
    }
}

