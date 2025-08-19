<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddSpecAttrAccess extends Migrator
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
        // 1704880795 收银机-点餐权限UUID
        // 1704880828 收银机-桌台权限UUID
        // 1724320508 点餐助手权限UUID
        // 1724220610 点餐规格/属性权限UUID
        // 1724220611 桌台规格/属性权限UUID
        // 1724220612 点餐助手规格/属性权限UUID
        $db = Db::connect(Db::getConfig('default'), true);
        // 规格/属性权限
        $accessList = [
            ['uuid' => 1724220610, 'name' => '规格/属性', 'path' => 'cashier_cash_spec_attr', 'parent_uuid' => 1704880795, 'sort' => 12, 'icon' => '', 'redirect_name' => '', 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'plus_category_uuid' => 0, 'remark' => '', 'is_supplier' => 0, 'create_time' => 1752479818, 'update_time' => 1752479818],
            ['uuid' => 1724220611, 'name' => '规格/属性', 'path' => 'cashier_table_spec_attr', 'parent_uuid' => 1704880828, 'sort' => 17, 'icon' => '', 'redirect_name' => '', 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'plus_category_uuid' => 0, 'remark' => '', 'is_supplier' => 0, 'create_time' => 1752479818, 'update_time' => 1752479818],
            ['uuid' => 1724220612, 'name' => '规格/属性', 'path' => 'spec_attr', 'parent_uuid' => 1724320508, 'sort' => 2, 'icon' => '', 'redirect_name' => '', 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'plus_category_uuid' => 0, 'remark' => '', 'is_supplier' => 0, 'create_time' => 1752479818, 'update_time' => 1752479818],
        ];
        foreach ($accessList as $access) {
            $row = $db->name('access')->where('uuid', $access['uuid'])->find();
            if (!$row) {
                $db->name('access')->insert($access);
            }
        }
        // 查找有收银机-点餐权限的角色，并默认选中规格/属性权限
        $this->insertRoleAccess($db, $accessList[0]['parent_uuid'], $accessList[0]['uuid']);
        // 查找有收银机-桌台权限的角色，并默认选中规格/属性权限
        $this->insertRoleAccess($db, $accessList[1]['parent_uuid'], $accessList[1]['uuid']);
        // 查找有点餐助手权限的角色，并默认选中规格/属性权限
        $this->insertRoleAccess($db, $accessList[2]['parent_uuid'], $accessList[2]['uuid']);
    }

    private function insertRoleAccess($db, $parentUuid, $accessUuid)
    {
        $cashRoleList = $db->name('role_access')
            ->where('access_uuid', $parentUuid)
            ->where('delete_time', 0)
            ->distinct(true)
            ->field('role_uuid')
            ->select();
        foreach ($cashRoleList as $cashRole) {
            $row = $db->name('role_access')->where('role_uuid', $cashRole['role_uuid'])->where('access_uuid', $accessUuid)->where('delete_time', 0)->find();
            if (!$row) {
                $uuid = createUuid();
                $db->name('role_access')->insert(['uuid' => $uuid, 'role_uuid' => $cashRole['role_uuid'], 'access_uuid' => $accessUuid, 'create_time' => time(), 'update_time' => time()]);
            }
        }
    }
}
