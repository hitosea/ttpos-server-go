<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddPriceToProductPackageTakeoutTable extends Migrator
{
    /**
     * 为 ttpos_product_package_takeout 表添加 price 字段
     * 用于存储外卖商品价格（套餐价格）
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('product_package_takeout')) {
            return;
        }

        $table = $this->table('product_package_takeout');

        // 检查字段是否已存在
        if (!$table->hasColumn('price')) {
            $table->addColumn('price', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0.0000, 'comment' => '外卖商品价格(套餐价格)', 'after' => 'product_type'])
                ->update();
        }
    }
}

