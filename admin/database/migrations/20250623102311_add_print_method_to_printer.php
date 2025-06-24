<?php

use think\migration\Migrator;

class AddPrintMethodToPrinter extends Migrator
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
        if (!$table->hasColumn('print_method')) {
            $table->addColumn('print_method', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '打印方式 1文本打印, 2图片打印', 'after' => 'sort']);
            $table->update();
        }
    }
}

