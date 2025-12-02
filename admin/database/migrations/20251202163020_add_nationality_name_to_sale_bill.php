<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddNationalityNameToSaleBill extends Migrator
{
    /**
     * 为销售账单表新增国籍名称快照字段
     * Requirement: story-main-nationality-snapshot-fix
     * Purpose: 保存下单时的国籍名称快照，不随后台配置变更而改变
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_bill')) {
            $table = $this->table('sale_bill');
            
            // 检查字段是否不存在（幂等性）
            if (!$table->hasColumn('nationality_name')) {
                $table->addColumn(
                    'nationality_name', 
                    'text', 
                    [
                        'default' => '',
                        'comment' => '国籍名称快照（JSON），不随后台更新',
                        'after' => 'nationality_uuid'
                    ]
                );
                
                $table->update();
            }
        }
    }
}

