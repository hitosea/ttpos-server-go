<?php

use think\migration\Migrator;
use think\migration\worker\Incr;
use Phinx\Db\Adapter\MysqlAdapter;

class CreateTTPOSFullReductionActivityRuleTable extends Migrator
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
        if ($this->hasTable('ttpos_full_reduction_activity_rule')) {
            return;
        }

        $table = $this->table('ttpos_full_reduction_activity_rule', [
            'comment' => '满减活动规则表',
            'engine' => 'InnoDB',
            'encoding' => 'utf8mb4',
            'collation' => 'utf8mb4_unicode_ci',
            'id' => false,
            'primary_key' => ['id']
        ]);

        // 定义表字段
        $table->addColumn('id', 'integer', ['identity' => true, 'comment' => '主键ID', 'signed' => false]);
        $table->addColumn('uuid', 'biginteger', ['comment' => '唯一标识', 'default' => 0, 'signed' => false]);
        $table->addColumn('full_reduction_activity_uuid', 'biginteger', ['comment' => '活动UUID', 'default' => 0, 'signed' => false]);
        $table->addColumn('threshold', 'decimal', ['comment' => '阈值（满减条件，如满200减20中的200）', 'default' => 0, 'signed' => false, 'precision' => 22, 'scale' => 4]);
        $table->addColumn('reduction_amount', 'decimal', ['comment' => '扣减值（满减金额，如满200减20中的20）', 'default' => 0, 'signed' => false, 'precision' => 22, 'scale' => 4]);
        $table->addColumn('create_time', 'integer', ['comment' => '创建时间', 'default' => 0, 'signed' => false]);
        $table->addColumn('update_time', 'integer', ['comment' => '更新时间', 'default' => 0, 'signed' => false]);
        $table->addColumn('delete_time', 'integer', ['comment' => '删除时间', 'default' => 0, 'signed' => false]);

        // 创建索引
        $table->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid']);
        $table->addIndex(['full_reduction_activity_uuid'], ['name' => 'idx_full_reduction_activity_uuid']);
        $table->addIndex(['threshold'], ['name' => 'idx_threshold']);

        $table->create();
    }
    
}

