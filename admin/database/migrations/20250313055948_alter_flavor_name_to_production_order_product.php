<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AlterFlavorNameToProductionOrderProduct extends Migrator
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
        if ($table->hasColumn('flavor_name')) {
            $table->changeColumn('flavor_name', 'text', [
                'null' => true,
                'default' => null,
                'comment' => '规格名称,不随后台改变',
                'after' => 'num',
            ]);
        }
        $table->update();
    }
}
