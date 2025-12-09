<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddBuffetPackageNameToSaleOrderBuffetCustomerType extends Migrator
{
    /**
     * 为销售订单自助餐顾客类型表新增自助餐套餐名称快照字段
     * Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
     * Purpose: 保存下单时的自助餐套餐名称快照（JSON），不随后台配置变更而改变
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_order_buffet_customer_type')) {
            $table = $this->table('sale_order_buffet_customer_type');

            // 检查字段是否不存在（幂等性）
            if (!$table->hasColumn('buffet_package_name')) {
                $table->addColumn(
                    'buffet_package_name',
                    'text',
                    [
                        'default' => '',
                        'comment' => '自助餐套餐名称快照（JSON），不随后台更新',
                        'after' => 'buffet_package_uuid'
                    ]
                );

                $table->update();
            }
        }
    }
}

