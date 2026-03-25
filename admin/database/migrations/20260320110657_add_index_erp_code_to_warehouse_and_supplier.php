<?php
/**
 * 为仓库表和供应商表的 erp_code 字段添加索引
 *
 * 慢查询：SELECT * FROM ttpos_warehouse WHERE erp_code = ? AND delete_time = 0 LIMIT 1
 * 慢查询：SELECT * FROM ttpos_supplier WHERE erp_code = ? AND delete_time = 0 LIMIT 1
 * 原因：erp_code 字段无索引，每次查询全表扫描
 * 影响：采购单创建、ERP审批、自动收货等 30+ 处调用
 */

use think\migration\Migrator;

class AddIndexErpCodeToWarehouseAndSupplier extends Migrator
{
    public function change()
    {
        // 仓库表：添加 erp_code 索引
        if ($this->hasTable('warehouse')) {
            $table = $this->table('warehouse');

            if (!$table->hasIndex('erp_code')) {
                $table->addIndex('erp_code', ['name' => 'idx_erp_code']);
            }

            $table->update();
        }

        // 供应商表：添加 erp_code 索引
        if ($this->hasTable('supplier')) {
            $table = $this->table('supplier');

            if (!$table->hasIndex('erp_code')) {
                $table->addIndex('erp_code', ['name' => 'idx_erp_code']);
            }

            $table->update();
        }
    }
}
