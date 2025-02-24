<?php

use think\migration\Migrator;
use think\migration\db\Column;

class addIsH5OrderProductToSaleOrderProduct extends Migrator
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
        if (!$table->hasColumn('is_h5_order_product')) {
            $table->addColumn(Column::tinyInteger('is_h5_order_product')->setDefault(0)->setComment('是否为扫码订单商品, 0-否 1-是')->setAfter('h5_order_product_uuid'));
            $table->update();
        }
    }
}
