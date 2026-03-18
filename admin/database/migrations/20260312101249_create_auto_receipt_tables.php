<?php
/**
 * 创建品采自动收货相关表（规则主表、规则门店子表、收货记录表）
 * 所属库：saas 主库
 */

use think\migration\Migrator;

class CreateAutoReceiptTables extends Migrator
{
    // 迁移目标：仅应用到 SaaS 主库
    const TARGET = 'main';

    public function change()
    {
        // ==================== 表1: 自动收货规则主表 ====================
        if (!$this->hasTable('auto_receipt_rule')) {
            $table = $this->table('auto_receipt_rule', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '品采自动收货规则主表',
            ]);

            $table->addColumn('id', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'identity' => true,
                    'comment' => '主键ID',
                ])
                ->addColumn('uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '唯一标识',
                ])
                ->addColumn('headquarter_company_uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '总部company_uuid（租户隔离）',
                ])
                ->addColumn('name', 'string', [
                    'limit' => 1000,
                    'default' => '',
                    'comment' => '规则名称（多语言JSON）',
                ])
                ->addColumn('warehouse_erp_code', 'string', [
                    'limit' => 100,
                    'default' => '',
                    'comment' => '发货仓库ERP编码',
                ])
                ->addColumn('delay_days', 'integer', [
                    'limit' => 11,
                    'default' => 0,
                    'comment' => 'DN发送后N天自动收货（0=当天24:00）',
                ])
                ->addColumn('status', 'integer', [
                    'limit' => \Phinx\Db\Adapter\MysqlAdapter::INT_TINY,
                    'default' => 1,
                    'comment' => '状态：1=启用，0=禁用',
                ])
                ->addColumn('create_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '创建时间',
                ])
                ->addColumn('update_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '更新时间',
                ])
                ->addColumn('delete_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '删除时间',
                ])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['headquarter_company_uuid', 'status', 'delete_time'], ['name' => 'idx_hq_status'])
                ->create();
        }

        // ==================== 表2: 自动收货规则门店子表 ====================
        if (!$this->hasTable('auto_receipt_rule_shop')) {
            $table = $this->table('auto_receipt_rule_shop', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '品采自动收货规则门店子表',
            ]);

            $table->addColumn('id', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'identity' => true,
                    'comment' => '主键ID',
                ])
                ->addColumn('uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '唯一标识',
                ])
                ->addColumn('rule_uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '关联规则主表UUID',
                ])
                ->addColumn('shop_company_uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '门店company_uuid',
                ])
                ->addColumn('create_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '创建时间',
                ])
                ->addColumn('update_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '更新时间',
                ])
                ->addColumn('delete_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '删除时间',
                ])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['rule_uuid', 'delete_time'], ['name' => 'idx_rule'])
                ->addIndex(['shop_company_uuid', 'delete_time'], ['name' => 'idx_shop'])
                ->create();
        }

        // ==================== 表3: 自动收货记录表 ====================
        if (!$this->hasTable('auto_receipt_log')) {
            $table = $this->table('auto_receipt_log', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '品采自动收货记录表',
            ]);

            $table->addColumn('id', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'identity' => true,
                    'comment' => '主键ID',
                ])
                ->addColumn('uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '唯一标识',
                ])
                ->addColumn('headquarter_company_uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '总部company_uuid（租户隔离）',
                ])
                ->addColumn('rule_uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '触发规则UUID（关联auto_receipt_rule）',
                ])
                ->addColumn('shop_company_uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '门店company_uuid',
                ])
                ->addColumn('receipt_order_uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '收货单UUID',
                ])
                ->addColumn('receipt_order_no', 'string', [
                    'limit' => 100,
                    'default' => '',
                    'comment' => '收货单号',
                ])
                ->addColumn('receipt_erp_order_no', 'string', [
                    'limit' => 255,
                    'default' => '',
                    'comment' => '收货单ERP单号',
                ])
                ->addColumn('receipt_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '收货日期（当天0点时间戳，筛选用）',
                ])
                ->addColumn('create_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '创建时间',
                ])
                ->addColumn('update_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '更新时间',
                ])
                ->addColumn('delete_time', 'integer', [
                    'limit' => 10,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '删除时间',
                ])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['headquarter_company_uuid', 'receipt_time'], ['name' => 'idx_hq_date'])
                ->create();
        }
    }
}
