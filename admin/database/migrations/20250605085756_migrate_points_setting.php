<?php

use think\facade\Cache;
use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class MigratePointsSetting extends Migrator
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
        $existData = $db->name("setting")->where("key", "=", "points")->find();
        if (!$existData) {
            return;
        }
        $companySetting = $db->name("company_setting")->where("id", ">", 0)->find();
        $values = json_decode($existData["values"], true);
        if ($values["is_shopping_gift"] == "1") {
            foreach ($values["shopping_gift_rules"] as $k => $rule) {
                $rule["meal_type"] = [];
                if ($rule["type"] == "payment_amount") {
                    // 按付款金额比例赠送
                    $rule["is_open"] = "1"; // 开启
                    $rule["is_member_level_related"] = "0"; // 不按会员等级赠送
                    $rule["value"] = $values["gift_ratio"]; // 积分比例
                    $rule["payment_amount_requirement"] = "0"; // 无付款金额要求
                    // 适用就餐类型
                    $rule["meal_type"][] = "non-buffet";
                    // 会员余额支付赠送积分
                    $rule["balance_payment_get_points"] = "1";
                    $rule["refund_return_points"] = "1"; // 退款自动退积分
                }
                if ($companySetting["is_open_buffet"] == 1) {
                    $rule["meal_type"][] = "buffet";
                }
                $values["shopping_gift_rules"][$k] = $rule;
            }
            $db->name("setting")->where("key", "=", "points")->update(["values" => json_encode($values)]);
            $key = sprintf("setting:company_id:%d", $companySetting["company_uuid"]);
            if (Cache::has($key)) {
                Cache::delete($key);
            }
        }
    }
}
