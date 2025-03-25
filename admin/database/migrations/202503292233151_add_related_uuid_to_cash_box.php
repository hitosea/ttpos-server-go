<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddRelatedUuidToCashBox extends Migrator
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
        $table = $this->table('cash_box_log');
        if (!$table->hasColumn('related_uuid')) {
            $table->addColumn('related_uuid', 'biginteger', [
                'null' => false,
                'default' => 0,
                'comment' => '关联订单ID',
                'after' => 'processed',
            ]);
            $table->addIndex(['related_uuid'], ['name' => 'cash_box_log_related_uuid']);
        }
        $table->update();
    }
}
