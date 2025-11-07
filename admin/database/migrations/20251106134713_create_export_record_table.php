<?php

use think\migration\Migrator;

class CreateExportRecordTable extends Migrator
{
    /**
     * 创建导出记录表
     */
    public function change()
    {
        // 检查表是否已经存在
        if ($this->hasTable('export_record')) {
            return;
        }

        $table = $this->table('export_record', [
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_general_ci',
            'comment' => '导出记录表',
        ]);

        $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '导出记录UUID'])
              ->addColumn('export_type', 'integer', ['limit' => 3, 'default' => 0, 'comment' => '导出类型: 1-时段营业统计, 2-综合运营统计, 3-营业应收统计, 4-菜品出品明细, 5-菜品出品详情'])
              ->addColumn('export_name', 'string', ['limit' => 200, 'default' => '', 'comment' => '导出文件名称'])
              ->addColumn('file_uuid', 'biginteger', ['default' => 0, 'comment' => '文件UUID，关联ttpos_file表'])
              ->addColumn('status', 'integer', ['limit' => 3, 'default' => 0, 'comment' => '状态: 0-导出中, 1-导出成功, 2-导出失败'])
              ->addColumn('error_msg', 'text', ['null' => true, 'comment' => '错误信息'])
              ->addColumn('export_params', 'text', ['null' => true, 'comment' => '导出参数JSON'])
              ->addColumn('staff_uuid', 'biginteger', ['default' => 0, 'comment' => '操作员工UUID'])
              ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
              ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
              ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
              ->addIndex(['export_type'], ['name' => 'idx_export_type'])
              ->addIndex(['status'], ['name' => 'idx_status'])
              ->addIndex(['create_time'], ['name' => 'idx_create_time'])
              ->create();
    }
}




