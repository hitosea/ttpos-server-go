<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnFinalPriceToSaleOrder extends Migrator
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
        $table = $this->table('sale_order');
        if (!$table->hasColumn('final_price')) {
            $table->addColumn('final_price', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '最终价格。 最终价格=订单金额-订单折扣金额。'])->update();
        }
    }
}
