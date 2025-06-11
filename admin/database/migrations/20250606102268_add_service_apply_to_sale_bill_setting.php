<?php

use think\migration\Migrator;

class AddServiceApplyToSaleBillSetting extends Migrator
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
        $table = $this->table('sale_bill_setting');
        if (!$table->hasColumn('service_apply')) {
            $table->addColumn('service_apply', 'integer', ['default' => 0, 'comment' => '是否收取服务费，0-不收取 1-收取。根据后台的服务费应用范围决定', 'after' => 'tax_fee_type']);
            $table->update();
        }
    }
}
