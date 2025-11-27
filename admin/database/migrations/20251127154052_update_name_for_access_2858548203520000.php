<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class UpdateNameForAccess2858548203520000 extends Migrator
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
        // 更新 ttpos_access 表中 uuid=2858548203520000 的 name 字段为 "品采收货(内部)"
        $accessData = [
            [
                'uuid' => 2858548203520000,
                'name' => '品采收货(内部)',
                'update_time' => time()
            ]
        ];
        
        $this->updateOrInsertData('access', 'uuid', $accessData);
    }

    /**
     * 更新或插入数据
     * 
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

