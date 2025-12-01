<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddIndexToExportRecord extends Migrator
{
    /**
     * Up Method.
     * 
     * 为 ttpos_export_record 表添加索引，优化查询同一天导出记录的性能
     */
    public function up()
    {
        // 为导出记录表添加复合索引，优化按导出类型、状态和日期范围查询的性能
        // 索引字段顺序：export_type, status, create_time
        // 这样可以优化 GetByDateAndType 方法的查询性能（查询条件包含 export_type, status=1, create_time 范围）
        $this->checkAndAddIndex('export_record', 'idx_export_type_status_date', ['export_type', 'status', 'create_time']);
    }

    /**
     * Down Method.
     * 
     * 删除索引（回退操作）
     */
    public function down()
    {
        // 检查表是否存在
        if (!$this->hasTable('export_record')) {
            return;
        }

        $table = $this->table('export_record');
        
        // 检查索引是否存在，如果存在则删除
        if ($table->hasIndex('idx_export_type_status_date')) {
            $table->removeIndex(['export_type', 'status', 'create_time'], [
                'name' => 'idx_export_type_status_date'
            ])->update();
        }
    }

    /**
     * 检查并添加索引
     * @param string $tableName 表名
     * @param string $indexName 索引名
     * @param array $columns 索引字段
     */
    protected function checkAndAddIndex($tableName, $indexName, $columns)
    {
        try {
            // 检查表是否存在
            if (!$this->hasTable($tableName)) {
                return;
            }

            $table = $this->table($tableName);
            
            // 检查索引是否已存在
            if ($table->hasIndex($indexName)) {
                return;
            }

            // 添加索引
            $table->addIndex($columns, [
                'name' => $indexName,
                'unique' => false
            ])->update();
        } catch (\Exception $e) {
            // 索引已存在或其他错误，忽略
        }
    }
}

