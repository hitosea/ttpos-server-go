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

        // 华莱士site
        $db = Db::connect(Db::getConfig('default'), true);
        // TODO 要先有总部才行
        // $headquarter = $db->name('company_setting')->where('erpnext_site_code', '2')->where('erpnext_company_abbr', 'CFG')->first();
        // $headquarter['uuid']
        $headquarterUuid = 0;
        $db->name('company_setting')->where('erpnext_site_code', '2')->update(['erpnext_headquarter_abbr' => 'CFG', 'headquarter_uuid' => $headquarterUuid]);

        // ttpos site (散户) 设置erpnext_headquarter_abbr为company_abbr的值
        $companies = $db->name('company_setting')->where('erpnext_site_code', '1')->select();
        foreach ($companies as $company) {
            $db->name('company_setting')->where('id', $company['id'])->update(['erpnext_headquarter_abbr' => $company['erpnext_company_abbr'], 'headquarter_uuid' => $company['uuid']]);
        }
    }
}
