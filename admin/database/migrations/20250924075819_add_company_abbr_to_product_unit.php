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
        if (!$table->hasColumn('company_abbr')) {
            $table->addColumn('company_abbr',  'string', ['limit' => 255, 'default' => '', 'comment' => '所属公司简称，如果是自己的company_abbr表示自己创建，其他值表示非自己创建', 'after' => 'erp_code'])
                ->update();
        }
        $table = $this->table('product_unit');
        if (!$table->hasColumn('company_abbr')) {
            $table->addColumn('company_abbr',  'string', ['limit' => 255, 'default' => '', 'comment' => '所属公司简称，如果是自己的company_abbr表示自己创建，其他值表示非自己创建', 'after' => 'erpnext_uom'])
                ->update();
        }
        $table = $this->table('product_sauce');
        if (!$table->hasColumn('company_abbr')) {
            $table->addColumn('company_abbr',  'string', ['limit' => 255, 'default' => '', 'comment' => '所属公司简称，如果是自己的company_abbr表示自己创建，其他值表示非自己创建', 'after' => 'erp_code'])
                ->update();
        }
        $table = $this->table('product_attribute_group');
        if (!$table->hasColumn('company_abbr')) {
            $table->addColumn('company_abbr',  'string', ['limit' => 255, 'default' => '', 'comment' => '所属公司简称，如果是自己的company_abbr表示自己创建，其他值表示非自己创建', 'after' => 'erpnext_attribute_group_name'])
                ->update();
        }
        $table = $this->table('product_attribute');
        if (!$table->hasColumn('company_abbr')) {
            $table->addColumn('company_abbr',  'string', ['limit' => 255, 'default' => '', 'comment' => '所属公司简称，如果是自己的company_abbr表示自己创建，其他值表示非自己创建', 'after' => 'erpnext_attribute_value'])
                ->update();
        }
    }
}
