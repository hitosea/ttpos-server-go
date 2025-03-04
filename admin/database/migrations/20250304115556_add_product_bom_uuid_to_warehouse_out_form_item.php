<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddProductBomUuidToWarehouseOutFormItem extends Migrator
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
        if (!$table->hasColumn('product_bom_uuid')) {
            $table->addColumn('product_bom_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '产品bom uuid', 'after' => 'warehouse_out_form_uuid']);
        }
        $table->update();
    }
}
