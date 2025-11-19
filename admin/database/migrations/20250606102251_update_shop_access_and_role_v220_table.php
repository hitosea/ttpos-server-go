<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class UpdateShopAccessAndRoleV220Table extends Migrator
{
    /**
     * 更新或插入数据
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
        $db = Db::connect(Db::getConfig('default'), true);
        $db->name('migrations')->where('migration_name', 'UpdateShopAccessAndRoleV220Table')->delete();

        // 删除所有插件权限
        $db->name('access')->where('path','like', '/plus%')->where('api_path', '')->delete();
        // 删除所有外卖权限
        $db->name('access')->where('path','like', '/takeout%')->where('api_path', '')->delete();

        // 店铺权限
        $shopAccessData = [
            // 商家后台
            ['uuid' => 1731155608, 'name' => '营销管理', 'path' => '/marketing', 'api_path' => '', 'parent_uuid' => 1705895562, 'sort' => 1, 'is_route' => 1, 'icon' => 'icon-caiwuguanli', 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1, 'redirect_name'=> '/marketing/coupon/index', 'create_time' => 1636183780, 'update_time' => time()],
            // 优惠券
            ['uuid' => 1731155609, 'name' => '优惠券', 'path' => '/marketing/coupon/index', 'api_path' => '', 'parent_uuid' => 1731155608, 'sort' => 1, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
            ['uuid' => 1731155610, 'name' => '优惠券管理', 'path' => '/marketing/coupon/manage/index', 'api_path' => '/marketing/coupon/list', 'parent_uuid' => 1731155609, 'sort' => 2, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
            ['uuid' => 1731155611, 'name' => '添加', 'path' => '/marketing/coupon/add', 'api_path' => '/marketing/coupon/add', 'parent_uuid' => 1731155610, 'sort' => 2, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
            ['uuid' => 1731155612, 'name' => '编辑', 'path' => '/marketing/coupon/edit', 'api_path' => '/marketing/coupon/edit', 'parent_uuid' => 1731155610, 'sort' => 2, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
            ['uuid' => 1731155613, 'name' => '删除', 'path' => '/marketing/coupon/delete', 'api_path' => '/marketing/coupon/delete', 'parent_uuid' => 1731155610, 'sort' => 3, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
            ['uuid' => 1731155614, 'name' => '状态', 'path' => '/marketing/coupon/status', 'api_path' => '/marketing/coupon/status', 'parent_uuid' => 1731155610, 'sort' => 4, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
            ['uuid' => 1731155620, 'name' => '优惠券记录', 'path' => '/marketing/coupon/record/index', 'api_path' => '/marketing/coupon/record', 'parent_uuid' => 1731155609, 'sort' => 2, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
            // 营销活动
            ['uuid' => 1731155710, 'name' => '营销活动', 'path' => '/marketing/activity/index', 'api_path' => '', 'parent_uuid' => 1731155608, 'sort' => 3, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
            ['uuid' => 1731155711, 'name' => '活动管理', 'path' => '/marketing/activity/manage/index', 'api_path' => '/marketing/activity/list', 'parent_uuid' => 1731155710, 'sort' => 3, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
            ['uuid' => 1731155712, 'name' => '添加', 'path' => '/marketing/activity/add', 'api_path' => '/marketing/activity/add', 'parent_uuid' => 1731155711, 'sort' => 1, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
            ['uuid' => 1731155713, 'name' => '编辑', 'path' => '/marketing/activity/edit', 'api_path' => '/marketing/activity/edit', 'parent_uuid' => 1731155711, 'sort' => 2, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
            ['uuid' => 1731155714, 'name' => '失效', 'path' => '/marketing/activity/disable', 'api_path' => '/marketing/activity/disable', 'parent_uuid' => 1731155711, 'sort' => 2, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
            ['uuid' => 1731155720, 'name' => '发放记录', 'path' => '/marketing/activity/record', 'api_path' => '/marketing/activity/record', 'parent_uuid' => 1731155711, 'sort' => 2, 'is_route' => 1,  'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 1,  'create_time' => time(), 'update_time' => time()],
        ];
        $this->updateOrInsertData('access', 'uuid', $shopAccessData);
        // 店长角色
        $manager = $db->name('role')->where('name', '=', 'Store Manager')->find();
        if ($manager && isset($manager['uuid'])) {
            $managerRoleData = [
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155608', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155609', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155610', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155611', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155612', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155613', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155614', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155620', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155710', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155711', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155712', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155713', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155714', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1731155720', 'create_time' => time()],
            ];
            $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], $managerRoleData);
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
                // $query->update($item);
            } else {
                $db->name($tableName)->insert($item);
            }
        }
    }
}
