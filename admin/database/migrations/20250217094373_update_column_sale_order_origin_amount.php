<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateColumnSaleOrderOriginAmount extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     *
     * The following commands can be used in this method and Phinx will
     * automatically reverse them when rolling back:
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
        // 销售订单表
        $table = $this->table('sale_order');
        if ($table->hasColumn('product_original_fee')) {
            $table->renameColumn('product_original_fee', 'product_original_amount')
                  ->changeColumn('product_original_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '原始商品金额(折前价)。 商品原始金额=订单商品的销售价(折前价)之和。', 'after' => 'uuid'])
                  ->update();
        }
    }
}
