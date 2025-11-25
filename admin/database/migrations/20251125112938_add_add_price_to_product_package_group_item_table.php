<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddAddPriceToProductPackageGroupItemTable extends Migrator
{
    /**
     * 添加加价字段到 ttpos_product_package_group_item 表
     */
    public function change()
    {
        $table = $this->table('product_package_group_item');

        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('add_price')) {
            $table->addColumn('add_price', 'decimal', [
                'precision' => 22,
                'scale' => 4,
                'null' => false,
                'default' => 0.00,
                'comment' => '加价金额，表示该商品需要加价多少钱',
                'after' => 'sort'
            ])->update();
        }
    }
}

