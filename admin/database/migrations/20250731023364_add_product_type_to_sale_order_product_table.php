<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddProductTypeToSaleOrderProductTable extends Migrator
{
    public function change()
    {
        $table = $this->table('sale_order_product');
        if (!$table->hasColumn('product_type')) {
            $table->addColumn('product_type', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '商品类型, 0-商品 1-套餐',
                'after' => 'package_uuid'
            ])
            ->update();
        }
    }
} 