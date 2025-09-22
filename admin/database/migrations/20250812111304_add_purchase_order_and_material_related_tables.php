<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddPurchaseOrderAndMaterialRelatedTables extends Migrator
{
    /**
     * 迁移
     */
    public function change()
    {
        // 1. 修改 material 表，添加新字段
        $this->updateMaterialTable();
        
        // 2. 创建 material_unit 表
        $this->createMaterialUnitTable();
        
        // 3. 创建 purchase_order 表
        $this->createPurchaseOrderTable();
        
        // 4. 创建 purchase_order_item 表
        $this->createPurchaseOrderItemTable();
        
        // 5. 创建 receipt_order 表
        $this->createReceiptOrderTable();
        
        // 6. 创建 receipt_order_item 表
        $this->createReceiptOrderItemTable();
        
        // 7. 创建 product_bom_card 表
        $this->createProductBomCardTable();
        
        // 8. 修改 related_material 表，添加新字段
        $this->updateRelatedMaterialTable();
    }

    /**
     * 修改原料信息表，添加新字段
     */
    private function updateMaterialTable()
    {
        $table = $this->table('material');
        
        // 检查字段是否已存在，避免重复添加
        if (!$table->hasColumn('code')) {
            $table->addColumn('code', 'string', ['limit' => 255, 'default' => '', 'comment' => '原料编码', 'after' => 'name']);
        }
        
        if (!$table->hasColumn('valuation')) {
            $table->addColumn('valuation', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '估值率', 'after' => 'code']);
        }
        
        if (!$table->hasColumn('init_stock')) {
            $table->addColumn('init_stock', 'decimal', ['precision' => 14, 'scale' => 4, 'default' => 0.0000, 'comment' => '期初库存', 'after' => 'valuation']);
        }
        
        if (!$table->hasColumn('purchase_unit_uuid')) {
            $table->addColumn('purchase_unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '采购单位ID', 'after' => 'unit_uuid']);
        }
        
        if (!$table->hasColumn('cost_unit_uuid')) {
            $table->addColumn('cost_unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '成本单位ID', 'after' => 'purchase_unit_uuid']);
        }
        
        $table->update();
    }

    /**
     * 创建原料单位表
     */
    private function createMaterialUnitTable()
    {
        if (!$this->hasTable('material_unit')) {
            $table = $this->table('material_unit', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '原料单位表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
            ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '原料单位ID'])
            ->addColumn('name', 'string', ['limit' => 255, 'default' => '', 'comment' => '原料单位名称'])
            ->addColumn('unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '单位ID'])
            ->addColumn('conversion_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1, 'comment' => '转换率'])
            ->addColumn('from_unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '来源单位ID. 来源单位为克，则转换率为1000，该原料单位为千克'])
            ->addColumn('is_default', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '是否为基准单位, 0-否 1-是'])
            ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间(时间戳)'])
            ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间(时间戳)'])
            ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间(时间戳)'])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
            ->create();
        }
    }

    /**
     * 创建采购申请表
     */
    private function createPurchaseOrderTable()
    {
        if (!$this->hasTable('purchase_order')) {
            $table = $this->table('purchase_order', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '采购申请表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false,'default' => 0,'comment' => '采购申请ID'])
                ->addColumn('order_no', 'string', ['limit' => 255, 'default' => '', 'comment' => '单号'])
                ->addColumn('order_type', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '申请类型, 0-仓库调拨'])
                ->addColumn('status', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '状态, 0-待提交 1-待审核 2-已通过 3-已驳回 4-部分收货 5-全部收货 6-待总部审核'])
                ->addColumn('num', 'decimal', ['precision' => 14, 'scale' => 4, 'default' => 0.0000, 'comment' => '物资数量，每种物品算一个'])
                ->addColumn('order_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '单据日期，采购单提交的时间（时间戳）'])
                ->addColumn('applicant_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '申请人ID'])
                ->addColumn('applicant_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '申请人姓名'])
                ->addColumn('approver_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '审批人ID'])
                ->addColumn('approver_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '审批人姓名'])
                ->addColumn('expect_arrival_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '期望到货日期（时间戳）'])
                ->addColumn('pass_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '通过时间（时间戳）'])
                ->addColumn('reject_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '驳回时间（时间戳）'])
                ->addColumn('first_receive_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '第一次收货时间（时间戳），从“已通过”状态变成“部分收货”状态的时间'])
                ->addColumn('final_receive_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '最终收货时间（时间戳），从“部分收货”状态变成“全部收货”状态的时间'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->create();
        }
    }

    /**
     * 创建采购申请物品表
     */
    private function createPurchaseOrderItemTable()
    {
        if (!$this->hasTable('purchase_order_item')) {
            $table = $this->table('purchase_order_item', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '采购申请物品表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
            ->addColumn('uuid', 'biginteger', ['signed' => false,'default' => 0,'comment' => '采购申请物品ID'])
            ->addColumn('purchase_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '采购申请ID'])
            ->addColumn('material_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '物品编码, 提交采购时记录后不再修改'])
            ->addColumn('material_name', 'text', ['default' => '', 'comment' => '物品名称JSON, 提交采购时记录后不再修改'])
            ->addColumn('material_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '物品ID'])
            ->addColumn('num', 'decimal', [ 'precision' => 14,'scale' => 4,'default' => 0.0000,'comment' => '申请数量'])
            ->addColumn('arrival_num', 'decimal', ['precision' => 14, 'scale' => 4, 'default' => 0.0000, 'comment' => '到货数量'])
            ->addColumn('unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '采购单位ID'])
            ->addColumn('unit_name', 'text', ['default' => '', 'comment' => '采购单位名称JSON, 提交采购时记录后不再修改'])
            ->addColumn('unit_conversion_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1, 'comment' => '基准单位转换率。申请数量*转换率=基准单位申请数量'])
            ->addColumn('base_unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '基准单位ID'])
            ->addColumn('base_unit_name', 'text', ['default' => '', 'comment' => '基准单位名称JSON, 提交采购时记录后不再修改'])
            ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间(时间戳)'])
            ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间(时间戳)'])
            ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间(时间戳)'])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
            ->create();
        }
    }

    /**
     * 创建收货单表
     */
    private function createReceiptOrderTable()
    {
        if (!$this->hasTable('purchase_receipt_order')) {
            $table = $this->table('purchase_receipt_order', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '收货单表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
            ->addColumn('uuid', 'biginteger', ['signed' => false,'default' => 0,'comment' => '收货单ID'])
            ->addColumn('order_no', 'string', ['limit' => 255, 'default' => '', 'comment' => '单号'])
            ->addColumn('status', 'integer', ['limit' => 10,'default' => 0,'comment' => '状态, 0-待收货 1-已收货 2-已取消'])
            ->addColumn('purchase_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '采购申请ID'])
            ->addColumn('purchase_order_no', 'string', ['limit' => 255, 'default' => '', 'comment' => '采购申请单号'])
            ->addColumn('num', 'decimal', ['precision' => 14, 'scale' => 4, 'default' => 0.0000, 'comment' => '物资数量，每种物品算一个'])
            ->addColumn('expect_arrival_time', 'integer', ['signed' => false,'limit' => 10,'default' => 0,'comment' => '期望到货日期（时间戳），与采购申请单的期望到货日期一致'])
            ->addColumn('receive_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '收货时间（时间戳）'])
            ->addColumn('cancel_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '取消时间（时间戳）'])
            ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间(时间戳)'])
            ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间(时间戳)'])
            ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间(时间戳)'])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
            ->create();
        }
    }

    /**
     * 创建收货单物品表
     */
    private function createReceiptOrderItemTable()
    {
        if (!$this->hasTable('purchase_receipt_order_item')) {
            $table = $this->table('purchase_receipt_order_item', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '收货单物品表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
            ->addColumn('uuid', 'biginteger', ['signed' => false,'default' => 0,'comment' => '收货单物品ID'])
            ->addColumn('receipt_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '收货单ID'])
            ->addColumn('purchase_order_item_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '采购申请物品ID'])
            ->addColumn('material_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '物品编码, 提交采购时记录后不再修改'])
            ->addColumn('material_name', 'text', ['default' => '', 'comment' => '物品名称JSON, 提交采购时记录后不再修改'])
            ->addColumn('material_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '物品ID'])
            ->addColumn('num', 'decimal', ['precision' => 14,'scale' => 4,'default' => 0.0000,'comment' => '收货数量'])
            ->addColumn('unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '单位ID'])
            ->addColumn('unit_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '单位名称, 提交采购时记录后不再修改'])
            ->addColumn('unit_conversion_rate', 'decimal', ['precision' => 12,'scale' => 4,'default' => 1,'comment' => '单位转换率。收货数量*转换率=基准单位收货数量'])
            ->addColumn('base_unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '基准单位ID'])
            ->addColumn('base_unit_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '基准单位名称, 确认收货时记录后不再修改'])
            ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间(时间戳)'])
            ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间(时间戳)'])
            ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间(时间戳)'])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
            ->create();
        }
    }

    /**
     * 创建成本卡表
     */
    private function createProductBomCardTable()
    {
        if (!$this->hasTable('product_bom_card')) {
            $table = $this->table('product_bom_card', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '成本卡表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
            ->addColumn('uuid', 'biginteger', ['signed' => false,'default' => 0,'comment' => '成本卡ID' ])
            ->addColumn('name', 'string', ['limit' => 255, 'default' => '', 'comment' => '名称'])
            ->addColumn('multi_language_name_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '多语言名称ID'])
            ->addColumn('num', 'decimal', ['precision' => 14,'scale' => 4,'default' => 0.0000,'comment' => '加工份数'])
            ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间(时间戳)'])
            ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间(时间戳)'])
            ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间(时间戳)'])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
            ->create();
        }
    }

    /**
     * 修改关联材料表，添加新字段
     */
    private function updateRelatedMaterialTable()
    {
        $table = $this->table('related_material');
        
        // 检查字段是否已存在，避免重复添加
        if (!$table->hasColumn('unit_uuid')) {
            $table->addColumn('unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '单位ID,物品单位', 'after' => 'num']);
        }
        
        if (!$table->hasColumn('unit_name')) {
            $table->addColumn('unit_name', 'text', ['default' => '', 'comment' => '单位名称JSON,物品单位名称', 'after' => 'unit_uuid']);
        }
        
        if (!$table->hasColumn('base_unit_uuid')) {
            $table->addColumn('base_unit_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '基准单位ID,物品基准单位', 'after' => 'unit_name']);
        }
        
        if (!$table->hasColumn('base_unit_name')) {
            $table->addColumn('base_unit_name', 'text', ['default' => '', 'comment' => '基准单位名称JSON,物品基准单位名称', 'after' => 'base_unit_uuid']);
        }
        
        if (!$table->hasColumn('base_unit_conversion_rate')) {
            $table->addColumn('base_unit_conversion_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1, 'comment' => '基准单位转换率。用量*转换率=基准单位用量', 'after' => 'base_unit_name']);
        }
        
        $table->update();
    }
}
