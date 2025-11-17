<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddFieldsToTransferOrderApprovalTable extends Migrator
{

    // 迁移目标
    const TARGET = 'all';
    
    /**
     * 添加字段到调拨单审批流程表
     */
    public function change()
    {
        $table = $this->table('transfer_order_approval');
        
        // 检查字段是否已经存在，如果不存在则添加
        if (!$table->hasColumn('is_via_company_warehouse')) {
            $table->addColumn('is_via_company_warehouse', 'integer', ['limit' => 4, 'default' => 0, 'comment' => '是否通过公司仓库：0-否 1-是', 'after' => 'remark'])
                  ->update();
        }
        
        if (!$table->hasColumn('erpnext_company_abbr')) {
            $table->addColumn('erpnext_company_abbr', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERP公司简称', 'after' => 'is_via_company_warehouse'])
                  ->update();
        }
    }
}

