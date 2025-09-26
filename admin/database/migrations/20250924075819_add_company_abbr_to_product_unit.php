<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddCompanyAbbrToProductUnit extends Migrator
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
        $table = $this->table('warehouse');
        if (!$table->hasColumn('headquarter_uuid')) {
            $table->addColumn('headquarter_uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '总部Uuid', 'after' => 'erp_code'])
                ->update();
        }
        $table = $this->table('product_unit');
        if (!$table->hasColumn('headquarter_uuid')) {
            $table->addColumn('headquarter_uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '总部Uuid', 'after' => 'erpnext_uom'])
                ->update();
        }
        $table = $this->table('product_sauce');
        if (!$table->hasColumn('headquarter_uuid')) {
            $table->addColumn('headquarter_uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '总部Uuid', 'after' => 'erp_code'])
                ->update();
        }
        $table = $this->table('product_attribute_group');
        if (!$table->hasColumn('headquarter_uuid')) {
            $table->addColumn('headquarter_uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '总部Uuid', 'after' => 'erpnext_attribute_group_name'])
                ->update();
        }
        $table = $this->table('product_attribute');
        if (!$table->hasColumn('headquarter_uuid')) {
            $table->addColumn('headquarter_uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '总部Uuid', 'after' => 'erpnext_attribute_value'])
                ->update();
        }
    }
}
