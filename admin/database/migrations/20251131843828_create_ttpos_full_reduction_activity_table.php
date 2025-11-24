<?php

use think\migration\Migrator;
use think\migration\worker\Incr;
use Phinx\Db\Adapter\MysqlAdapter;

class CreateTTPOSFullReductionActivityTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-change-method
     *
     * @return void
     */
    public function change()
    {
        // 检查表是否已经存在
        if ($this->hasTable('full_reduction_activity')) {
            return;
        }

        $table = $this->table('full_reduction_activity', [
            'comment' => '满减活动表',
            'engine' => 'InnoDB',
            'encoding' => 'utf8mb4',
            'collation' => 'utf8mb4_unicode_ci',
            'id' => false,
            'primary_key' => ['id']
        ]);

        // 定义表字段
        $table->addColumn('id', 'integer', ['identity' => true, 'comment' => '主键ID', 'signed' => false]);
        $table->addColumn('uuid', 'biginteger', ['comment' => '唯一标识', 'default' => 0, 'signed' => false]);
        $table->addColumn('name', 'string', ['limit' => 1000, 'comment' => '活动名称（JSON格式）', 'default' => '']);
        $table->addColumn('multi_language_name_uuid', 'biginteger', ['comment' => '多语言名称UUID', 'default' => 0, 'signed' => false]);
        $table->addColumn('start_date', 'integer', ['comment' => '活动开始日期（时间戳，当天00:00:00）', 'default' => 0, 'signed' => false]);
        $table->addColumn('end_date', 'integer', ['comment' => '活动结束日期（时间戳，当天23:59:59）', 'default' => 0, 'signed' => false]);
        $table->addColumn('start_time', 'string', ['limit' => 255, 'comment' => '适用时间开始（格式：HH:mm，如09:00）', 'default' => '']);
        $table->addColumn('end_time', 'string', ['limit' => 255, 'comment' => '适用时间结束（格式：HH:mm，如22:00）', 'default' => '']);
        $table->addColumn('is_all_day', 'integer', ['comment' => '是否全天（1=全天，0=特定时段）', 'default' => 0]);
        $table->addColumn('reduction_type', 'integer', ['comment' => '满减方式（0=阶梯满减，1=循环满减）', 'default' => 0]);
        $table->addColumn('is_disabled', 'integer', ['comment' => '是否失效（1=失效，0=未失效）', 'default' => 0]);
        $table->addColumn('create_time', 'integer', ['comment' => '创建时间', 'default' => 0, 'signed' => false]);
        $table->addColumn('update_time', 'integer', ['comment' => '更新时间', 'default' => 0, 'signed' => false]);
        $table->addColumn('delete_time', 'integer', ['comment' => '删除时间', 'default' => 0, 'signed' => false]);

        // 创建索引
        $table->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid']);
        $table->addIndex(['start_date'], ['name' => 'idx_start_date']);
        $table->addIndex(['end_date'], ['name' => 'idx_end_date']);
        $table->addIndex(['multi_language_name_uuid'], ['name' => 'idx_multi_language_name_uuid']);

        $table->create();
    }
    
}

