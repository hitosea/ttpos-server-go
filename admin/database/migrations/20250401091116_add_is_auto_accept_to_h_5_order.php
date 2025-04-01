<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddIsAutoAcceptToH5Order extends Migrator
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
        $table = $this->table('h5_order');
        if (!$table->hasColumn('is_auto_accept')) {
            $table->addColumn(Column::Integer('is_auto_accept')->setDefault(0)->setComment('是否自动接单, 0-否 1-是')->setAfter('status'));
        }
        $table->update();
    }
}
