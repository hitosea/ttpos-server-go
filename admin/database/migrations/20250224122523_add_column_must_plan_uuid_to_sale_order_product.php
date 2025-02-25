<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnMustPlanUuidToSaleOrderProduct extends Migrator
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
        if (!$table->hasColumn('must_plan_uuid')) {
            $table->addColumn(Column::bigInteger('must_plan_uuid')->setDefault(0)->setComment('必点方案UUID，用于关联必点方案，用于判断是否为必点方案商品')->setAfter('sale_order_uuid'));
            $table->update();
        }
    }
}
