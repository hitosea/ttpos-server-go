<?php

use think\migration\Migrator;

class AddErpRespToTransferOrder extends Migrator
{
    // 迁移目标
    const TARGET = 'all';
    
    /**
     * 添加调拨单ERP响应字段
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('transfer_order')) {
            return;
        }

        $table = $this->table('transfer_order');

        // 检查字段是否存在，不存在则添加
        if (!$table->hasColumn('erp_resp')) {
            $table->addColumn('erp_resp', 'text', ['comment' => 'ERP响应数据', 'after' => 'item_count'])->update();
        }
    }
}

