<?php
/**
 * 在采购物品相关表中添加erpnext_uom和base_erpnext_uom字段
 */

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpnextUomFieldsToPurchaseItemTables extends Migrator
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
        // 给ttpos_purchase_order_item表添加erpnext_uom和base_erpnext_uom字段
        if ($this->hasTable('purchase_order_item')) {
            $table = $this->table('purchase_order_item');
            
            // 检查字段是否不存在
            if (!$table->hasColumn('erpnext_uom')) {
                $table->addColumn('erpnext_uom', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERPNext单位', 'after' => 'material_uuid']);
                $table->update();
            }
            
            if (!$table->hasColumn('base_erpnext_uom')) {
                $table->addColumn('base_erpnext_uom', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERPNext基准单位', 'after' => 'erpnext_uom']);
                $table->update();
            }
        }

        // 给ttpos_purchase_receipt_order_item表添加erpnext_uom和base_erpnext_uom字段
        if ($this->hasTable('purchase_receipt_order_item')) {
            $table = $this->table('purchase_receipt_order_item');
            
            // 检查字段是否不存在
            if (!$table->hasColumn('erpnext_uom')) {
                $table->addColumn('erpnext_uom', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERPNext单位', 'after' => 'material_uuid']);
                $table->update();
            }
            
            if (!$table->hasColumn('base_erpnext_uom')) {
                $table->addColumn('base_erpnext_uom', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERPNext基准单位', 'after' => 'erpnext_uom']);
                $table->update();
            }
        }
    }
}
