<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class UpdateAccessApiPath extends Migrator
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
        $db->name('access')->where('api_path', '=', '/marketing.activity/list')->update(['api_path' => '/marketing/activity/list']);
        $db->name('access')->where('api_path', '=', '/marketing.activity/add')->update(['api_path' => '/marketing/activity/add']);
        $db->name('access')->where('api_path', '=', '/marketing.activity/edit')->update(['api_path' => '/marketing/activity/edit']);
        $db->name('access')->where('api_path', '=', '/marketing.activity/disable')->update(['api_path' => '/marketing/activity/disable']);
        $db->name('access')->where('api_path', '=', '/marketing.activity/record')->update(['api_path' => '/marketing/activity/record']);
        $db->name('access')->where('api_path', '=', '/marketing.coupon/list')->update(['api_path' => '/marketing/coupon/list']);
        $db->name('access')->where('api_path', '=', '/marketing.coupon/add')->update(['api_path' => '/marketing/coupon/add']);
        $db->name('access')->where('api_path', '=', '/marketing.coupon/edit')->update(['api_path' => '/marketing/coupon/edit']);
        $db->name('access')->where('api_path', '=', '/marketing.coupon/record')->update(['api_path' => '/marketing/coupon/record']);
    }
}
