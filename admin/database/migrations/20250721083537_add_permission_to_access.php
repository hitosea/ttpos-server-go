<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddPermissionToAccess extends Migrator
{
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

        $db = Db::connect(Db::getConfig('default'), true);

        $db->name('access')->where('api_path', '/store/DeliveryOrder/index')->update(['api_path' => '/store/MemberOrder/index']);
        $db->name('access')->where('api_path', '/store/DeliveryOrder/detail')->update(['api_path' => '/store/MemberOrder/detail']);

        $shopAccessData = [
            ['uuid' => 1753087456, 'name' => '拒单', 'path' => '/store/takeout/reject', 'api_path' => '/store/MemberOrder/reject', 'parent_uuid' => 1741590679, 'sort' => 0, 'icon' => '', 'redirect_name' => '', 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'plus_category_uuid' => 0, 'remark' => '', 'is_supplier' => 1, 'create_time' => time(), 'update_time' => time(), 'delete_time' => 0,],
            ['uuid' => 1753087457, 'name' => '取消订单', 'path' => '/store/takeout/cancel', 'api_path' => '/store/MemberOrder/cancel', 'parent_uuid' => 1741590679, 'sort' => 0, 'icon' => '', 'redirect_name' => '', 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'plus_category_uuid' => 0, 'remark' => '', 'is_supplier' => 1, 'create_time' => time(), 'update_time' => time(), 'delete_time' => 0,],
            ['uuid' => 1753087458, 'name' => '退款', 'path' => '/store/takeout/refund', 'api_path' => '/store/MemberOrder/refund', 'parent_uuid' => 1741590679, 'sort' => 0, 'icon' => '', 'redirect_name' => '', 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'plus_category_uuid' => 0, 'remark' => '', 'is_supplier' => 1, 'create_time' => time(), 'update_time' => time(), 'delete_time' => 0,],
            ['uuid' => 1753087459, 'name' => '导出', 'path' => '/store/takeout/export', 'api_path' => '/store/MemberOrder/export', 'parent_uuid' => 1741590679, 'sort' => 0, 'icon' => '', 'redirect_name' => '', 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'plus_category_uuid' => 0, 'remark' => '', 'is_supplier' => 1, 'create_time' => time(), 'update_time' => time(), 'delete_time' => 0,],
        ];
        $this->updateOrInsertData('access', 'uuid', $shopAccessData);

        // 店长角色
        $manager = $db->name('role')->where('name', '=', 'Store Manager')->find();
        if ($manager && isset($manager['uuid'])) {
            $managerRoleData = [
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1753087456', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1753087457', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1753087458', 'create_time' => time()],
                ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '1753087459', 'create_time' => time()],
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
                $query->update($item);
            } else {
                $db->name($tableName)->insert($item);
            }
        }
    }
}
