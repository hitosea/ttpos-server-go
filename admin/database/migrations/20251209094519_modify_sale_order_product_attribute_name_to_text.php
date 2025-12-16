<?php
use think\migration\Migrator;

class ModifySaleOrderProductAttributeNameToText extends Migrator
{
    /**
     * 修改 sale_order_product_attribute 表的 name 字段类型为 TEXT
     * Requirement: story-main-product-attribute-snapshot-fix
     * Purpose: 保存商品属性名称快照（JSON），不随后台配置变更而改变
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_order_product_attribute')) {
            $table = $this->table('sale_order_product_attribute');

            // 检查字段是否存在，如果存在则修改类型
            if ($table->hasColumn('name')) {
                $table->changeColumn('name', 'text', [
                    'comment' => '商品属性名称快照（JSON），不随后台更新'
                ])->update();
            }
        }
    }
}

