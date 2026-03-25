<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddMobileOrderAccess extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        // 手机点餐权限（在餐厅设置下，排在打印设置后面）
        $mobileOrderAccess = [
            // 手机点餐（二维码、菜单样式）
            ['uuid' => 1724220614, 'name' => '手机点餐', 'path' => 'mobile_order_setting', 'api_path' => '', 'parent_uuid' => 2859064102912000, 'sort' => 5, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            // 二维码
            ['uuid' => 1724220615, 'name' => '二维码', 'path' => 'mobile_order_qrcode', 'api_path' => '', 'parent_uuid' => 1724220614, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            // 菜单样式
            ['uuid' => 1724220616, 'name' => '菜单样式', 'path' => 'mobile_order_menu_style', 'api_path' => '', 'parent_uuid' => 1724220614, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
        ];
        $this->updateOrInsertData('access', 'uuid', $mobileOrderAccess);

        // 给所有角色赋权限
        $roles = $db->name('role')->where('id', '>', '0')->where('delete_time', '=', '0')->select();
        foreach ($roles as $role) {
            if ($role && isset($role['uuid'])) {
                $roleAccessData = [
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '1724220614', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '1724220615', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '1724220616', 'create_time' => time()],
                ];
                $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], $roleAccessData);
            }
        }
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
                // 已存在则不更新，避免覆盖
            } else {
                $db->name($tableName)->insert($item);
            }
        }
    }
}
