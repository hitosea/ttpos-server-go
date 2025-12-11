<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddStaffAccess extends Migrator
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
        $currentTime = time();
        
        // 云平台权限 - 人员管理（统一账号管理）
        $adminAccessData = [
            // 人员管理（主权限，parent_id = 9 是用户管理）
            [
                'id' => 209,
                'name' => '人员管理',
                'path' => '/user/staff',
                'api_path' => '/admin/admin.staff/index',
                'parent_id' => 9,
                'sort' => 5,
                'icon' => '',
                'redirect_name' => '',
                'is_route' => 1,
                'is_menu' => 1,
                'is_show' => 1,
                'remark' => '管理所有使用TTPOS侧的用户（统一账号管理）',
                'is_supplier' => 0,
                'create_time' => $currentTime,
                'update_time' => $currentTime,
                'delete_time' => 0,
            ],
            // 列表
            [
                'id' => 210,
                'name' => '列表',
                'path' => '',
                'api_path' => '/admin/admin.staff/index',
                'parent_id' => 209,
                'sort' => 1,
                'icon' => '',
                'redirect_name' => '',
                'is_route' => 0,
                'is_menu' => 0,
                'is_show' => 1,
                'remark' => '',
                'is_supplier' => 0,
                'create_time' => $currentTime,
                'update_time' => $currentTime,
                'delete_time' => 0,
            ],
            // 添加
            [
                'id' => 211,
                'name' => '添加',
                'path' => '',
                'api_path' => '/admin/admin.staff/add',
                'parent_id' => 209,
                'sort' => 2,
                'icon' => '',
                'redirect_name' => '',
                'is_route' => 0,
                'is_menu' => 0,
                'is_show' => 1,
                'remark' => '',
                'is_supplier' => 0,
                'create_time' => $currentTime,
                'update_time' => $currentTime,
                'delete_time' => 0,
            ],
            // 编辑
            [
                'id' => 212,
                'name' => '编辑',
                'path' => '',
                'api_path' => '/admin/admin.staff/edit',
                'parent_id' => 209,
                'sort' => 3,
                'icon' => '',
                'redirect_name' => '',
                'is_route' => 0,
                'is_menu' => 0,
                'is_show' => 1,
                'remark' => '',
                'is_supplier' => 0,
                'create_time' => $currentTime,
                'update_time' => $currentTime,
                'delete_time' => 0,
            ],
            // 启用禁用状态
            [
                'id' => 213,
                'name' => '启用禁用状态',
                'path' => '',
                'api_path' => '/admin/admin.staff/updateStatus',
                'parent_id' => 209,
                'sort' => 4,
                'icon' => '',
                'redirect_name' => '',
                'is_route' => 0,
                'is_menu' => 0,
                'is_show' => 1,
                'remark' => '',
                'is_supplier' => 0,
                'create_time' => $currentTime,
                'update_time' => $currentTime,
                'delete_time' => 0,
            ],
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
