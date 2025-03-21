<?php

use think\facade\Db;
use think\facade\Log;
use think\migration\Migrator;
use think\migration\db\Column;

class MigrateAccessSoldOutBusinessDataAddUser extends Migrator
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
        $db = Db::connect(Db::getConfig('default'), true);
        $db->execute("update ttpos_access set path = 'cashier_sold_out' where path = 'cashier_guqing'");
        $db->execute("update ttpos_access set path = 'cashier_business_data' where path = 'cashier_Business_data'");
        $db->execute("update ttpos_access set path = 'cashier_add_member' where path = 'cashier_add_user'");
    }
}
