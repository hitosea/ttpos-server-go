<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnNeedServiceFeeDeskTable extends Migrator
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
        $table = $this->table('desk');
        if (!$table->hasColumn('need_service_fee')) {
            $table->addColumn('need_service_fee', 'integer', ['default' => 1, 'comment' => '是否需要服务费, 0-否 1-是.标记该桌台收取服务费'])
                ->update();
        }
    }
}
