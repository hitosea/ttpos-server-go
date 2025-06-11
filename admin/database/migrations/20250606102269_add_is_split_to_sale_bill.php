<?php

use think\migration\Migrator;

class AddIsSplitToSaleBill extends Migrator
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
        $table = $this->table('sale_bill');
        if (!$table->hasColumn('is_split_order')) {
            $table->addColumn('is_split_order', 'integer', ['default' => 0, 'comment' => '是否拆单, 0-否 1-是', 'after' => 'is_lock']);
            $table->update();
        }
    }
}
