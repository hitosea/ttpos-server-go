<?php

use think\migration\Migrator;
use think\migration\worker\Incr;
use Phinx\Db\Adapter\MysqlAdapter;

class CreateTTPOSKitchenEfficiencyAnalysisTable extends Migrator
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
        if ($this->hasTable('kitchen_efficiency_analysis')) {
            return;
        }

        $table = $this->table('kitchen_efficiency_analysis', [
            'comment' => '后厨效率分析表',
            'engine' => 'InnoDB',
            'encoding' => 'utf8mb4',
            'collation' => 'utf8mb4_unicode_ci',
            'id' => false,
            'primary_key' => ['id']
        ]);

        // 定义表字段
        $table->addColumn('id', 'integer', ['identity' => true, 'comment' => '主键ID', 'signed' => false]);
        $table->addColumn('uuid', 'biginteger', ['comment' => 'UUID', 'default' => 0, 'signed' => false]);
        $table->addColumn('product_package_uuid', 'biginteger', ['comment' => '商品包UUID', 'default' => 0, 'signed' => false]);
        $table->addColumn('min', 'decimal', ['comment' => '最短出品时长', 'default' => 0, 'signed' => false, 'precision' => 22, 'scale' => 4]);
        $table->addColumn('max', 'decimal', ['comment' => '最长出品时长', 'default' => 0, 'signed' => false, 'precision' => 22, 'scale' => 4]);
        $table->addColumn('avg', 'decimal', ['comment' => '平均出品时长', 'default' => 0, 'signed' => false, 'precision' => 22, 'scale' => 4]);
        $table->addColumn('total', 'decimal', ['comment' => '总出品时长', 'default' => 0, 'signed' => false, 'precision' => 22, 'scale' => 4]);
        $table->addColumn('count', 'decimal', ['comment' => '出品次数', 'default' => 0, 'signed' => false, 'precision' => 22, 'scale' => 4]);
        $table->addColumn('date', 'integer', ['comment' => '统计日期,格式:yyyyMMdd,如20251103.一个商品一天只有唯一的一条记录', 'default' => 0, 'signed' => false]);
        $table->addColumn('create_time', 'integer', ['comment' => '创建时间', 'default' => 0, 'signed' => false]);
        $table->addColumn('update_time', 'integer', ['comment' => '更新时间', 'default' => 0, 'signed' => false]);
        $table->addColumn('delete_time', 'integer', ['comment' => '删除时间', 'default' => 0, 'signed' => false]);

        // 创建索引
        $table->addIndex(['uuid'], ['unique' => true, 'name' => 'idx_uuid']);
        $table->addIndex(['product_package_uuid', 'date'], ['unique' => true, 'name' => 'unique_product_package_date']);

        $table->create();
    }
    
}
