<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddPreBatchPrintTimeToSaleOrderProductTable extends Migrator
{
    /**
     * 添加 pre_batch_print_time 字段到 ttpos_sale_order_product 表
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_order_product')) {
            $table = $this->table('sale_order_product');

            // 检查字段是否已存在，如果不存在则添加
            if (!$table->hasColumn('pre_batch_print_time')) {
                $table->addColumn('pre_batch_print_time', 'integer', [
                    'null' => false,
                    'default' => 0,
                    'comment' => '预先分批打印送厨单的时间(时间戳)，0表示未打印',
                    'after' => 'batch_time',
                ])->update();
            }
        }
    }
}

