<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddReduceStockToWarehouseOutFormItem extends Migrator
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
        $table = $this->table('warehouse_out_form_item');
        if (!$table->hasColumn('reduce_stock')) {
            $table->addColumn('reduce_stock', 'tinyinteger', [
                'null' => false,
                'default' => 0,
                'comment' => '是否已经减库存,0-未减库存 1-已减库存。用于判断该出库记录是否已经将对应的货物减库存，若没减库存将在下次检查时减该货物的库存',
                'after' => 'status',
            ]);
        }
        $table->update();
    }
}
