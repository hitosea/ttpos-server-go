<?php

use think\migration\Migrator;
use think\migration\db\Column;

class RemoveColumnSyncStatus extends Migrator
{
    const TARGET = 'all';
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
        // TODO 更新后删除
        $table = $this->table('company');
        if ($table->hasColumn('sync_status')) {
            $table->removeColumn('sync_status')->update();
        }
    }
}
