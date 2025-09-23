<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddErpnextHeadquarterAbbrToCompanySetting extends Migrator
{
    const TARGET = 'all';
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
        $table = $this->table('company_setting');
        if (!$table->hasColumn('erpnext_headquarter_abbr')) {
            $table->addColumn('erpnext_headquarter_abbr', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => 'ERPNext总部简称', 'after' => 'erpnext_company_abbr']);
        }
        if (!$table->hasColumn('headquarter_uuid')) {
            $table->addColumn('headquarter_uuid', 'biginteger', ['null' => false, 'default' => 0, 'comment' => '总部Uuid', 'after' => 'erpnext_headquarter_abbr']);
        }

        $table->update();

        // TODO 上线前，需要修改saas.ttpos_setting中key=erpnext_site的配置
        // 指定总部后，修改 saas.ttpos_company_setting 和商家 ttpos_company_setting 的 erpnext_headquarter_abbr 和 headquarter_uuid 的值
    }
}
