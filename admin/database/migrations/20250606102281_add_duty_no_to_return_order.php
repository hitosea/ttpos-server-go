<?php

use think\migration\Migrator;

class AddDutyNoToReturnOrder extends Migrator
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
        $table = $this->table('return_order');
        if (!$table->hasColumn('duty_no')) {
            $table->addColumn('duty_no', 'string', ['default' => '', 'comment' => '当班编号', 'after' => 'account_name']);
        }
        $table->update();
    }
}
