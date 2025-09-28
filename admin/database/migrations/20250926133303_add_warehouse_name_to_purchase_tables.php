<?php
/**
 * 在采购相关表中添加warehouse_name字段
 */

use think\migration\Migrator;
use think\migration\db\Column;

class AddWarehouseNameToPurchaseTables extends Migrator
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
        // 给ttpos_purchase_order表添加warehouse_name字段
        if ($this->hasTable('purchase_order')) {
            $table = $this->table('purchase_order');
            
            // 检查字段是否不存在
            if (!$table->hasColumn('warehouse_name')) {
                $table->addColumn('warehouse_name', 'text', ['comment' => '仓库名称', 'after' => 'warehouse_erp_code']);
                $table->update();
            }
        }

        // 给ttpos_purchase_receipt_order表添加warehouse_name字段
        if ($this->hasTable('purchase_receipt_order')) {
            $table = $this->table('purchase_receipt_order');
            
            // 检查字段是否不存在
            if (!$table->hasColumn('source_warehouse_name')) {
                $table->addColumn('source_warehouse_name', 'text', ['comment' => '仓库名称', 'after' => 'source_warehouse_erp_code']);
                $table->update();
            }

            // 检查字段是否不存在
            if (!$table->hasColumn('target_warehouse_name')) {
                $table->addColumn('target_warehouse_name', 'text', ['comment' => '仓库名称', 'after' => 'target_warehouse_erp_code']);
                $table->update();
            }
        }
    }
}
