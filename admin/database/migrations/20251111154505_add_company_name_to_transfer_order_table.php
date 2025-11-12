<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddCompanyNameToTransferOrderTable extends Migrator
{

    // 迁移目标
    const TARGET = 'all';
    
    /**
     * 添加 company_name 字段到调拨单表
     */
    public function change()
    {
        // 给 transfer_order 表添加 company_name 字段
        $this->addCompanyNameToTransferOrder();
    }

    /**
     * 添加 company_name 字段到调拨单表
     */
    private function addCompanyNameToTransferOrder()
    {
        $table = $this->table('transfer_order');
        
        // 检查字段是否已存在
        if (!$table->hasColumn('company_name')) {
            $table->addColumn('company_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '所属公司名称', 'after' => 'company_uuid'])
                ->update();
        }
    }

}


