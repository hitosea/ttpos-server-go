<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddStockLossAccess extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        // 更新参数设置的 sort 值（报损管理排在参数设置前面）
        $db->name('access')->where('uuid', '=', '2858908913663000')->update(['sort' => 15]); // 参数设置 14 -> 15

        // 报损管理权限（在进销存下，排在参数设置前面）
        $stockLossAccess = [
            // 报损管理
            ['uuid' => 2860100000000000, 'name' => '报损管理', 'path' => 'stock_loss_management', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 14, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            // 添加
            ['uuid' => 2860100000001000, 'name' => '添加', 'path' => 'stock_loss_add', 'api_path' => '', 'parent_uuid' => 2860100000000000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            // 删除
            ['uuid' => 2860100000003000, 'name' => '删除', 'path' => 'stock_loss_delete', 'api_path' => '', 'parent_uuid' => 2860100000000000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            // 审核
            ['uuid' => 2860100000004000, 'name' => '审核', 'path' => 'stock_loss_approve', 'api_path' => '', 'parent_uuid' => 2860100000000000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
        ];
        $this->updateOrInsertData('access', 'uuid', $stockLossAccess);

        // 给所有角色赋权限
        $roles = $db->name('role')->where('id', '>', '0')->where('delete_time', '=', '0')->select();
        foreach ($roles as $role) {
            if ($role && isset($role['uuid'])) {
                $roleAccessData = [
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '2860100000000000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '2860100000001000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '2860100000003000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '2860100000004000', 'create_time' => time()],
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
