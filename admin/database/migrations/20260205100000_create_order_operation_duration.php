<?php

use think\migration\Migrator;

/**
 * 创建订单操作耗时记录表
 *
 * 用于记录订单操作的耗时数据，支持性能分析和分布式追踪
 */
class CreateOrderOperationDuration extends Migrator
{
    // 迁移目标：仅应用到 SaaS 主库
    const TARGET = 'main';

    /**
     * Change Method.
     *
     * 创建订单操作耗时记录表
     */
    public function change()
    {
        // 检查表是否已存在
        if ($this->hasTable('order_operation_duration')) {
            return;
        }

        $table = $this->table('order_operation_duration', [
            'id' => false,
            'primary_key' => ['id'],
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_unicode_ci',
            'comment' => '订单操作耗时记录表',
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
            ->addColumn('company_uuid', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'default' => 0,
                'comment' => '商户UUID',
            ])
            ->addColumn('sale_bill_uuid', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'default' => 0,
                'comment' => '账单UUID',
            ])
            ->addColumn('sale_order_uuid', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'default' => 0,
                'comment' => '订单UUID',
            ])
            ->addColumn('action', 'string', [
                'limit' => 100,
                'default' => '',
                'comment' => '操作类型(cancel_order/checkout/discount等)',
            ])
            ->addColumn('source', 'string', [
                'limit' => 50,
                'default' => '',
                'comment' => '来源终端(cashier/assistant/shop/h5等)',
            ])
            ->addColumn('staff_uuid', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'default' => 0,
                'comment' => '操作员工UUID',
            ])
            ->addColumn('device_sn', 'string', [
                'limit' => 255,
                'default' => '',
                'comment' => '设备序列号',
            ])
            ->addColumn('instance_id', 'string', [
                'limit' => 128,
                'default' => '',
                'comment' => '服务实例标识',
            ])
            ->addColumn('start_time', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'default' => 0,
                'comment' => '开始时间戳(毫秒)',
            ])
            ->addColumn('end_time', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'default' => 0,
                'comment' => '结束时间戳(毫秒)',
            ])
            ->addColumn('duration_ms', 'integer', [
                'limit' => 11,
                'signed' => false,
                'default' => 0,
                'comment' => '耗时(毫秒)',
            ])
            ->addColumn('request_path', 'string', [
                'limit' => 255,
                'default' => '',
                'comment' => '请求路径',
            ])
            ->addColumn('status', 'integer', [
                'limit' => \Phinx\Db\Adapter\MysqlAdapter::INT_TINY,
                'signed' => false,
                'default' => 1,
                'comment' => '操作结果(1成功 0失败)',
            ])
            ->addColumn('error_msg', 'text', [
                'null' => true,
                'comment' => '错误信息',
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
            ->addIndex(['company_uuid'], ['name' => 'idx_company_uuid'])
            ->addIndex(['sale_bill_uuid'], ['name' => 'idx_sale_bill_uuid'])
            ->addIndex(['action'], ['name' => 'idx_action'])
            ->addIndex(['instance_id'], ['name' => 'idx_instance_id'])
            ->addIndex(['create_time'], ['name' => 'idx_create_time'])
            ->create();
    }
}
