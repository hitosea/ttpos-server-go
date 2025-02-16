<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateTtposSaleOrderProductQrcodeTable extends Migrator
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
        if ($table->hasColumn('qrcode_order_uuid')) {
            $table->removeColumn('qrcode_order_uuid')
                  ->update();
        }

        $table = $this->table('product_bom');
        if (!$table->hasColumn('status')) {
            $table->addColumn('status', 'tinyinteger', ['null' => false, 'default' => 0, 'comment' => '状态, 0-下架 1-上架. 同步商品包的状态'])
                  ->update();
        }
    }
}