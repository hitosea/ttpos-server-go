<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddStatusToProductBomTable extends Migrator
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
        $table = $this->table('product_bom');
        if (!$table->hasColumn('status')) {
            $table->addColumn(Column::tinyInteger('status')->setDefault(1)->setComment('状态,0-下架 1-上架')->setAfter('is_default_select'));
        }
        if (!$table->hasColumn('is_sold_out')) {
            $table->addColumn(Column::tinyInteger('is_sold_out')->setDefault(0)->setComment('是否沽清, 0-否 1-是')->setAfter('is_default_select'));
        }
        $table->update();
    }
}
