<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class InitDeskType extends Migrator
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
        $types = [
            ['name' => 'Small table', 'range_min' => 1, 'range_max' => 4, 'create_time' => time(), 'update_time' => time()],
            ['name' => 'Middle table', 'range_min' => 5, 'range_max' => 8, 'create_time' => time(), 'update_time' => time()],
            ['name' => 'large table', 'range_min' => 9, 'range_max' => 12, 'create_time' => time(), 'update_time' => time()],
        ];
        foreach ($types as $type) {
            $data = $db->name('desk_type')->where('name', $type['name'])->find();
            if (!$data) {
                $type['uuid'] = createUuid();
                $db->name('desk_type')->insert($type);
            }
        }
    }
}
