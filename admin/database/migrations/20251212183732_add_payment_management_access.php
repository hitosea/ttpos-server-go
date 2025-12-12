<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddPaymentManagementAccess extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        // 支付管理权限
        $paymentManagementAccess = [
            // 支付管理（和各端设置平级，排在各端设置后面）
            ['uuid' => 2859373891584000, 'name' => '支付管理', 'path' => 'payment_management', 'api_path' => '', 'parent_uuid' => 2859273818112000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            // 支付配置
            ['uuid' => 2859394666496000, 'name' => '支付配置', 'path' => 'payment_config', 'api_path' => '', 'parent_uuid' => 2859373891584000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            // 添加
            ['uuid' => 2859415441408000, 'name' => '添加', 'path' => 'payment_add', 'api_path' => '', 'parent_uuid' => 2859373891584000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            // 编辑
            ['uuid' => 2859436216320000, 'name' => '编辑', 'path' => 'payment_edit', 'api_path' => '', 'parent_uuid' => 2859373891584000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            // 排序
            ['uuid' => 2859456991232000, 'name' => '排序', 'path' => 'payment_sort', 'api_path' => '', 'parent_uuid' => 2859373891584000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            // 删除
            ['uuid' => 2859477766144000, 'name' => '删除', 'path' => 'payment_delete', 'api_path' => '', 'parent_uuid' => 2859373891584000, 'sort' => 5, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
        ];
        $this->updateOrInsertData('access', 'uuid', $paymentManagementAccess);

        // 给所有角色赋权限
        $roles = $db->name('role')->where('id', '>', '0')->where('delete_time', '=', '0')->select();
        foreach ($roles as $role) {
            if ($role && isset($role['uuid'])) {
                $roleAccessData = [
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '2859373891584000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '2859394666496000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '2859415441408000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '2859436216320000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '2859456991232000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $role['uuid'], 'access_uuid' => '2859477766144000', 'create_time' => time()],
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

