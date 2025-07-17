<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddOverallDiscountAccess extends Migrator
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
        // 1724220609 整单折扣权限UUID
        $db = Db::connect(Db::getConfig('default'), true);
        $access = ['id' => 1724220609, 'uuid' => 1724220609, 'name' => '整单折扣', 'path' => '/product/buffet/list/overallDiscount', 'api_path' => '/product/buffet/buffet/overallDiscount', 'parent_uuid' => 1708671752, 'sort' => 1, 'icon' => '', 'redirect_name' => '', 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'plus_category_uuid' => 0, 'remark' => '', 'is_supplier' => 0, 'create_time' => 1752479818, 'update_time' => 1752479818];
        $row = $db->name('access')->where('uuid', $access['id'])->find();
        if (!$row) {
            $db->name('access')->insert($access);
        }
        $roleList = $db->name('role_access')
            ->where('access_uuid', 1708671752)
            ->where('delete_time', 0)
            ->distinct(true)
            ->field('role_uuid')
            ->select();
        foreach ($roleList as $role) {
            $row = $db->name('role_access')->where('role_uuid', $role['role_uuid'])->where('access_uuid', 1724220609)->where('delete_time', 0)->find();
            if (!$row) {
                $uuid = createUuid();
                $db->name('role_access')->insert(['uuid' => $uuid, 'role_uuid' => $role['role_uuid'], 'access_uuid' => 1724220609, 'create_time' => time(), 'update_time' => time()]);
            }
        }
    }
}
