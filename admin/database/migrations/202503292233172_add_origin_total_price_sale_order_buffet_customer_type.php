<?php

use think\migration\Migrator;

class AddOriginTotalPriceSaleOrderBuffetCustomerType extends Migrator
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
        if (!$table->hasColumn('origin_total_price')) {
            $table->addColumn('origin_total_price', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '原始应收金额(单人)。商品已含税时，应收金额(单人)=（原始单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=原始单价+服务费+总税费', 'after' => 'total_price']);
            $table->update();
        }
    }
}
