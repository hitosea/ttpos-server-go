<?php

use think\migration\Migrator;

class AddReceiptOrderErpFieldsToTransferOrder extends Migrator
{
    // 迁移目标
    const TARGET = 'all';
    
    /**
     * 添加调拨单收货单ERP相关字段
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('transfer_order')) {
            return;
        }

        $table = $this->table('transfer_order');

        // 检查字段是否存在，不存在则添加
        if (!$table->hasColumn('receipt_order_erp_code')) {
            $table->addColumn('receipt_order_erp_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '收货单ERP编码', 'after' => 'erp_resp'])->update();
        }

        if (!$table->hasColumn('receipt_order_erp_resp')) {
            $table->addColumn('receipt_order_erp_resp', 'text', ['comment' => '收货单ERP响应数据', 'after' => 'receipt_order_erp_code'])->update();
        }
    }
}

