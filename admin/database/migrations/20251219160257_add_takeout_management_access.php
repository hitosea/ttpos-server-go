<?php

use think\facade\Db;
use think\migration\Migrator;

class AddTakeoutManagementAccess extends Migrator
{
    /**
     * 新增外卖管理权限和打印设置权限
     * 1. 商品管理下的批量操作权限（Grab和LINE MAN各4个）
     * 2. 外卖管理一级菜单（工作台下）
     * 3. Grab和LINE MAN二级菜单
     * 4. 打印设置权限（餐厅设置下）
     * 5. 调整"餐厅设置"和"其他"的sort值以避免冲突
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        // 1. 商品管理下的批量操作权限（8个）
        $productBatchAccessData = [
            // Grab批量操作（4个）
            ['uuid' => 2857076002816000, 'name' => '批量创建Grab', 'path' => 'product_batch_create_grab', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 6, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2857096974336000, 'name' => '批量上架Grab', 'path' => 'product_batch_online_grab', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 7, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2857117945856000, 'name' => '批量下架Grab', 'path' => 'product_batch_offline_grab', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 8, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2857138917376000, 'name' => '批量删除Grab', 'path' => 'product_batch_delete_grab', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 9, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            
            // LINE MAN批量操作（4个）
            ['uuid' => 2857159888896001, 'name' => '批量创建LINE MAN', 'path' => 'product_batch_create_lineman', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 10, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2857180860416001, 'name' => '批量上架LINE MAN', 'path' => 'product_batch_online_lineman', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 11, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2857201831936000, 'name' => '批量下架LINE MAN', 'path' => 'product_batch_offline_lineman', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 12, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2857222803456001, 'name' => '批量删除LINE MAN', 'path' => 'product_batch_delete_lineman', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 13, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
        ];
        $this->updateOrInsertData('access', 'uuid', $productBatchAccessData);

        // 2. 外卖管理菜单（3个：1个一级菜单 + 2个二级菜单）
        $takeoutManagementAccessData = [
            // 外卖管理（工作台下的一级菜单，sort=5在营销管理和餐厅设置之间）
            ['uuid' => 2858986936320000, 'name' => '外卖管理', 'path' => 'takeout_management', 'api_path' => '', 'parent_uuid' => 2856757235712000, 'sort' => 5, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            
            // Grab（外卖管理下的二级菜单）
            ['uuid' => 2859007907840000, 'name' => 'Grab', 'path' => 'takeout_grab', 'api_path' => '', 'parent_uuid' => 2858986936320000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            
            // LINE MAN（外卖管理下的二级菜单）
            ['uuid' => 2859028879360000, 'name' => 'LINE MAN', 'path' => 'takeout_lineman', 'api_path' => '', 'parent_uuid' => 2858986936320000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
        ];
        $this->updateOrInsertData('access', 'uuid', $takeoutManagementAccessData);

        // 3. 打印设置权限（餐厅设置下的二级菜单）
        $printerSettingAccessData = [
            // 打印设置（餐厅设置下，sort=4在桌台地图之后）
            ['uuid' => 2859123823104000, 'name' => '打印设置', 'path' => 'printer_setting', 'api_path' => '', 'parent_uuid' => 2859064102912000, 'sort' => 4, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
        ];
        $this->updateOrInsertData('access', 'uuid', $printerSettingAccessData);

        // 4. 调整"餐厅设置"和"其他"的sort值，避免冲突
        // 餐厅设置从 sort=5 调整为 sort=6
        $db->name('access')->where('uuid', 2859064102912000)->update(['sort' => 6, 'update_time' => time()]);
        
        // 其他从 sort=6 调整为 sort=7
        $db->name('access')->where('uuid', 2859273818112000)->update(['sort' => 7, 'update_time' => time()]);

        // 5. 为所有角色分配新权限
        $roles = $db->name('role')->where('id', '>', '0')->where('delete_time', '=', '0')->select();
        
        // 所有新增权限的UUID列表
        $newAccessUuids = [
            // 商品批量操作
            2857076002816000, 2857096974336000, 2857117945856000, 2857138917376000,
            2857159888896001, 2857180860416001, 2857201831936000, 2857222803456001,
            // 外卖管理菜单
            2858986936320000, 2859007907840000, 2859028879360000,
            // 打印设置
            2859123823104000
        ];
        
        foreach ($roles as $role) {
            if ($role && isset($role['uuid'])) {
                $roleAccessData = [];
                foreach ($newAccessUuids as $accessUuid) {
                    $roleAccessData[] = [
                        'uuid' => createUuid(),
                        'role_uuid' => $role['uuid'],
                        'access_uuid' => $accessUuid,
                        'create_time' => time()
                    ];
                }
                $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], $roleAccessData);
            }
        }
    }

    /**
     * 更新或插入数据（幂等操作）
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
                // 已存在则不更新，避免覆盖
            } else {
                $db->name($tableName)->insert($item);
            }
        }
    }
}
