<?php

use think\migration\Migrator;

class UpdateNameTypeToProductionOrderProduct extends Migrator
{
    /**
     * 更新生产订单商品表的名称类型
     */
    public function change()
    {


        $table = $this->table('production_order_product');
        
        // 检查表是否存在
        if (!$this->hasTable('production_order_product')) {
            return;
        }

        // 检查package_sub_product_params字段是否存在
        if ($table->hasColumn('name')) {
            // 修改字段去掉NOT NULL约束
            $table->changeColumn('name', 'text', ['null' => true, 'comment' => '名称'])->update();
        }
    }
}
