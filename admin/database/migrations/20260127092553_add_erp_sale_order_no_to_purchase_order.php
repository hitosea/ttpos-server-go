<?php

use think\migration\Migrator;

/**
 * 为采购单表添加 ERP 销售单号字段
 *
 * 在 ttpos_purchase_order 表中添加：
 * - erp_sale_order_no: ERP销售单号
 */
class AddErpSaleOrderNoToPurchaseOrder extends Migrator
{
    /**
     * Change Method.
     *
     * 为采购单表添加 ERP 销售单号字段
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('purchase_order')) {
            return;
        }

        // 获取表对象
        $table = $this->table('purchase_order');

        // 添加 erp_sale_order_no 字段 - ERP销售单号
        if (!$table->hasColumn('erp_sale_order_no')) {
            $table->addColumn('erp_sale_order_no', 'string', [
                'limit' => 255,
                'null' => false,
                'default' => '',
                'comment' => 'ERP销售单号',
                'after' => 'erp_order_no'
            ]);
        }

        $table->update();
    }
}
