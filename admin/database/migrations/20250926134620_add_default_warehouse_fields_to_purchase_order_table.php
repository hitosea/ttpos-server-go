<?php
/**
 * 添加DefaultWarehouseErpCode和DefaultWarehouseName字段到采购申请表
 */

use think\migration\Migrator;

class AddDefaultWarehouseFieldsToPurchaseOrderTable extends Migrator
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
        if ($this->hasTable('purchase_order')) {
            $table = $this->table('purchase_order');
            
            // 检查DefaultWarehouseErpCode字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('default_warehouse_erp_code')) {
                $table->addColumn('default_warehouse_erp_code', 'string', ['limit' => 500, 'default' => '', 'null' => false, 'comment' => '默认仓库ERP编码', 'after' => 'headquarter_status']);
            }
            
            // 检查DefaultWarehouseName字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('default_warehouse_name')) {
                $table->addColumn('default_warehouse_name', 'text', ['null' => true, 'comment' => '默认仓库名称', 'after' => 'default_warehouse_erp_code']);
            }
            
            $table->update();
        }
    }
}
