<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AlterNumToWarehouseOutFormItem extends Migrator
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
        if ($table->hasColumn('num')) {
            $table->changeColumn('num', 'decimal', ['precision' => 20, 'scale' => 4, 'null' => false, 'default' => 0, 'comment' => '数量', 'after' => 'uuid'])
                ->update();
        }
    }
}
