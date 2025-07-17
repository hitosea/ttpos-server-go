<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddCashierPermissionDeliveryOrderToAccess extends Migrator
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
        $shopAccessData = [
            // 收银机
            ['uuid' => 1752716650, 'name' => '外送', 'path' => 'cashier_accept_delivery', 'api_path' => '/cashier/member_order/list', 'parent_uuid' => 1704880670, 'sort' => 20, 'icon' => '', 'redirect_name' => '', 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'plus_category_uuid' => 0, 'remark' => '', 'is_supplier' => 0, 'create_time' => 1752716650, 'update_time' => time(), 'delete_time' => 0,],
        ];
        $this->updateOrInsertData('access', 'uuid', $shopAccessData);
        // 店长角色
        $cashier = $db->name('role')->where('name', '=', 'Cashier')->find();
        if ($cashier && isset($cashier['uuid'])) {
            $cashierRoleData = [
                ['uuid' => createUuid(), 'role_uuid' => $cashier['uuid'], 'access_uuid' => '1752716650', 'create_time' => time()],
            ];
            $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], $cashierRoleData);
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
