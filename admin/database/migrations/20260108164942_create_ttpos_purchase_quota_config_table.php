<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTtposPurchaseQuotaConfigTable extends Migrator
{
    /**
     * Change Method.
     *
     * 创建品牌采购限购配置主表
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('purchase_quota_config')) {
            return;
        }

        $table = $this->table('purchase_quota_config', [
            'id' => false,
            'primary_key' => ['id'],
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_general_ci',
            'comment' => '品牌采购限购配置',
        ]);

        $table
            ->addColumn('id', 'integer', [
                'limit' => 11,
                'signed' => false,
                'identity' => true,
                'comment' => '自增ID',
            ])
            ->addColumn('uuid', 'biginteger', [
                'limit' => 20,
                'signed' => false,
                'default' => 0,
                'comment' => '绑定记录ID（雪花算法生成）',
            ])
            ->addColumn('material_code', 'string', [
                'limit' => 100,
                'null' => false,
                'comment' => '物品编码',
            ])
            ->addColumn('unit_code', 'string', [
                'limit' => 50,
                'null' => false,
                'comment' => '限购单位编码',
            ])
            ->addColumn('quota_limit', 'decimal', [
                'precision' => 10,
                'scale' => 2,
                'default' => '0.00',
                'comment' => '限购数量',
            ])
            ->addColumn('apply_to_all_shops', 'integer', [
                'limit' => 4,
                'default' => 1,
                'comment' => '是否应用到全部店铺: 1=是 0=否',
            ])
            ->addColumn('period_type', 'integer', [
                'limit' => 4,
                'default' => 0,
                'comment' => '周期类型: 0=按天(默认) 1=月度',
            ])
            ->addColumn('strict_mode', 'integer', [
                'limit' => 4,
                'default' => 1,
                'comment' => '超限策略: 1=严格拒绝',
            ])
            ->addColumn('config_source', 'integer', [
                'limit' => 4,
                'default' => 1,
                'comment' => '配置来源: 1=门店 2=总部',
            ])
            ->addColumn('status', 'integer', [
                'limit' => 4,
                'default' => 1,
                'comment' => '状态: 1=启用 0=禁用',
            ])
            ->addColumn('create_time', 'integer', [
                'limit' => 10,
                'default' => 0,
                'comment' => '创建时间(时间戳)',
            ])
            ->addColumn('update_time', 'integer', [
                'limit' => 10,
                'default' => 0,
                'comment' => '更新时间(时间戳)',
            ])
            ->addColumn('delete_time', 'integer', [
                'limit' => 10,
                'default' => 0,
                'comment' => '删除时间(时间戳)',
            ])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
            ->addIndex(['material_code'], ['name' => 'idx_material'])
            ->addIndex(['status'], ['name' => 'idx_status'])
            ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])
            ->create();
    }
}

