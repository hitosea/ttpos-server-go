<?php

use think\migration\Migrator;

class AddCopiesToProductPrinter extends Migrator
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
        $table = $this->table('product_printer');
        if (!$table->hasColumn('copies')) {
            $table->addColumn('copies', 'integer', [
                'default' => 1,
                'null' => false,
                'comment' => '打印份数',
                'after' => 'print_mode_scene'
            ]);
            $table->update();
        }
    }
} 