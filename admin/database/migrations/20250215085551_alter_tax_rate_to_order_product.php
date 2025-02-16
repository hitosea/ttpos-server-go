<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AlterTaxRateToOrderProduct extends Migrator
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
        $table->changeColumn('tax_rate', 'decimal', [
                'precision' => 10, 
                'scale' => 2, 
                'null' => false, 
                'default' => 0, 
                'after' => 'sale_price', 
                'comment' => '税率,单位%.加购时记录税率,结账时再重新核算'
            ])
            ->update();
    }
}
