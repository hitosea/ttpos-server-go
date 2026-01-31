<?php

use think\migration\Migrator;

class AddSupplierFieldsToPurchaseOrderItem extends Migrator
{
    /**
     * 为 ttpos_purchase_order_item 添加供应商配送相关字段
     * - delivered_by_supplier: 是否由供应商配送
     * - supplier_erp_code: 供应商ERP编码
     */
    public function change()
    {
        $table = $this->table('purchase_order_item');
        if (!$table->hasColumn('delivered_by_supplier')) {
            $table->addColumn('delivered_by_supplier', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '是否由供应商配送，0-否，1-是',
                'after' => 'base_erpnext_uom'
            ]);
        }
        if (!$table->hasColumn('supplier_erp_code')) {
            $table->addColumn('supplier_erp_code', 'string', [
                'limit' => 255,
                'null' => false,
                'default' => '',
                'comment' => '供应商ERP编码',
                'after' => 'delivered_by_supplier'
            ]);
        }
        $table->update();
    }
}
