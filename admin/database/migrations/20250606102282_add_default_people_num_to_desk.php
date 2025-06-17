<?php

use think\migration\Migrator;

class addDefaultPeopleNumToDesk extends Migrator
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
        if (!$table->hasColumn('is_open_default_people_num')) {
            $table->addColumn('is_open_default_people_num', 'integer', ['default' => 0, 'comment' => '是否开启默认人数', 'after' => 'device_uuid']);
        }
        if (!$table->hasColumn('default_people_num')) {
            $table->addColumn('default_people_num', 'integer', ['default' => 0, 'comment' => '默认人数', 'after' => 'is_open_default_people_num']);
        }
        $table->update();
    }
}
