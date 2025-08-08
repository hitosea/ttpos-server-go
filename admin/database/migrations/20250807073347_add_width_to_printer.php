<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddWidthToPrinter extends Migrator
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
        $table = $this->table('printer');
        if (!$table->hasColumn('width')) {
            $table->addColumn('width', 'integer', ['null' => false, 'default' => 80, 'comment' => '纸张宽度（mm）', 'after' => 'print_method']);
        }
        $table->update();
    }
}
