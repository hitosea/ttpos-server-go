<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddOriginAmountSaleBill extends Migrator
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
        $table = $this->table('sale_bill');
        if (!$table->hasColumn('product_original_amount')) {
            $table->addColumn('product_original_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '原始商品金额。 商品原始金额=(订单.原始商品金额)之和。'])->update();
        }
    }
}
