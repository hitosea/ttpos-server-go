<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class InsertDefaultPaymentMethod extends Migrator
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

        // 查询支付配置
        $payment = $db->name('setting')->where('key', '=', 'payment')->find();
        $payment_values = [];
        if ($payment) {
            $payment_values = json_decode($payment['values'] ?? [], true);
        }
        // 默认支付方式: Cash-现金 Balance Payment-余额
        $paymentMethodList = [
            [
                'uuid' => createUuid(),
                'name' => 'Cash',
                'code' => 40,
                'payment_name' => 'Cash',
                'source' => 0,
                'is_show_cashier' => 1,
                'is_show_assistant' => 1,
                'is_show_member_recharge' => 1,
                'status' => intval($payment_values['is_cash'] ?? 1), // 默认开启
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'uuid' => createUuid(),
                'name' => 'Balance Payment',
                'code' => 10,
                'payment_name' => 'Balance Payment',
                'source' => 0,
                'is_show_cashier' => 1,
                'is_show_assistant' => 1,
                'is_show_member_recharge' => 0,
                'status' => intval($payment_values['is_balance'] ?? 1), // 默认开启
                'create_time' => time(),
                'update_time' => time(),
            ]
        ];
        // 插入数据
        foreach ($paymentMethodList as $paymentMethodItem) {
            $paymentMethod = $db->name('payment_method')->where('name', '=', $paymentMethodItem['name'])->find();
            if ($paymentMethod) {
                $db->name('payment_method')->insert($paymentMethodItem);
            }
        }
    }
}
