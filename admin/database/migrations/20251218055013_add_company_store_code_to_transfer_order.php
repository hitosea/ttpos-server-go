<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddCompanyStoreCodeToTransferOrder extends Migrator
{

    // 迁移目标
    const TARGET = 'all';

    /**
     * 为调拨单表添加公司店铺编码字段
     */
    public function change()
    {
        $table = $this->table('transfer_order');
        
        // 检查字段是否已存在（幂等性）
        if (!$table->hasColumn('company_store_code')) {
            $table->addColumn(
                'company_store_code',
                'string',
                [
                    'limit' => 255,
                    'null' => false,
                    'default' => '',
                    'comment' => '公司店铺编码',
                    'after' => 'company_name'
                ]
            );
            
            $table->update();
        }
    }
}

