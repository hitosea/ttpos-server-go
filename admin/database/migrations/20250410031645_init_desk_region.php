<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class InitDeskRegion extends Migrator
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
        $db = Db::connect(Db::getConfig('default'), true);
        $region = [
            'name' => 'All regions',
            'create_time' => time(),
            'update_time' => time(),
        ];
        $data = $db->name('desk_region')->where('name', $region['name'])->find();
        if (!$data) {
            $region['uuid'] = createUuid();
            $db->name('desk_region')->insert($region);
        }
    }
}
