<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddErpnextSetting extends Migrator
{

    // 迁移目标
    const TARGET = 'main';

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
        $db = Db::connect(Db::getConfig('default'), true);
        // 查询erpnext_site是否存在
        $erpnextSite = $db->name('setting')->where('key', 'erpnext_site')->find();
        if (!$erpnextSite) {
            $db->name('setting')->insert([
                'key' => 'erpnext_site',
                'describe' => 'erpnext站点',
                'values' => '[{"name":"TTPOS","code":"1"},{"name":"华莱士","code":"2"},{"name":"UAT","code":"0"}]',
                'create_time' => time(),
                'update_time' => time(),
                'delete_time' => 0,
            ]);
        }
    }
}
