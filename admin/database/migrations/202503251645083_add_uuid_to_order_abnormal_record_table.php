<?php

use think\migration\Migrator;

class AddUuidToOrderAbnormalRecordTable extends Migrator
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
        $table = $this->table('sale_order_abnormal_record');
        if (!$table->hasColumn('uuid')) {
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => 'UUID', 'after' => 'id']);
            $table->update();
        }
    }
}
