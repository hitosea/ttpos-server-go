<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddStatusToWarehouseOutFormItem extends Migrator
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
        if (!$table->hasColumn('status')) {
                $table->addColumn('status', 'biginteger', [
                    'null' => false,
                    'default' => 0,
                    'comment' => '状态,0-预出库 1-已出库。预出库时，表示库存扣减但未在出库记录页面显示.已出库时才在出库记录页面显示',
                    'after' => 'uuid',
                ]);
        }
        $table->update();
    }
}
