<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddFlavorPriceSaucePriceToStatisticsProduct extends Migrator
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
        $table = $this->table('statistics_product');
        if (!$table->hasColumn('flavor_price')) {
            $table->addColumn('flavor_price', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '商品原价(规格价)', 'after' => 'product_final_price']);
        }
        if (!$table->hasColumn('sauce_price')) {
            $table->addColumn('sauce_price', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '加料价格', 'after' => 'flavor_price']);
        }
        $table->update();
    }
}
