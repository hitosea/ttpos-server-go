<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddStoreStatisticsAccess extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        // 门店统计权限
        $storeStatisticsAccess = [
            // 门店统计（和用户分析平级，排在用户分析后面）
            ['uuid' => 2856606240768000, 'name' => '门店统计', 'path' => 'store_statistics', 'api_path' => '', 'parent_uuid' => 2856430080000000, 'sort' => 7, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
        ];
        $this->updateOrInsertData('access', 'uuid', $storeStatisticsAccess);

        // 给所有角色赋权限
        $roles = $db->name('role')->where('id', '>', '0')->where('delete_time', '=', '0')->select();
        foreach ($roles as $role) {
            if ($role && isset($role['uuid'])) {
                $roleAccessData = [
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '2856606240768000', 'create_time' => time()],
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

