<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreatePurchaseOrderItemUnitTable extends Migrator
{
    /**
     * 创建采购申请物品单位表
     */
    public function change()
    {
        // 检查表是否已存在
        if (!$this->hasTable('purchase_order_item_unit')) {
            $table = $this->table('purchase_order_item_unit', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '采购申请物品单位表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '采购申请物品单位ID'])
                ->addColumn('item_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'ItemID'])
                ->addColumn('num', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0.0000, 'comment' => '数量'])
                ->addColumn('arrival_num', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0.0000, 'comment' => '到货数量'])
                ->addColumn('unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '单位ID'])
                ->addColumn('unit_name', 'text', ['null' => true, 'comment' => '单位名称'])
                ->addColumn('unit_conversion_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1.0000, 'comment' => '基准单位转换率。申请数量*转换率=基准单位申请数量'])
                ->addColumn('base_unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '基准单位ID'])
                ->addColumn('base_unit_name', 'text', ['null' => true, 'comment' => '基准单位名称'])
                ->addColumn('erpnext_uom', 'string', ['limit' => 255, 'null' => true, 'comment' => 'ERPNext单位'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['item_uuid'], ['name' => 'idx_item_uuid'])
                ->create();
        }

        // 检查表是否已存在
        if (!$this->hasTable('purchase_receipt_order_item_unit')) {
            $table = $this->table('purchase_receipt_order_item_unit', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '收货单物品单位表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '收货单物品单位ID'])
                ->addColumn('num', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0.0000, 'comment' => '数量'])
                ->addColumn('arrival_num', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0.0000, 'comment' => '到货数量'])
                ->addColumn('item_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'ItemID'])
                ->addColumn('unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '单位ID'])
                ->addColumn('unit_name', 'text', ['null' => true, 'comment' => '单位名称'])
                ->addColumn('unit_conversion_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1.0000, 'comment' => '基准单位转换率。申请数量*转换率=基准单位申请数量'])
                ->addColumn('base_unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '基准单位ID'])
                ->addColumn('base_unit_name', 'text', ['null' => true, 'comment' => '基准单位名称'])
                ->addColumn('erpnext_uom', 'string', ['limit' => 255, 'null' => true, 'comment' => 'ERPNext单位'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['item_uuid'], ['name' => 'idx_item_uuid'])
                ->create();
        }
    }
}

