<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpnextAliasNameErpnextValueNoToProductFlavor extends Migrator
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
        $table = $this->table('product_flavor');
        if (!$table->hasColumn('erpnext_alias_name')) {
            $table->addColumn('erpnext_alias_name', 'string', ['null' => false, 'default' => '', 'comment' => 'ERPNext规格值别名', 'after' => 'erpnext_value_name']);
        }
        if (!$table->hasColumn('erpnext_value_no')) {
            $table->addColumn('erpnext_value_no', 'integer', ['null' => false, 'default' => 0, 'comment' => 'ERPNext规格值编号', 'after' => 'erpnext_alias_name']);
        }
        $table->update();
    }
}
