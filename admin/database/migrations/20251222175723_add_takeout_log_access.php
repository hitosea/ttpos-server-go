<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddTakeoutLogAccess extends Migrator
{
    // 迁移目标
    const TARGET = 'main';

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
        $currentTime = time();
        
        // 云平台权限 - 日志管理
        $adminAccessData = [
            // 日志管理（主权限，parent_id = 101）
            [
                'id' => 214,
                'name' => '日志管理',
                'path' => '/log/index',
                'api_path' => '/admin/takeout/logs',
                'parent_id' => 101,
                'sort' => 25,
                'icon' => '',
                'redirect_name' => '',
                'is_route' => 1,
                'is_menu' => 1,
                'is_show' => 1,
                'remark' => '外卖相关日志管理',
                'is_supplier' => 0,
                'create_time' => $currentTime,
                'update_time' => $currentTime,
                'delete_time' => 0,
            ],
        ];
        $this->updateOrInsertData('admin_access', 'id', $adminAccessData);
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

