<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnSaleBillUuidAndProductPackageUuidToProductOrderProduct extends Migrator
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
        $table = $this->table('production_order_product');
        if (!$table->hasColumn('product_package_uuid')) {
            $table->addColumn('product_package_uuid', 'biginteger', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '商品包ID',
                'after' => 'has_material',
            ]);
        }
        if (!$table->hasColumn('sale_bill_uuid')) {
            $table->addColumn('sale_bill_uuid', 'biginteger', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '销售账单ID',
                'after' => 'has_material',
            ]);
        }
        $table->update();
    }
}
