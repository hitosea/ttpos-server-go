<?php

use think\facade\Db;
use think\migration\Migrator;

class InsertLianlianpayToPaymentMethodTable extends Migrator
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
    public function up()
    {
        $db = Db::connect(Db::getConfig('default'), true);
        $db->name('migrations')->where('migration_name', 'InsertLianlianpayToPaymentMethodTable')->delete();
        // 删除支付方式
        $db->name('payment_method')->where('code', '=', 90111)->delete();
        $db->name('payment_method')->where('code', '=', 90222)->delete();
        $db->name('payment_method')->where('code', '=', 90333)->delete();
        // LianLian渠道支付方式
        $typeData = [
            [
                'uuid' => createUuid(),
                'name' => 'WeChat Pay',
                'payment_name' => 'WeChat Pay',
                'source' => 2,
                'code' => 90111,
                'sort' => 0,
                'status' => 1,
                'create_time' => time(),
                'update_time' => time(),
                'is_show_cashier' => 1,
                'is_show_assistant' => 1,
                'is_show_member_recharge' => 0,
            ],
            [
                'uuid' => createUuid(),
                'name' => 'Alipay',
                'payment_name' => 'Alipay',
                'source' => 2,
                'code' => 90222,
                'sort' => 0,
                'status' => 1,
                'create_time' => time(),
                'update_time' => time(),
                'is_show_cashier' => 1,
                'is_show_assistant' => 1,
                'is_show_member_recharge' => 0,
            ],
            [
                'uuid' => createUuid(),
                'name' => 'QR PromptPay',
                'payment_name' => 'QR PromptPay',
                'source' => 2,
                'code' => 90333,
                'sort' => 0,
                'status' => 1,
                'create_time' => time(),
                'update_time' => time(),
                'is_show_cashier' => 1,
                'is_show_assistant' => 1,
                'is_show_member_recharge' => 0,
            ],
        ];
        $db->name('payment_method')->insertAll($typeData);
    }

    /**
     * Migrate Down.
     */
    public function down()
    {
        $db = Db::connect(Db::getConfig('default'), true);
        $db->name('payment_method')->where('code', '=', 90111)->delete();
        $db->name('payment_method')->where('code', '=', 90222)->delete();
        $db->name('payment_method')->where('code', '=', 90333)->delete();
    }
}
