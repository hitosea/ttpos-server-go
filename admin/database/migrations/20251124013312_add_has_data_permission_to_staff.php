<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddHasDataPermissionToStaff extends Migrator
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
        $table = $this->table("staff");
        if (!$table->hasColumn('has_data_permission')) {
            $table->addColumn('has_data_permission', 'smallinteger', [
                'limit' => 3,
                'default' => 0,
                'null' => false,
                'comment' => '是否有数据管理权限0否1是',
                'after' => 'is_super',
            ]);
        }
        $table->update();
    }
}
