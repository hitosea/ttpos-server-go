<?php

use think\migration\Migrator;
use think\migration\db\Column;

/**
 * 创建 ProductMap 表（支持多平台商品映射）
 */
class CreateProductMap extends Migrator
{
    /**
     * 执行迁移
     */
    public function change()
    {
        // 创建 product_map 表
        if (!$this->hasTable('product_map')) {
            $table = $this->table('product_map', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_general_ci',
                'comment' => '外卖平台商品映射表'
            ]);

            $table
                // 主键ID
                ->addColumn('id', 'integer', ['limit' => 10, 'signed' => false, 'identity' => true, 'comment' => '主键ID'])
                // 唯一标识
                ->addColumn('uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'null' => false,
                    'comment' => '唯一标识'
                ])
                // 来源平台
                ->addColumn('source', 'string', [
                    'limit' => 50,
                    'default' => '',
                    'comment' => '来源平台(grab/foodpanda/lineman等)'
                ])
                // 来源平台商品ID
                ->addColumn('source_product_id', 'string', [
                    'limit' => 500,
                    'default' => '',
                    'comment' => '来源平台商品唯一ID'
                ])
                // 店内商品包UUID
                ->addColumn('product_package_uuid', 'biginteger', [
                    'limit' => 20,
                    'signed' => false,
                    'default' => 0,
                    'comment' => '店内商品包UUID'
                ])
                // 状态
                ->addColumn('status', 'integer', [
                    'limit' => 11,
                    'default' => 1,
                    'comment' => '状态 1-有效'
                ])
                // 同步时间
                ->addColumn('sync_time', 'biginteger', [
                    'limit' => 20,
                    'default' => 0,
                    'comment' => '同步时间戳'
                ])
                // 时间戳
                ->addColumn('create_time', 'biginteger', [
                    'limit' => 20,
                    'default' => 0,
                    'comment' => '创建时间'
                ])
                ->addColumn('update_time', 'biginteger', [
                    'limit' => 20,
                    'default' => 0,
                    'comment' => '更新时间'
                ])
                ->addColumn('delete_time', 'biginteger', [
                    'limit' => 20,
                    'default' => 0,
                    'comment' => '删除时间'
                ])
                // 添加索引
                ->addIndex(['product_package_uuid'], [
                    'name' => 'idx_product_package_uuid'
                ])
                ->addIndex(['delete_time'], [
                    'name' => 'idx_delete_time'
                ])
                ->create();
        }
    }
}

