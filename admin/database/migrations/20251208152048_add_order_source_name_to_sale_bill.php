<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddOrderSourceNameToSaleBill extends Migrator
{
    /**
     * 为销售账单表新增外卖来源名称快照字段
     * Requirement: story-main-order-source-snapshot-fix
     * Purpose: 保存下单时的外卖来源名称快照（JSON），不随后台配置变更而改变
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_bill')) {
            $table = $this->table('sale_bill');
            
            // 检查字段是否不存在（幂等性）
            if (!$table->hasColumn('order_source_name')) {
                $table->addColumn(
                    'order_source_name', 
                    'text', 
                    [
                        'default' => '',
                        'comment' => '外卖来源名称快照（JSON），不随后台更新',
                        'after' => 'order_source_uuid'
                    ]
                );
                
                $table->update();
            }
        }
    }
}

