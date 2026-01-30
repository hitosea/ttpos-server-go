<?php

use think\migration\Migrator;

/**
 * 创建外卖订单更新日志表
 * 
 * 用于记录外卖订单的更新历史，包含更新前后的完整订单数据
 */
class AddTakeoutOrderUpdateLogTable extends Migrator
{
    /**
     * Change Method.
     *
     * 创建外卖订单更新日志表
     */
    public function change()
    {
        // 检查表是否已存在
        if ($this->hasTable('takeout_order_update_log')) {
            return;
        }

        $table = $this->table('takeout_order_update_log', [
            'id' => false,
            'primary_key' => ['id'],
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_unicode_ci',
            'comment' => '外卖订单更新日志表',
        ]);

        $table->addColumn('id', 'integer', [
                'limit' => 11,
                'signed' => false,
                'identity' => true,
                'comment' => '自增ID',
            ])
            ->addColumn('uuid', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'default' => 0,
                'comment' => '日志UUID',
            ])
            ->addColumn('takeout_order_uuid', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'default' => 0,
                'comment' => '外卖订单UUID',
            ])
            ->addColumn('old_data', 'text', [
                'limit' => \Phinx\Db\Adapter\MysqlAdapter::TEXT_LONG,
                'null' => true,
                'comment' => '更新前订单数据(JSON格式,包含订单主表、商品、修饰符等完整数据)',
            ])
            ->addColumn('new_data', 'text', [
                'limit' => \Phinx\Db\Adapter\MysqlAdapter::TEXT_LONG,
                'null' => true,
                'comment' => '更新后订单数据(JSON格式,包含订单主表、商品、修饰符等完整数据)',
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
            ->addIndex(['takeout_order_uuid'], ['name' => 'idx_takeout_order_uuid'])
            ->addIndex(['create_time'], ['name' => 'idx_create_time'])
            ->create();
    }
}
