<?php
/**
 * 创建 ttpos_takeout_import_log 表
 * 
 * 任务: story-shop-takeout-import-progress Phase 1.3
 * 需求: Requirement 3.1 (导入历史日志记录)
 */

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTtposTakeoutImportLogTable extends Migrator
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
        if ($this->hasTable('takeout')) {
            $table = $this->table('takeout');
            
            // 检查 import_status 字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('import_status')) {
                $table->addColumn('import_status', 'integer', ['limit' => 3, 'default' => 0, 'null' => false, 'comment' => '导入状态(0-未导入 1-导入中 2-导入成功 3-导入失败)', 'after' => 'menu']);
            }
            
            // 检查索引是否不存在，如果不存在则添加
            if (!$table->hasIndexByName('idx_import_status')) {
                $table->addIndex('import_status', ['name' => 'idx_import_status']);
            }
            
            $table->update();
        }

        // 检查表是否不存在，如果不存在则创建
        if (!$this->hasTable('takeout_import_log')) {
            $table = $this->table('takeout_import_log', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '外卖导入日志表',
                'id' => false,
                'primary_key' => ['id']
            ]);

            $table
                // 基础字段
                ->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识'])
                
                // 导入信息字段
                ->addColumn('platform', 'string', ['limit' => 50, 'default' => '', 'null' => false, 'comment' => '外卖平台(grab/lineman等)'])
                ->addColumn('import_type', 'integer', ['limit' => 3, 'default' => 0, 'null' => false, 'comment' => '导入类型(1-TTPOS推送到平台 2-平台推送到TTPOS)'])
                ->addColumn('import_direction', 'string', ['limit' => 200, 'default' => '', 'null' => false, 'comment' => '导入方向描述'])
                
                // 状态和进度字段
                ->addColumn('status', 'integer', ['limit' => 3, 'default' => 0, 'null' => false, 'comment' => '导入状态(0-进行中 1-成功 2-失败)'])
                ->addColumn('progress', 'integer', ['limit' => 10, 'default' => 0, 'null' => false, 'comment' => '进度百分比(0-100)'])
                
                // 统计信息字段
                ->addColumn('success_count', 'integer', ['limit' => 10, 'default' => 0, 'null' => false, 'comment' => '成功数量'])
                ->addColumn('failure_count', 'integer', ['limit' => 10, 'default' => 0, 'null' => false, 'comment' => '失败数量'])
                ->addColumn('total_count', 'integer', ['limit' => 10, 'default' => 0, 'null' => false, 'comment' => '总数量'])
                ->addColumn('error_message', 'text', ['null' => true, 'comment' => '错误信息'])
                
                // 时间字段
                ->addColumn('start_time', 'integer', ['limit' => 10, 'default' => 0, 'null' => false, 'comment' => '开始时间'])
                ->addColumn('end_time', 'integer', ['limit' => 10, 'default' => 0, 'null' => false, 'comment' => '结束时间'])
                ->addColumn('duration', 'integer', ['limit' => 10, 'default' => 0, 'null' => false, 'comment' => '耗时(秒)'])
                ->addColumn('create_time', 'integer', ['limit' => 10, 'default' => 0, 'null' => false, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['limit' => 10, 'default' => 0, 'null' => false, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['limit' => 10, 'default' => 0, 'null' => false, 'comment' => '删除时间'])
                
                // 索引
                ->addIndex('uuid', ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex('platform', ['name' => 'idx_platform'])
                ->addIndex('import_type', ['name' => 'idx_import_type'])
                ->addIndex('status', ['name' => 'idx_status'])
                ->addIndex('create_time', ['name' => 'idx_create_time'])
                ->addIndex('delete_time', ['name' => 'idx_delete_time'])
                
                ->create();
        }
    }
}

