<?php
/**
 * 创建原料供应商关联表
 */

use think\migration\Migrator;

class CreateMaterialSupplierTable extends Migrator
{
    /**
     * 执行迁移
     */
    public function up()
    {
        // 检查表是否已存在
        if (!$this->hasTable('material_supplier')) {
            $table = $this->table('material_supplier', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '原料供应商关联表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '主键ID'])
                  ->addColumn('uuid', 'biginteger', ['default' => 0, 'signed' => false, 'comment' => '唯一标识'])
                  ->addColumn('material_uuid', 'biginteger', ['default' => 0, 'signed' => false, 'comment' => '原料UUID'])
                  ->addColumn('material_code', 'string', ['limit' => 100, 'default' => '', 'comment' => '原料编码'])
                  ->addColumn('supplier_uuid', 'biginteger', ['default' => 0, 'signed' => false, 'comment' => '供应商UUID'])
                  ->addColumn('supplier_erp_code', 'string', ['limit' => 100, 'default' => '', 'comment' => '供应商ERP编码'])
                  ->addColumn('headquarter_uuid', 'biginteger', ['default' => 0, 'signed' => false, 'comment' => '总部UUID'])
                  ->addColumn('create_time', 'integer', ['default' => 0, 'signed' => false, 'comment' => '创建时间'])
                  ->addColumn('update_time', 'integer', ['default' => 0, 'signed' => false, 'comment' => '更新时间'])
                  ->addColumn('delete_time', 'integer', ['default' => 0, 'signed' => false, 'comment' => '删除时间'])
                  ->addIndex(['material_uuid'], ['name' => 'idx_material_uuid'])
                  ->addIndex(['supplier_uuid'], ['name' => 'idx_supplier_uuid'])
                  ->addIndex(['headquarter_uuid'], ['name' => 'idx_headquarter_uuid'])
                  ->addIndex(['material_code'], ['name' => 'idx_material_code'])
                  ->addIndex(['supplier_erp_code'], ['name' => 'idx_supplier_erp_code'])
                  ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])
                  ->create();
        }
    }

    /**
     * 回滚迁移
     */
    public function down()
    {
        if ($this->hasTable('material_supplier')) {
            $this->table('material_supplier')->drop()->save();
        }
    }
}
