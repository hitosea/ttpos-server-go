<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateColumnSaleOrderProductAmountAga extends Migrator
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
        if ($table->hasColumn('product_fee')) {
            $table->renameColumn('product_fee', 'product_amount')
                  ->changeColumn('product_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '商品金额，(订单商品.总最终单价)之和。商品已含税时，该金额包括了税费。当商品未含税时，该金额不包括税费', 'after' => 'uuid'])
                  ->update();
        }
        $table->update();
    }
}
