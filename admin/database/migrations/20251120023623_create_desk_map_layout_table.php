<?php
/**
 * 创建桌台地图布局表
 * 
 * 任务: story-admin-desktop-table-map Phase 2.1
 * 需求: R1.1-R1.6
 * 
 * 用途: 存储各区域的桌台地图布局配置（坐标、尺寸、样式等）
 * 
 * @version v2.10.0
 */

use think\migration\Migrator;
use think\migration\db\Column;

class CreateDeskMapLayoutTable extends Migrator
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
        // 检查表是否已存在
        if (!$this->hasTable('desk_map_layout')) {
            $table = $this->table('desk_map_layout', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_general_ci',
                'comment' => '桌台地图布局表'
            ]);

            $table
                // 主键ID
                ->addColumn('id', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'identity' => true,
                    'comment' => '主键ID'
                ])
                // UUID - 唯一标识
                ->addColumn('uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'null' => false,
                    'comment' => '布局UUID'
                ])
                // 区域UUID
                ->addColumn('region_uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'null' => false,
                    'comment' => '区域UUID'
                ])
                // 布局JSON数据
                ->addColumn('layout_json', 'text', [
                    'null' => false,
                    'comment' => '画布布局JSON（含桌台坐标、尺寸、样式等）'
                ])
                // 创建时间
                ->addColumn('create_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'null' => false,
                    'comment' => '创建时间'
                ])
                // 更新时间
                ->addColumn('update_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'null' => false,
                    'comment' => '更新时间'
                ])
                // 删除时间（软删除）
                ->addColumn('delete_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'null' => false,
                    'comment' => '删除时间（软删除）'
                ])
                // 添加唯一索引：UUID
                ->addIndex(['uuid'], [
                    'unique' => true,
                    'name' => 'uk_uuid'
                ])
                // 添加唯一索引：一个区域只能有一个布局配置
                ->addIndex(['region_uuid'], [
                    'unique' => true,
                    'name' => 'uk_region_uuid'
                ])
                // 添加普通索引：软删除查询
                ->addIndex(['delete_time'], [
                    'name' => 'idx_delete_time'
                ])
                ->create();
        }
    }
}

