<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddCopyNumToSaleOrderProductTable extends Migrator
{
    /**
     * 添加 copy_num 字段到 ttpos_sale_order_product 表
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_order_product')) {
            $table = $this->table('sale_order_product');

            // 检查字段是否已存在，如果不存在则添加
            if (!$table->hasColumn('copy_num')) {
                $table->addColumn('copy_num', 'decimal', ['precision' => 12, 'scale' => 4, 'null' => false, 'default' => 0, 'comment' => '表示该子商品在分组中被选择多少份', 'after' => 'unit_num'])->update();
            }
        }
    }
}


