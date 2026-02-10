<?php

use think\migration\Migrator;

/**
 * 创建数据库连接池统计记录表
 *
 * 用于记录数据库连接池的健康状态数据，支持诊断 bad connection 问题
 */
class CreateDbPoolStats extends Migrator
{
    // 迁移目标：仅应用到 SaaS 主库
    const TARGET = 'main';

    /**
     * Change Method.
     *
     * 创建数据库连接池统计记录表
     */
    public function change()
    {
        // 检查表是否已存在
        if ($this->hasTable('db_pool_stats')) {
            return;
        }

        $table = $this->table('db_pool_stats', [
            'id' => false,
            'primary_key' => ['id'],
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_unicode_ci',
            'comment' => '数据库连接池统计记录表',
        ]);

        $table->addColumn('id', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'identity' => true,
                'comment' => '自增ID',
            ])
            ->addColumn('uuid', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'default' => 0,
                'comment' => '记录UUID',
            ])
            ->addColumn('instance_id', 'string', [
                'limit' => 128,
                'default' => '',
                'comment' => '服务实例标识',
            ])
            ->addColumn('db_name', 'string', [
                'limit' => 100,
                'default' => '',
                'comment' => '数据库名称',
            ])
            ->addColumn('max_open_conns', 'integer', [
                'limit' => 11,
                'default' => 0,
                'comment' => '最大打开连接数',
            ])
            ->addColumn('open_conns', 'integer', [
                'limit' => 11,
                'default' => 0,
                'comment' => '当前打开连接数',
            ])
            ->addColumn('in_use', 'integer', [
                'limit' => 11,
                'default' => 0,
                'comment' => '正在使用的连接数',
            ])
            ->addColumn('idle', 'integer', [
                'limit' => 11,
                'default' => 0,
                'comment' => '空闲连接数',
            ])
            ->addColumn('wait_count', 'biginteger', [
                'limit' => 20,
                'default' => 0,
                'comment' => '等待连接的累计次数',
            ])
            ->addColumn('wait_duration_ms', 'biginteger', [
                'limit' => 20,
                'default' => 0,
                'comment' => '等待连接的累计时间(毫秒)',
            ])
            ->addColumn('max_idle_closed', 'biginteger', [
                'limit' => 20,
                'default' => 0,
                'comment' => '因超过MaxIdleConns被关闭的连接数',
            ])
            ->addColumn('max_idle_time_closed', 'biginteger', [
                'limit' => 20,
                'default' => 0,
                'comment' => '因ConnMaxIdleTime被关闭的连接数',
            ])
            ->addColumn('max_lifetime_closed', 'biginteger', [
                'limit' => 20,
                'default' => 0,
                'comment' => '因ConnMaxLifetime被关闭的连接数',
            ])
            ->addColumn('sample_time', 'biginteger', [
                'limit' => 20,
                'default' => 0,
                'comment' => '采样时间戳(毫秒)',
            ])
            ->addColumn('create_time', 'integer', [
                'limit' => 10,
                'signed' => false,
                'default' => 0,
                'comment' => '创建时间(时间戳)',
            ])
            ->addColumn('update_time', 'integer', [
                'limit' => 10,
                'signed' => false,
                'default' => 0,
                'comment' => '更新时间(时间戳)',
            ])
            ->addColumn('delete_time', 'integer', [
                'limit' => 10,
                'signed' => false,
                'default' => 0,
                'comment' => '删除时间(时间戳)',
            ])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
            ->addIndex(['instance_id'], ['name' => 'idx_instance_id'])
            ->addIndex(['db_name'], ['name' => 'idx_db_name'])
            ->addIndex(['sample_time'], ['name' => 'idx_sample_time'])
            ->addIndex(['create_time'], ['name' => 'idx_create_time'])
            ->create();
    }
}
