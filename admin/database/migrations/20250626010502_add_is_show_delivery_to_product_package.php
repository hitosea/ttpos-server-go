<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddIsShowDeliveryToProductPackage extends Migrator
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
        $table = $this->table('product_package');
        if (!$table->hasColumn('is_show_delivery')) {
            $table->addColumn('is_show_delivery', 'integer', ['default' => 0, 'comment' => '是否在外送显示, 0-否 1-是', 'after' => 'is_show_h5']);
            $table->update();
        }
        // 店铺权限
        $shopAccessData = [
            // 商家后台
            ['uuid' => 1741590679, 'name' => '外送订单', 'path' => '/store/takeout/index', 'api_path' => '/store/DeliveryOrder/index', 'parent_uuid' => 1626506017, 'sort' => 2, 'icon' => '', 'redirect_name' => '', 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'plus_category_uuid' => 0, 'remark' => '', 'is_supplier' => 1, 'create_time' => 1741590679, 'update_time' => 1741590679, 'delete_time' => 0,],
            ['uuid' => 1741590680, 'name' => '详情', 'path' => '/store/takeout/detail', 'api_path' => '/store/DeliveryOrder/detail', 'parent_uuid' => 1741590679, 'sort' => 0, 'icon' => '', 'redirect_name' => '', 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'plus_category_uuid' => 0, 'remark' => '', 'is_supplier' => 1, 'create_time' => 1741590679, 'update_time' => 1741590679, 'delete_time' => 0,]
        ];
        $this->updateOrInsertData('access', 'uuid', $shopAccessData);
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
