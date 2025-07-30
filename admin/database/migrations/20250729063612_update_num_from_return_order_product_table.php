<?php

use Phinx\Migration\AbstractMigration;

class UpdateNumFromReturnOrderProductTable extends AbstractMigration
{
    /**
     * 修改 ttpos_return_order_product 表的 num 字段类型为 decimal(12,8)
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('return_order_product')) {
            return;
        }

        $table = $this->table('return_order_product');
        
        // 检查字段是否存在
        if (!$table->hasColumn('num')) {
            return;
        }

        // 修改 num 字段类型为 decimal(12,8)
        $table->changeColumn('num', 'decimal', [
            'precision' => 12,
            'scale' => 8,
            'null' => false,
            'default' => 0.00000000,
            'comment' => '商品数量,退货的商品数量',
        ]);

        $table->save();
    }
}
