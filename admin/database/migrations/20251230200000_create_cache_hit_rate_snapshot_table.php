<?php
use think\migration\Migrator;
use think\migration\db\Column;

class CreateCacheHitRateSnapshotTable extends Migrator
{
    const TARGET = 'main';

    /**
     * 创建缓存命中率快照表（saas库）
     * 用于定时保存缓存命中率统计快照，避免程序重启导致统计丢失
     */
    public function change()
    {
        if (!$this->hasTable('cache_hit_rate_snapshot')) {
            $table = $this->table('cache_hit_rate_snapshot', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '缓存命中率快照表'
            ]);

            $table->addColumn('id', 'biginteger', ['limit' => 20, 'signed' => false, 'identity' => true, 'comment' => '主键ID'])
                ->addColumn('instance_id', 'string', ['limit' => 128, 'default' => '', 'comment' => '实例标识（hostname或配置的实例ID）'])
                ->addColumn('snapshot_time', 'datetime', ['comment' => '快照时间'])
                ->addColumn('hits', 'biginteger', ['limit' => 20, 'signed' => false, 'default' => 0, 'comment' => '命中次数'])
                ->addColumn('misses', 'biginteger', ['limit' => 20, 'signed' => false, 'default' => 0, 'comment' => '未命中次数'])
                ->addColumn('total', 'biginteger', ['limit' => 20, 'signed' => false, 'default' => 0, 'comment' => '总请求数'])
                ->addColumn('hit_rate', 'decimal', ['precision' => 5, 'scale' => 2, 'default' => 0.00, 'comment' => '命中率（百分比）'])
                ->addColumn('key_count', 'integer', ['limit' => 11, 'signed' => false, 'default' => 0, 'comment' => 'Key数量'])
                ->addColumn('key_stats', 'json', ['null' => true, 'comment' => 'Key级别统计（JSON格式）'])
                ->addColumn('is_restart', 'integer', ['limit' => 1, 'signed' => false, 'default' => 0, 'comment' => '是否是重启后的第一条快照（1:是 0:否）'])
                ->addColumn('create_time', 'biginteger', ['limit' => 20, 'signed' => false, 'default' => 0, 'comment' => '创建时间（时间戳）'])
                ->addIndex(['instance_id', 'snapshot_time'], ['name' => 'idx_instance_snapshot'])
                ->addIndex(['snapshot_time'], ['name' => 'idx_snapshot_time'])
                ->addIndex(['instance_id', 'is_restart'], ['name' => 'idx_instance_restart'])
                ->create();
        }
    }
}

