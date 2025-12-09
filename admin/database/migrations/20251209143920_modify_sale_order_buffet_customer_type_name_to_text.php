<?php
use think\migration\Migrator;

class ModifySaleOrderBuffetCustomerTypeNameToText extends Migrator
{
    /**
     * 修改 sale_order_buffet_customer_type 表的 name 字段类型为 TEXT
     * Requirement: story-main-buffet-customer-type-name-snapshot-fix
     * Purpose: 保存顾客类型名称快照（JSON），不随后台配置变更而改变
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_order_buffet_customer_type')) {
            $table = $this->table('sale_order_buffet_customer_type');

            // 检查字段是否存在，如果存在则修改类型
            if ($table->hasColumn('name')) {
                $table->changeColumn('name', 'text', [
                    'comment' => '顾客类型名称快照（JSON），不随后台更新'
                ])->update();
            }
        }
    }
}

