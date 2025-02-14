<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddStockNumToProductBomTable extends Migrator
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
        $table = $this->table('product_bom');
        if (!$table->hasColumn('stock_num')) {
            $table->addColumn(Column::decimal('stock_num', 12, 4)->setDefault(0)->setComment('库存数量')->setAfter('product_package_uuid'));
        }
        if (!$table->hasColumn('barcode_value')) {
            $table->addColumn(Column::string('barcode_value')->setDefault('')->setComment('条形码值')->setAfter('product_package_uuid'));
        }
        $table->update();
    }
}
