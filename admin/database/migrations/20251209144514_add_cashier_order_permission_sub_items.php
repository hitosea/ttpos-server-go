<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

/**
 * 收银机接单权限子项管理迁移脚本
 * 
 * 任务: story-shop-cashier-order-permission-sub-items Phase 1
 * 需求: R1.1-R1.7
 * 
 * 功能：
 * 1. 在收银机权限下新增"接单"权限（与沽清权限平级）
 * 2. 将原有接单权限调整为扫码接单子项
 * 3. 将外送权限调整为接单权限的子项
 * 4. 新增外卖权限作为接单权限的子项
 * 5. 为默认角色分配新接单权限和外卖权限
 */
class AddCashierOrderPermissionSubItems extends Migrator
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
        
        // 新接单权限UUID（父级）
        $newAcceptOrderUuid = 1734000000;
        // 外卖权限UUID（子级）
        $deliveryPlatformUuid = 1734000001;
        
        // 1. 新增接单权限（父级，与沽清权限平级）
        $newAcceptOrderData = [
            [
                'uuid' => $newAcceptOrderUuid,
                'name' => '接单',
                'path' => 'cashier_accept_order_parent', // 新接单权限的 path，避免与原有接单权限（cashier_accept_order）冲突
                'api_path' => '',
                'parent_uuid' => 1704880670, // 收银机权限UUID
                'sort' => 11, // 沽清权限sort为10，接单权限sort为11，排序在沽清后
                'icon' => '',
                'redirect_name' => '',
                'is_route' => 0,
                'is_menu' => 0,
                'is_show' => 1,
                'plus_category_uuid' => 0,
                'remark' => '',
                'is_supplier' => 0,
                'create_time' => time(),
                'update_time' => time(),
                'delete_time' => 0,
            ],
        ];
        $this->updateOrInsertData('access', 'uuid', $newAcceptOrderData);
        
        // 2. 更新原有接单权限（1724320522）为扫码接单子项
        $scanOrderData = [
            [
                'uuid' => 1724320522,
                'name' => '扫码接单', // 名称修改
                'path' => 'cashier_accept_order', // 路径保持原样，不修改
                'api_path' => '/store/TakeOrder/list', // 保持不变
                'parent_uuid' => $newAcceptOrderUuid, // 从1704880670改为新接单权限UUID
                'sort' => 10, // 作为第一项子级
                'icon' => '',
                'redirect_name' => '',
                'is_route' => 0,
                'is_menu' => 0,
                'is_show' => 1,
                'plus_category_uuid' => 0,
                'remark' => '',
                'is_supplier' => 0,
                'create_time' => 1733367005, // 保持原有创建时间
                'update_time' => time(),
                'delete_time' => 0,
            ],
        ];
        $this->updateOrInsertData('access', 'uuid', $scanOrderData);
        
        // 3. 更新外送权限（1752716650）的 parent_uuid
        $deliveryData = [
            [
                'uuid' => 1752716650,
                'name' => '外送', // 保持不变
                'path' => 'cashier_accept_delivery', // 保持不变
                'api_path' => '/cashier/member_order/list', // 保持不变
                'parent_uuid' => $newAcceptOrderUuid, // 从1704880670改为新接单权限UUID
                'sort' => 20, // 作为第二项子级
                'icon' => '',
                'redirect_name' => '',
                'is_route' => 0,
                'is_menu' => 0,
                'is_show' => 1,
                'plus_category_uuid' => 0,
                'remark' => '',
                'is_supplier' => 0,
                'create_time' => 1752716650, // 保持原有创建时间
                'update_time' => time(),
                'delete_time' => 0,
            ],
        ];
        $this->updateOrInsertData('access', 'uuid', $deliveryData);
        
        // 4. 新增外卖权限（子级）
        $deliveryPlatformData = [
            [
                'uuid' => $deliveryPlatformUuid,
                'name' => '外卖',
                'path' => 'cashier_accept_delivery_platform',
                'api_path' => '/cashier/grab_order/list', // 外卖订单列表，待确认
                'parent_uuid' => $newAcceptOrderUuid, // 新接单权限UUID
                'sort' => 30, // 作为第三项子级
                'icon' => '',
                'redirect_name' => '',
                'is_route' => 0,
                'is_menu' => 0,
                'is_show' => 1,
                'plus_category_uuid' => 0,
                'remark' => '',
                'is_supplier' => 0,
                'create_time' => time(),
                'update_time' => time(),
                'delete_time' => 0,
            ],
        ];
        $this->updateOrInsertData('access', 'uuid', $deliveryPlatformData);
        
        // 5. 为默认角色分配新接单权限
        $storeManager = $db->name('role')->where('name', '=', 'Store Manager')->find();
        if ($storeManager && isset($storeManager['uuid'])) {
            $storeManagerRoleData = [
                [
                    'uuid' => createUuid(),
                    'role_uuid' => $storeManager['uuid'],
                    'access_uuid' => $newAcceptOrderUuid,
                    'create_time' => time(),
                    'update_time' => time(),
                    'delete_time' => 0,
                ],
            ];
            $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], $storeManagerRoleData);
        }
        
        $cashier = $db->name('role')->where('name', '=', 'Cashier')->find();
        if ($cashier && isset($cashier['uuid'])) {
            $cashierRoleData = [
                [
                    'uuid' => createUuid(),
                    'role_uuid' => $cashier['uuid'],
                    'access_uuid' => $newAcceptOrderUuid,
                    'create_time' => time(),
                    'update_time' => time(),
                    'delete_time' => 0,
                ],
            ];
            $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], $cashierRoleData);
        }
        
        // 6. 为默认角色分配外卖权限
        if ($storeManager && isset($storeManager['uuid'])) {
            $storeManagerDeliveryRoleData = [
                [
                    'uuid' => createUuid(),
                    'role_uuid' => $storeManager['uuid'],
                    'access_uuid' => $deliveryPlatformUuid,
                    'create_time' => time(),
                    'update_time' => time(),
                    'delete_time' => 0,
                ],
            ];
            $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], $storeManagerDeliveryRoleData);
        }
        
        if ($cashier && isset($cashier['uuid'])) {
            $cashierDeliveryRoleData = [
                [
                    'uuid' => createUuid(),
                    'role_uuid' => $cashier['uuid'],
                    'access_uuid' => $deliveryPlatformUuid,
                    'create_time' => time(),
                    'update_time' => time(),
                    'delete_time' => 0,
                ],
            ];
            $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], $cashierDeliveryRoleData);
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
