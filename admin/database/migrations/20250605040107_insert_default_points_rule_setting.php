<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class InsertDefaultPointsRuleSetting extends Migrator
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
        $query = $db->name("setting");
        $existData = $query->where("key", "=", "points")->find();
        if (!$existData) {
            return;
        }
        $values = json_decode($existData["values"], true);
        $values["shopping_gift_rules"] = [
            [
                'type' => 'payment_amount',
                'is_open' => '0',
                'is_member_level_related' => '0',
                'value' => '',
                'payment_amount_requirement' => '',
                'meal_type' => [
                ],
                'balance_payment_get_points' => '0',
                'refund_return_points' => '0',
                'member_levels' => [
                ],
            ],
            [
                'type' => 'desk',
                'is_open' => '0',
                'is_member_level_related' => '0',
                'value' => '',
                'payment_amount_requirement' => '',
                'meal_type' => [
                ],
                'balance_payment_get_points' => '0',
                'refund_return_points' => '0',
                'member_levels' => [
                ],
            ],
        ];
        $values["exchange"] = [
            'open_points_exchange' => '0',
            'points_exchange_rate' => '',
            'auto_points_exchange' => '0',
        ];
        $existData["values"] = json_encode($values);
        $query->update($existData);
    }
}
