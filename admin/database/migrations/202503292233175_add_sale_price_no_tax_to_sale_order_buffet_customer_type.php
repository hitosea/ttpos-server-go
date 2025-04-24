<?php

use think\facade\Db;
use think\facade\Config;
use think\migration\Migrator;

class AddSalePriceNoTaxToSaleOrderBuffetCustomerType extends Migrator
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
        $table = $this->table('sale_order_buffet_customer_type');
        if (!$table->hasColumn('sale_price_no_tax')) {
            $table->addColumn('sale_price_no_tax', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '销售价,未含税价格（折前）', 'after' => 'sale_price']);
            $table->update();
        }
    }
}
