<?php

use think\migration\Migrator;
use think\migration\db\Column;

class ModifyPackageSubProductParamsNullableInSaleOrderProductTable extends Migrator
{
    /**
     * 迁移
     */
    public function change()
    {
        $table = $this->table('sale_order_product');
        
        // 检查表是否存在
        if (!$this->hasTable('sale_order_product')) {
            return;
        }

        // 检查package_sub_product_params字段是否存在
        if ($table->hasColumn('package_sub_product_params')) {
            // 修改字段去掉NOT NULL约束
            $table->changeColumn('package_sub_product_params', 'text', ['null' => true, 'comment' => '套餐子商品参数'])->update();
        }
    }
}
