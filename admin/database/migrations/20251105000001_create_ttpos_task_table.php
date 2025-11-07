<?php

use think\migration\Migrator;
use think\migration\worker\Incr;
use Phinx\Db\Adapter\MysqlAdapter;

class CreateTTPOSTaskTable extends Migrator
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
        $table = $this->table('ttpos_task', [
            'comment' => '任务中心表',
            'engine' => 'InnoDB',
            'encoding' => 'utf8mb4',
            'collation' => 'utf8mb4_unicode_ci',
            'id' => false,
            'primary_key' => ['id']
        ]);

        // 定义表字段
        $table->addColumn('id', 'integer', ['identity' => true, 'comment' => '主键ID', 'signed' => false]);
        $table->addColumn('uuid', 'biginteger', ['comment' => 'UUID', 'default' => 0, 'signed' => false]);
        $table->addColumn('company_uuid', 'biginteger', ['comment' => '所属公司UUID', 'default' => 0, 'signed' => false]);
        $table->addColumn('type', 'string', ['comment' => '任务类型', 'limit' => 50, 'default' => '']);
        $table->addColumn('status', 'integer', ['comment' => '任务状态', 'default' => 0, 'signed' => false]);
        $table->addColumn('params', 'text', ['comment' => '任务参数']);
        $table->addColumn('result', 'text', ['comment' => '任务结果']);
        $table->addColumn('error', 'text', ['comment' => '任务错误']);
        $table->addColumn('log', 'text', ['comment' => '任务日志']);
        $table->addColumn('priority', 'integer', ['comment' => '优先级,数字越大优先级越高', 'default' => 0, 'signed' => false]);
        $table->addColumn('create_time', 'integer', ['comment' => '创建时间', 'default' => 0, 'signed' => false]);
        $table->addColumn('update_time', 'integer', ['comment' => '更新时间', 'default' => 0, 'signed' => false]);
        $table->addColumn('delete_time', 'integer', ['comment' => '删除时间', 'default' => 0, 'signed' => false]);

        // 创建索引
        $table->addIndex(['uuid'], ['unique' => true, 'name' => 'idx_uuid']);

        $table->create();
    }

    /**
     * Down Method.
     *
     * @return void
     */
    public function down()
    {
        // 检查表是否存在，如果存在则删除
        if ($this->hasTable('ttpos_task')) {
            $this->dropTable('ttpos_task');
        }
    }
}
