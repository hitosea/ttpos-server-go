<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnIsBuffetSaleOrderProduct extends Migrator
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
        $table = $this->table('sale_order_product');
        if (!$table->hasColumn('is_buffet')) {
            $table->addColumn('is_buffet', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '是否为自助餐商品,0-否 1-是. 如果是自助餐商品，则sale_price为0', 'after' => 'change_price_time'])->update();
        }
    }
}
