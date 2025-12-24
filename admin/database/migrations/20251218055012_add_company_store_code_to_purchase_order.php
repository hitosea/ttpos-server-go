<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddCompanyStoreCodeToPurchaseOrder extends Migrator
{
    /**
     * 为采购订单表添加公司店铺编码字段
     */
    public function change()
    {
        $table = $this->table('purchase_order');
        
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
