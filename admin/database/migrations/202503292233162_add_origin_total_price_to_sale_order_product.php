<?php

use think\migration\Migrator;

class AddOriginTotalPriceToSaleOrderProduct extends Migrator
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
        if (!$table->hasColumn('origin_total_price')) {
            $table->addColumn('origin_total_price', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '原始应收金额(单商品)。商品已含税时，应收金额(单商品)=(销售价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=销售价+服务费+总税费', 'after' => 'total_price']);
            $table->update();
        }
        $table = $this->table('sale_order');
        if (!$table->hasColumn('origin_amount')) {
            $table->addColumn('origin_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '原始应收金额。原始应收金额=商品金额+服务费+消费税。商品未含税时，原始应收金额=商品金额+服务费+消费税（商品消费税税费+服务费税费）。商品已含税时，原始应收金额=商品金额（包含商品消费税税费）+服务费+服务费税费。', 'after' => 'amount']);
            $table->update();
        }
    }
}
