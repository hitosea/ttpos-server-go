<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateDisableToDesk extends Migrator
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
        if ($table->hasColumn('is_disable')) {
            $table->changeColumn('is_disable', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '是否禁用, 0-否 1-是',
            ]);
        }
        $table->update();
    }
}
