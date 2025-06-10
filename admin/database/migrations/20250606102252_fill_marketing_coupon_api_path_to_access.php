<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class FillMarketingCouponApiPathToAccess extends Migrator
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
        $accesses = $db->name("access")->where('uuid', 'in', [
            1731155610,1731155611,1731155612,1731155620
        ])->select();

        foreach ($accesses as   $access) {
            switch ($access['uuid']) {
                case 1731155610:
                    $db->name("access")->where("id", "=", $access['id'])->update(['api_path' => '/marketing.coupon/list']);
                    break;
                case 1731155611:
                    $db->name("access")->where("id", "=", $access['id'])->update(['api_path' => '/marketing.coupon/add']);
                    break;
                case 1731155612:
                    $db->name("access")->where("id", "=", $access['id'])->update(['api_path' => '/marketing.coupon/edit']);
                    break;
                case 1731155620:
                    $db->name("access")->where("id", "=", $access['id'])->update(['api_path' => '/marketing.coupon/record']);
                    break;
            }
        }
    }
}
