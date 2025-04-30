<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddValueToPaymentMethod extends Migrator
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
        $cash = $db->name('payment_method')->where('code', 40)->find();
        if ($cash && $cash['default_img'] == '') {
            $db->name('payment_method')->where('code', 40)->update(['default_img' => '/image/pay/ja_pay.png']);
        }
        $balance = $db->name('payment_method')->where('code', 10)->find();
        if ($balance && $balance['default_img'] == '') {
            $db->name('payment_method')->where('code', 10)->update(['default_img' => '/image/pay/ja_pay.png']);
        }
    }
}
