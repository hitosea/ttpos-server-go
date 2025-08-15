<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateProductBomCardLogTable extends Migrator
{
    /**
     * 创建成本卡日志表
     */
    public function change()
    {
         // 检查表是否已存在
         if ($this->hasTable('product_bom_card_log')) {
            return;
        }

        $table = $this->table('product_bom_card_log', [
            'id' => false,
            'primary_key' => ['id'],
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_unicode_ci',
            'comment' => '成本卡日志表'
        ]);

        $table->addColumn('id', 'integer', [
            'identity' => true,
            'signed' => false,
            'limit' => 11,
            'comment' => '自增ID'
        ])
        ->addColumn('uuid', 'biginteger', [
            'signed' => false,
            'default' => 0,
            'comment' => '成本卡日志ID'
        ])
        ->addColumn('product_bom_card_uuid', 'biginteger', [
            'signed' => false,
            'default' => 0,
            'comment' => '成本卡ID'
        ])
        ->addColumn('product_bom_card_name', 'string', [
            'limit' => 255,
            'default' => '',
            'comment' => '成本卡名称'
        ])
        ->addColumn('related_uuid', 'biginteger', [
            'signed' => false,
            'default' => 0,
            'comment' => '关联ID'
        ])
        ->addColumn('related_name', 'string', [
            'limit' => 255,
            'default' => '',
            'comment' => '关联名称'
        ])
        ->addColumn('data', 'text', [
            'null' => true,
            'comment' => '成本卡数据'
        ])
        ->addColumn('staff_uuid', 'biginteger', [
            'signed' => false,
            'default' => 0,
            'comment' => '操作员工UUID'
        ])
        ->addColumn('operation_type', 'integer', [
            'signed' => false,
            'limit' => 10,
            'default' => 0,
            'comment' => '操作类型'
        ])
        ->addColumn('create_time', 'integer', [
            'signed' => false,
            'limit' => 10,
            'default' => 0,
            'comment' => '创建时间(时间戳)'
        ])
        ->addColumn('update_time', 'integer', [
            'signed' => false,
            'limit' => 10,
            'default' => 0,
            'comment' => '更新时间(时间戳)'
        ])
        ->addColumn('delete_time', 'integer', [
            'signed' => false,
            'limit' => 10,
            'default' => 0,
            'comment' => '删除时间(时间戳)'
        ])
        ->addIndex(['product_bom_card_uuid'], [
            'name' => 'idx_product_bom_card_uuid'
        ])
        ->addIndex(['staff_uuid'], [
            'name' => 'idx_staff_uuid'
        ])
        ->addIndex(['operation_type'], [
            'name' => 'idx_operation_type'
        ])
        ->addIndex(['create_time'], [
            'name' => 'idx_create_time'
        ])
        ->create(); 
    }
    
    
} 