<?php

use think\migration\Migrator;

/**
 * 为物品表添加供应商配送相关字段
 *
 * 在 ttpos_material 表中添加：
 * - delivered_by_supplier: 是否由供应商配送，0-否，1-是
 * - supplier_erp_code: 供应商ERP编码
 */
class AddSupplierFieldsToMaterialTable extends Migrator
{
    /**
     * Change Method.
     *
     * 为物品表添加供应商配送相关字段
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('material')) {
            return;
        }

        // 获取表对象
        $table = $this->table('material');

        // 添加 delivered_by_supplier 字段 - 是否由供应商配送
        if (!$table->hasColumn('delivered_by_supplier')) {
            $table->addColumn('delivered_by_supplier', 'integer', [
                'limit' => 1,
                'null' => false,
                'default' => 0,
                'comment' => '是否由供应商配送，0-否，1-是',
                'after' => 'allow_negative_stock'
            ]);
        }

        // 添加 supplier_erp_code 字段 - 供应商ERP编码
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
