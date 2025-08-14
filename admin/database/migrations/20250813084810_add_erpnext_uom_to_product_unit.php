<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpnextUomToProductUnit extends Migrator
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
        $table = $this->table('product_unit');
        $hasColumn = $table->hasColumn('erpnext_uom');
        if (!$hasColumn) {
            $table->addColumn('erpnext_uom', 'string', ['limit' => 255, 'null' => true, 'default' => null, 'comment' => 'ERPNext UOM', 'after' => 'sort'])->update();
        }

        $table = $this->table('product_attribute_group');
        $hasColumn = $table->hasColumn('erpnext_attribute_group_name');
        if (!$hasColumn) {
            $table->addColumn('erpnext_attribute_group_name', 'string', ['limit' => 255, 'null' => true, 'default' => null, 'comment' => 'ERPNext Attribute Group Name', 'after' => 'sort'])->update();
        }

        $table = $this->table('product_attribute');
        $hasColumn = $table->hasColumn('erpnext_attribute_value');
        if (!$hasColumn) {
            $table->addColumn('erpnext_attribute_value', 'string', ['limit' => 255, 'null' => true, 'default' => null, 'comment' => 'ERPNext Attribute Name', 'after' => 'sort'])->update();
        }
    }
}
