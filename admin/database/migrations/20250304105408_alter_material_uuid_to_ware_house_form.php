<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AlterMaterialUuidToWareHouseForm extends Migrator
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
        $table = $this->table('warehouse_form');
        if ($table->hasColumn('material_uuid')) {
            $table->changeColumn('material_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '材料uuid', 'after' => 'product_bom_uuid']);
        }
        $table->update();
    }
}
