<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateSaleOrderBuffetCustomerTypePrice extends Migrator
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
        $table = $this->table('sale_order_buffet_customer_type');
        if (!$table->hasColumn('price')) {
            $table->addColumn('price', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '价格,下单后价格不受后台改变', 'after' => 'num'])
                  ->update();
        }
        if ($table->hasColumn('buffet_customer_type_uuid')) {
            $table->renameColumn('buffet_customer_type_uuid', 'buffet_customer_type_price_uuid')
                  ->update();
        }
    }
}
