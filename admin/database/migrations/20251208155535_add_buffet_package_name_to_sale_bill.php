<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddBuffetPackageNameToSaleBill extends Migrator
{
    /**
     * 为销售账单表新增自助餐名称快照字段
     * Requirement: story-main-buffet-package-name-snapshot-fix
     * Purpose: 保存下单时的自助餐名称快照（JSON），不随后台配置变更而改变
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_bill')) {
            $table = $this->table('sale_bill');

            // 检查字段是否不存在（幂等性）
            if (!$table->hasColumn('buffet_package1_name')) {
                $table->addColumn(
                    'buffet_package1_name',
                    'text',
                    [
                        'default' => '',
                        'comment' => '自助餐套餐1名称快照（JSON），不随后台更新',
                        'after' => 'buffet_package1_uuid'
                    ]
                );
            }

            // 检查字段是否不存在（幂等性）
            if (!$table->hasColumn('buffet_package2_name')) {
                $table->addColumn(
                    'buffet_package2_name',
                    'text',
                    [
                        'default' => '',
                        'comment' => '自助餐套餐2名称快照（JSON），不随后台更新',
                        'after' => 'buffet_package2_uuid'
                    ]
                );
            }

            $table->update();
        }
    }
}

