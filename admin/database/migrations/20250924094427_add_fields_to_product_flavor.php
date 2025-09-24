<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddFieldsToProductFlavor extends Migrator
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
        if (!$table->hasColumn('erpnext_company_abbr')) {
            $table->addColumn('erpnext_company_abbr', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => 'ERPNext公司简称', 'after' => 'sort']);
        }
        if (!$table->hasColumn('erpnext_group_name')) {
            $table->addColumn('erpnext_group_name', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => 'ERPNext规格组名称', 'after' => 'erpnext_company_abbr']);
        }
        if (!$table->hasColumn('erpnext_value_name')) {
            $table->addColumn('erpnext_value_name', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => 'ERPNext规格值名称', 'after' => 'erpnext_group_name']);
        }
        $table->update();
    }
}
