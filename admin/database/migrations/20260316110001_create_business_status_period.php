<?php
/**
 * 创建测试营业时段记录表
 *
 * 任务: #40165 新管理端-增加"营业状态"字段，区分正式营业、测试营业的数据统计
 * 仅记录测试营业时段，end_time=0 表示测试进行中
 */

use think\migration\Migrator;

class CreateBusinessStatusPeriod extends Migrator
{
    public function change()
    {
        if (!$this->hasTable('business_status_period')) {
            $table = $this->table('business_status_period', [
                'id' => false,
                'primary_key' => 'id',
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '测试营业时段记录表',
            ]);

            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'limit' => 11])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '记录UUID'])
                ->addColumn('start_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '测试营业开始时间(时间戳)'])
                ->addColumn('end_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '测试营业结束时间(时间戳), 0=进行中'])
                ->addColumn('create_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['end_time'], ['name' => 'idx_end_time'])
                ->addIndex(['start_time', 'end_time'], ['name' => 'idx_time_range'])
                ->create();
        }
    }
}
