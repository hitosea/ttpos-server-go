<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddAccess extends Migrator
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
        // 云平台权限
        $adminAccessData = [
            ['id' => 202, 'name' => '外送管理', 'path' => '/delivery/index', 'api_path' => '/admin/delivery/index', 'parent_id' => 101, 'sort' => 21, 'icon' => '', 'redirect_name' => '', 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'remark' => '', 'is_supplier' => 0, 'create_time' => 1750770290, 'update_time' => 1750770290, 'delete_time' => 0,],
            ['id' => 203, 'name' => '外送商家', 'path' => '', 'api_path' => '/admin/delivery/companyList', 'parent_id' => 202, 'sort' => 0, 'icon' => '', 'redirect_name' => '', 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'remark' => '', 'is_supplier' => 0, 'create_time' => 1750770290, 'update_time' => 1750770290, 'delete_time' => 0,],
            ['id' => 204, 'name' => '状态', 'path' => '', 'api_path' => '/admin/shop/updateStatus', 'parent_id' => 203, 'sort' => 0, 'icon' => '', 'redirect_name' => '', 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'remark' => '', 'is_supplier' => 0, 'create_time' => 1750770290, 'update_time' => 1750770290, 'delete_time' => 0,],
            ['id' => 205, 'name' => '保存', 'path' => '', 'api_path' => '/admin/delivery/company', 'parent_id' => 203, 'sort' => 0, 'icon' => '', 'redirect_name' => '', 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'remark' => '', 'is_supplier' => 0, 'create_time' => 1750770290, 'update_time' => 1750770290, 'delete_time' => 0,],
            ['id' => 206, 'name' => '外送台账', 'path' => '', 'api_path' => '/admin/delivery/ledger', 'parent_id' => 202, 'sort' => 0, 'icon' => '', 'redirect_name' => '', 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'remark' => '', 'is_supplier' => 0, 'create_time' => 1750770290, 'update_time' => 1750770290, 'delete_time' => 0,]
        ];
        $this->updateOrInsertData('admin_access', 'id', $adminAccessData);
    }

    /**
     * @param string $tableName 表名
     * @param string|array $uniqueKey 唯一键
     * @param array $data 数据
     */
    private function updateOrInsertData($tableName, $uniqueKey, $data)
    {
        $db = Db::connect(Db::getConfig('default'), true);
        //
        foreach ($data as $item) {
            $query = $db->name($tableName);
            if (is_array($uniqueKey)) {
                foreach ($uniqueKey as $key) {
                    $query->where($key, '=', $item[$key]);
                }
            } else {
                $query->where($uniqueKey, '=', $item[$uniqueKey]);
            }

            $existingData = $query->find();
            if ($existingData) {
                $query->update($item);
            } else {
                $db->name($tableName)->insert($item);
            }
        }
    }
}
