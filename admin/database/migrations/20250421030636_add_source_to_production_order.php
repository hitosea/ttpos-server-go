<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSourceToProductionOrder extends Migrator
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
        $table = $this->table('production_order');
        if (!$table->hasColumn('source')) {
            $table->addColumn('source', 'string', ['default' => '', 'comment' => '操作来源 shop-商家、cashier-收银机、tablet-平板端、kitchen-厨显端、assistant-点餐助手、h5-H5', 'after' => 'sale_bill_uuid']);
            $table->update();
        }
    }
}
