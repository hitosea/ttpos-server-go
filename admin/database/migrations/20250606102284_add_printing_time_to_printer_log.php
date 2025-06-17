<?php

use think\migration\Migrator;

class AddPrintingTimeToPrinterLog extends Migrator
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
        $table = $this->table('printer_log');
        if (!$table->hasColumn('printing_time')) {
            $table->addColumn('printing_time', 'integer', ['default' => 0, 'comment' => '打印耗时(毫秒)', 'after' => 'first_execution']);
            $table->update();
        }
    }
}
