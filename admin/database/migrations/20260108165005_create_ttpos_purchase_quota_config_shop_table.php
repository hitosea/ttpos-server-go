<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTtposPurchaseQuotaConfigShopTable extends Migrator
{
    /**
     * Change Method.
     *
     * 创建品牌采购限购配置门店关联表
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('purchase_quota_config_shop')) {
            return;
        }

        $table = $this->table('purchase_quota_config_shop', [
            'id' => false,
            'primary_key' => ['id'],
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_general_ci',
            'comment' => '品牌采购限购配置门店关联',
        ]);

        $table
            ->addColumn('id', 'integer', [
                'limit' => 11,
                'signed' => false,
                'identity' => true,
                'comment' => '自增ID',
            ])
            ->addColumn('config_uuid', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'null' => false,
                'comment' => '限购配置UUID',
            ])
            ->addColumn('company_uuid', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'null' => false,
                'comment' => '公司UUID（门店UUID）',
            ])
            ->addColumn('create_time', 'integer', [
                'limit' => 10,
                'default' => 0,
                'comment' => '创建时间(时间戳)',
            ])
            ->addColumn('delete_time', 'integer', [
                'limit' => 10,
                'default' => 0,
                'comment' => '删除时间(时间戳)',
            ])
            ->addIndex(['config_uuid', 'company_uuid'], ['unique' => true, 'name' => 'uk_config_company'])
            ->addIndex(['config_uuid'], ['name' => 'idx_config'])
            ->addIndex(['company_uuid'], ['name' => 'idx_company'])
            ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])
            ->create();
    }
}

