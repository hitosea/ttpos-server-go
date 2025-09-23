<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddWarehouseErpCodeToPurchaseReceiptOrderTable extends Migrator
{
    /**
     * 添加 source_warehouse_erp_code 和 target_warehouse_erp_code 字段到 ttpos_purchase_receipt_order 表
     */
    public function change()
    {
        $table = $this->table('purchase_receipt_order');

        // 检查字段是否已存在
        if (!$table->hasColumn('source_warehouse_erp_code')) {
            $table->addColumn('source_warehouse_erp_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '源仓库ERP编码', 'after' => 'purchase_time'])
                  ->update();
        }

        if (!$table->hasColumn('target_warehouse_erp_code')) {
            $table->addColumn('target_warehouse_erp_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '目标仓库ERP编码', 'after' => 'source_warehouse_erp_code'])
                  ->update();
        }
    }
}
