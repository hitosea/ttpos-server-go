<?php

use think\facade\Db;
use think\facade\Env;
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

        $db = Db::connect(Db::getConfig('default'), true);
        // 查询erpnext_site是否存在
        $erpnextSite = $db->name('setting')->where('key', 'erpnext_site')->find();
        $values = '[{"name":"TTPOS","code":"1","headquarter_abbr":""},{"name":"华莱士","code":"2","headquarter_abbr":""},{"name":"UAT","code":"0","headquarter_abbr":""},{"name":"Wallace","code":"4","headquarter_abbr":"CFG"}]';
        if ($erpnextSite) {
            $db->name('setting')->where('key', 'erpnext_site')->update(['values' => $values]);
        }
        // 只处理CFG，默认site_code = 4
        $headquarterAbbr = 'CFG';
        $headquarterSite = '4';
        $sites = json_decode($values, true);
        foreach ($sites as $site) {
            if ($site['headquarter_abbr'] == $headquarterAbbr) {
                $headquarterSite = $site['code'];
                break;
            }
        }
        // 从saas获取总部
        $host = Env::get('DB_HOST');
        $port = Env::get('DB_PORT');
        $username = Env::get('DB_USERNAME');
        $password = Env::get('DB_PASSWORD');
        $dsn = "mysql:host={$host};port={$port}";
        $pdo = new PDO($dsn, $username, $password);
        // 将sql结果映射为数组
        $headquarter = $pdo->query("SELECT * FROM saas.ttpos_company_setting WHERE erpnext_site_code = '{$headquarterSite}' AND erpnext_company_abbr = '{$headquarterAbbr}'")->fetch(PDO::FETCH_ASSOC);
        $headquarterUuid = 0;
        if ($headquarter !== false) {
            $headquarterUuid = $headquarter['uuid'];
        }
        $db->name('company_setting')->where('erpnext_site_code', $headquarterSite)->update(['erpnext_headquarter_abbr' => $headquarterAbbr, 'headquarter_uuid' => $headquarterUuid]);
    }
}
