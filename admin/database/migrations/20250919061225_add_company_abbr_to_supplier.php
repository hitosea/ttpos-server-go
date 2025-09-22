<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddCompanyAbbrToSupplier extends Migrator
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
        // TODO 上线前需要删除
        $table = $this->table('supplier');
        if (!$table->hasColumn('company_abbr')) {
            $table->addColumn('company_abbr', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '所属公司简称，如果为空表示来自总部', 'after' => 'erp_code']);
        }
        $table->update();
    }
}
