<?php
/**
 * 将 desk_map_layout 表的 area_uuid 字段重命名为 region_uuid
 * 
 * 任务: 字段重命名
 * 需求: 统一使用 region_uuid 命名
 * 
 * @version v2.10.0
 */

use think\migration\Migrator;
use think\migration\db\Column;

class RenameAreaUuidToRegionUuidInDeskMapLayout extends Migrator
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
        // 检查表是否存在
        if (!$this->hasTable('desk_map_layout')) {
            return;
        }

        $table = $this->table('desk_map_layout');

        // 检查字段是否存在
        if ($table->hasColumn('area_uuid')) {
            // 删除旧的唯一索引（如果存在）
            try {
                if ($table->hasIndex('uk_area_uuid')) {
                    $table->removeIndex(['area_uuid'], ['name' => 'uk_area_uuid']);
                }
            } catch (\Exception $e) {
                // 索引不存在，忽略错误
            }

            // 重命名字段
            $table->renameColumn('area_uuid', 'region_uuid');

            // 创建新的唯一索引
            $table->addIndex(['region_uuid'], [
                'unique' => true,
                'name' => 'uk_region_uuid'
            ]);

            $table->update();
        }
    }
}

