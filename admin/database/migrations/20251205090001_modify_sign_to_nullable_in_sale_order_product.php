<?php

use think\migration\Migrator;

class ModifySignToNullableInSaleOrderProduct extends Migrator
{
    /**
     * 修改 sale_order_product 表的 sign 字段，允许为空
     */
    public function change()
    {
        $table = $this->table('sale_order_product');
        
        if ($table->hasColumn('sign')) {
            // 修改字段为 text 类型，并允许为空 (null => true)
            // 注意：ThinkPHP Migration (Phinx) 中 text 类型长度通常不需要指定，或者指定 limit
            $table->changeColumn('sign', 'text', [
                'null' => true,
                'comment' => '商品签名,规格、属性、加料、是否改价、是否赠菜、送厨批次、销售价相同的商品签名相同,用于取消拆单时合并商品'
            ])->update();
        }
        if ($table->hasColumn('flavor_name')) {
            // 修改字段为 text 类型，并允许为空 (null => true)
            // 注意：ThinkPHP Migration (Phinx) 中 text 类型长度通常不需要指定，或者指定 limit
            $table->changeColumn('flavor_name', 'text', [
                'null' => true,
                'default' => '',
                'comment' => '规格名称'
            ])->update();
        }
    }
}

