<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTransferOrderTables extends Migrator
{

    // 迁移目标
    const TARGET = 'all';
    
    /**
     * 创建调拨单相关表
     */
    public function change()
    {
        // 1. 创建调拨单主表
        $this->createTransferOrderTable();
    
        // 2. 创建调拨单审批流程表
        $this->createTransferOrderApprovalTable();
        
        // 3. 创建调拨单操作日志表
        $this->createTransferOrderLogTable();
    }

    /**
     * 创建调拨单主表
     */
    private function createTransferOrderTable()
    {
        if (!$this->hasTable('transfer_order')) {
            $table = $this->table('transfer_order', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '调拨单主表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '主键UUID'])
                ->addColumn('company_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '所属公司UUID'])
                ->addColumn('headquarter_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '总部UUID'])
                ->addColumn('order_no', 'string', ['limit' => 255, 'default' => '', 'comment' => '单据编号TR+12位数字'])
                ->addColumn('erp_order_no', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERP调拨单号（销售单号）'])
                ->addColumn('transfer_type', 'integer', ['limit' => 4, 'default' => 1, 'comment' => '调拨类型：1-调入 2-调出'])
                ->addColumn('sender_company_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '发货门店UUID'])
                ->addColumn('sender_company_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '发货门店名称'])
                ->addColumn('receiver_company_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '收货门店UUID'])
                ->addColumn('receiver_company_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '收货门店名称'])
                ->addColumn('out_warehouse_erp_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '出库仓库ERP编码'])
                ->addColumn('out_warehouse_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '出库仓库名称'])
                ->addColumn('in_warehouse_erp_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '入库仓库ERP编码'])
                ->addColumn('in_warehouse_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '入库仓库名称'])
                ->addColumn('order_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '单据日期（提交时间戳）'])
                ->addColumn('submit_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '提交时间'])
                ->addColumn('status', 'integer', ['limit' => 4, 'default' => 0, 'comment' => '状态：0-待提交 1-待审核 2-已驳回 3-待收货 4-已完成'])
                ->addColumn('creator_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '创建人UUID'])
                ->addColumn('creator_name', 'string', ['limit' => 100, 'default' => '', 'comment' => '创建人姓名'])
                ->addColumn('next_approval_company_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '下一个审批门店UUID'])
                ->addColumn('next_approval_company_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '下一个审批门店名称'])
                ->addColumn('remark', 'text', ['comment' => '备注'])
                ->addColumn('item_count', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '物品种类数量'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['company_uuid'], ['name' => 'idx_company_uuid'])
                ->addIndex(['headquarter_uuid'], ['name' => 'idx_headquarter_uuid'])
                ->addIndex(['order_no'], ['name' => 'idx_order_no'])
                ->addIndex(['sender_company_uuid'], ['name' => 'idx_sender_company_uuid'])
                ->addIndex(['receiver_company_uuid'], ['name' => 'idx_receiver_company_uuid'])
                ->addIndex(['status'], ['name' => 'idx_status'])
                ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])
                ->create();
        }
    }

   
    /**
     * 创建调拨单审批流程表
     */
    private function createTransferOrderApprovalTable()
    {
        if (!$this->hasTable('transfer_order_approval')) {
            $table = $this->table('transfer_order_approval', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '调拨单审批流程表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '主键UUID'])
                ->addColumn('transfer_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '调拨单UUID'])
                ->addColumn('company_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '所属公司UUID'])
                ->addColumn('headquarter_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '总部UUID'])
                ->addColumn('approval_type', 'string', ['limit' => 50, 'default' => '', 'comment' => '审批类型：initiator-发起人公司 sender-发货门店 sender_parent-发货门店上级 receiver_parent-收货门店上级 receiver-收货门店'])
                ->addColumn('approval_company_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '审批门店UUID'])
                ->addColumn('approval_company_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '审批门店名称'])
                ->addColumn('sequence', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '审批顺序，从1开始'])
                ->addColumn('status', 'integer', ['limit' => 4, 'default' => 0, 'comment' => '审批状态：0-待审批 1-已通过 2-已驳回 3-已跳过'])
                ->addColumn('approver_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '审批人UUID'])
                ->addColumn('approver_name', 'string', ['limit' => 100, 'default' => '', 'comment' => '审批人姓名'])
                ->addColumn('approve_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '审批时间'])
                ->addColumn('reject_reason', 'text', ['comment' => '驳回原因'])
                ->addColumn('is_required', 'integer', ['limit' => 4, 'default' => 1, 'comment' => '是否必须审批：0-否 1-是'])
                ->addColumn('remark', 'text', ['comment' => '备注'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['transfer_order_uuid'], ['name' => 'idx_transfer_order_uuid'])
                ->addIndex(['approval_company_uuid'], ['name' => 'idx_approval_company_uuid'])
                ->addIndex(['status'], ['name' => 'idx_status'])
                ->addIndex(['sequence'], ['name' => 'idx_sequence'])
                ->create();
        }
    }

    /**
     * 创建调拨单操作日志表
     */
    private function createTransferOrderLogTable()
    {
        if (!$this->hasTable('transfer_order_log')) {
            $table = $this->table('transfer_order_log', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '调拨单操作日志表'
            ]);
            
            $table->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '主键UUID'])
                ->addColumn('transfer_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '调拨单UUID'])
                ->addColumn('company_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '所属公司UUID'])
                ->addColumn('action', 'string', ['limit' => 50, 'default' => '', 'comment' => '操作动作：create/submit/approve/reject/receive'])
                ->addColumn('action_desc', 'string', ['limit' => 255, 'default' => '', 'comment' => '操作描述'])
                ->addColumn('old_status', 'integer', ['limit' => 4, 'default' => 0, 'comment' => '操作前状态'])
                ->addColumn('new_status', 'integer', ['limit' => 4, 'default' => 0, 'comment' => '操作后状态'])
                ->addColumn('operator_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '操作人UUID'])
                ->addColumn('operator_name', 'string', ['limit' => 100, 'default' => '', 'comment' => '操作人姓名'])
                ->addColumn('operator_role', 'string', ['limit' => 50, 'default' => '', 'comment' => '操作人角色：sender/sender_parent/receiver_parent/receiver'])
                ->addColumn('content', 'text', ['comment' => '操作内容详情JSON'])
                ->addColumn('remark', 'text', ['comment' => '备注'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'limit' => 10, 'default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['transfer_order_uuid'], ['name' => 'idx_transfer_order_uuid'])
                ->addIndex(['operator_uuid'], ['name' => 'idx_operator_uuid'])
                ->addIndex(['action'], ['name' => 'idx_action'])
                ->create();
        }
    }
}

