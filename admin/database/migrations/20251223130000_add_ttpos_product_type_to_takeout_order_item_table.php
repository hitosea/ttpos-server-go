<?php

use think\migration\Migrator;

class AddTtposProductTypeToTakeoutOrderItemTable extends Migrator
{
    /**
     * 为 takeout_order_item 表添加 ttpos_product_type 字段
     */
    public function change()
    {
        $table = $this->table('takeout_order_item');
        
        // 检查字段是否存在，如果不存在则添加
        if (!$table->hasColumn('ttpos_product_type')) {
            $table->addColumn('ttpos_product_type', 'integer', [
                'limit' => 4,
                'signed' => false,
                'default' => 0,
                'comment' => 'TTPOS商品类型: 0-商品, 1-套餐',
                'after' => 'item_name',
            ])->update();
        }
    }
}

