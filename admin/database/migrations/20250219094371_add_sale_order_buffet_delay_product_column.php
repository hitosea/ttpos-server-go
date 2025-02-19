<?php

use think\migration\Migrator;

class AddSaleOrderBuffetDelayProductColumn extends Migrator
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
        $table = $this->table('sale_order_buffet_delay_product');
        if (!$table->hasColumn('name')) {
            $table->addColumn('name', 'string', ['default' => '', 'comment' => '自助餐加钟商品名称，下单时固定不受后台改变', 'after' => 'num'])
                  ->update();
        }
        if (!$table->hasColumn('price')) {
            $table->addColumn('price', 'decimal', ['default' => 0, 'comment' => '价格,下单时固定不受后台改变，结账时再检查是否改变', 'after' => 'name'])
                  ->update();
        }
        $table->update();
    }
}
