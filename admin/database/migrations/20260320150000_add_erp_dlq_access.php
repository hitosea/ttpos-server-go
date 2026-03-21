<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddErpDlqAccess extends Migrator
{
    // 迁移目标：仅 SAAS 主库
    const TARGET = 'main';

    /**
     * Change Method.
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        $adminAccessData = [
            // 父菜单：ERP 死信管理（挂在平台管理 parent_id=101 下）
            ['id' => 215, 'name' => 'ERP 死信管理', 'path' => '/erp-dlq/index', 'api_path' => '/admin/erpnext/siFailedStats', 'parent_id' => 101, 'sort' => 10, 'icon' => '', 'redirect_name' => '', 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'remark' => 'Sales Invoice 失败记录管理', 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time(), 'delete_time' => 0],
            // 子操作：重试
            ['id' => 216, 'name' => '重试失败记录', 'path' => '', 'api_path' => '/admin/erpnext/siRetry', 'parent_id' => 215, 'sort' => 0, 'icon' => '', 'redirect_name' => '', 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'remark' => '', 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time(), 'delete_time' => 0],
        ];
        $this->updateOrInsertData('admin_access', 'id', $adminAccessData);

        // 给所有现有角色授权
        $roles = $db->name('admin_role')->where('delete_time', '=', 0)->select();
        foreach ($roles as $role) {
            $roleAccessData = [
                ['role_id' => $role['id'], 'access_id' => 215, 'create_time' => time(), 'update_time' => time(), 'delete_time' => 0],
                ['role_id' => $role['id'], 'access_id' => 216, 'create_time' => time(), 'update_time' => time(), 'delete_time' => 0],
            ];
            $this->updateOrInsertData('admin_role_access', ['role_id', 'access_id'], $roleAccessData);
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
