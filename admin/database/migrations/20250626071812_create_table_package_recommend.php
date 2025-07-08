<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class CreateTablePackageRecommend extends Migrator
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
        // 商品推荐
        if (!$this->hasTable('product_package_recommend')) {
            $table = $this->table('product_package_recommend', ['comment' => '商品推荐']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
                ->addColumn('status', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '是否开启推荐 0否 1是'])
                ->addColumn('title', 'string', ['limit' => 30, 'default' => '', 'comment' => '推荐标题'])
                ->addColumn('product_packages', 'text', ['null' => true, 'comment' => '推荐商品，对象数组'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->create();
        }
        // 店铺权限
        $shopAccessData = [
            // 商家后台
            ['uuid' => 1750930155, 'name' => '商品推荐', 'path' => '/product/store/recommend/index', 'api_path' => '/product.store.product/recommend', 'parent_uuid' => 1625541086, 'sort' => 4, 'icon' => '', 'redirect_name' => '', 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'plus_category_uuid' => 0, 'remark' => '', 'is_supplier' => 1, 'create_time' => 1750930155, 'update_time' => 1750930155, 'delete_time' => 0,],
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
