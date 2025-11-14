<?php

use think\migration\Migrator;

class AddH5OrderUuidIndexToSaleOrderProductTable extends Migrator
{
    /**
     * 添加h5_order_uuid索引到销售订单商品表
     */
    public function change()
    {
        $table = $this->table('sale_order_product');

        // 检查索引是否已经存在
        if ($table->hasIndex(['h5_order_uuid'])) {
            return;
        }

        // 添加h5_order_uuid索引
        $table->addIndex(['h5_order_uuid'], ['name' => 'idx_h5_order_uuid'])
            ->update();
    }
}
