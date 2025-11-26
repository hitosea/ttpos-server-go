<?php

use think\facade\Db;
use think\migration\Migrator;

class AddTakeMealSettingPermission extends Migrator
{
    /**
     * 添加点餐设置权限
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        // 点餐设置权限数据
        $accessData = [
            ['uuid' => 2859479339008001, 'name' => '点餐设置', 'path' => 'take_meal_setting', 'api_path' => '', 'parent_uuid' => 2859345121280000, 'sort' => 8, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
        ];
        $this->updateOrInsertData('access', 'uuid', $accessData);

        // 为所有角色添加此权限
        $roles = $db->name('role')->where('id', '>', '0')->where('delete_time', '=', '0')->select();
        foreach ($roles as $role) {
            if ($role && isset($role['uuid'])) {
                $roleAccessData = [
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '2859479339008001', 'create_time' => time()],
                ];
                $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], $roleAccessData);
            }
        }
    }

    /**
     * 更新或插入数据
     *
     * @param string $tableName 表名
     * @param string|array $uniqueKey 唯一键
     * @param array $data 数据
     */
    private function updateOrInsertData($tableName, $uniqueKey, $data)
    {
        $db = Db::connect(Db::getConfig('default'), true);
        
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
                // 已存在，不做更新
            } else {
                $db->name($tableName)->insert($item);
            }
        }
    }
}

