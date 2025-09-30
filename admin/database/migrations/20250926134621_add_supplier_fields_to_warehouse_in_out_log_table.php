<?php
/**
 * 添加supplier_erp_code和supplier_name字段到仓库出入库记录表
 */

use think\migration\Migrator;

class AddSupplierFieldsToWarehouseInOutLogTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     *
     * The following commands can be used in this method and Phinx will
     * automatically reverse them when rolling back:
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('warehouse_in_out_log')) {
            $table = $this->table('warehouse_in_out_log');
            
            // 检查supplier_erp_code字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('supplier_erp_code')) {
                $table->addColumn('supplier_erp_code', 'string', ['limit' => 500, 'default' => '', 'null' => false, 'comment' => '供应商ERP编码', 'after' => 'supplier_uuid']);
            }
            
            // 检查supplier_name字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('supplier_name')) {
                $table->addColumn('supplier_name', 'text', ['null' => true, 'comment' => '供应商名称', 'after' => 'supplier_erp_code']);
            }
            
            $table->update();
        }
    }
}
