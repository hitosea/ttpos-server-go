<?php

use think\facade\Db;
use think\migration\Migrator;

class AddAuthPhoneToMemberAddress extends Migrator
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
        $table = $this->table('member_address');
        if (!$table->hasColumn('auth_phone')) {
            $table->addColumn('auth_phone', 'string', ['limit' => 20, 'null' => false, 'default' => '', 'comment' => '认证手机号', 'after' => 'location']);
            $table->addColumn('auth_time', 'integer', ['limit' => 11, 'null' => false, 'default' => 0, 'comment' => '认证时间', 'after' => 'auth_phone']);
            $table->update();
        }
    }
}
