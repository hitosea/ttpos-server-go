<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateDecimal204Part2 extends Migrator
{
    /**
     * Change Method.
     *
     * 继续执行剩余的decimal字段类型修改
     */
    public function change()
    {
        // 继续执行剩余的SQL语句
        $allSqls = [
            "ALTER TABLE `ttpos_member_card` MODIFY COLUMN `discount` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '折扣,单位%, 如80%为打8折，discount值为0.8 .不随后台改变,按领取时的折扣。后续会员卡类型折扣改变时,不改变此字段';",
            "ALTER TABLE `ttpos_member_card_log` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '价格,会员卡价格,不随后台改变,记录领取时的价格';",
            "ALTER TABLE `ttpos_member_card_log` MODIFY COLUMN `discount` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '折扣,单位%,不随后台改变,记录领取时的折扣';",
            "ALTER TABLE `ttpos_member_card_log` MODIFY COLUMN `give_money` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠送余额';",
            "ALTER TABLE `ttpos_member_card_log` MODIFY COLUMN `give_point` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠送积分';",
            "ALTER TABLE `ttpos_member_card_type` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '价格';",
            "ALTER TABLE `ttpos_member_card_type` MODIFY COLUMN `discount` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '折扣,单位%';",
            "ALTER TABLE `ttpos_member_card_type` MODIFY COLUMN `open_point_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '开卡赠送积分数';",
            "ALTER TABLE `ttpos_member_card_type` MODIFY COLUMN `open_money_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '开卡赠送余额数';",
            "ALTER TABLE `ttpos_member_coupon` MODIFY COLUMN `amount` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '优惠券面值';",
            "ALTER TABLE `ttpos_member_coupon_use_record` MODIFY COLUMN `use_order_amount` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '优惠券使用订单金额';",
            "ALTER TABLE `ttpos_member_level` MODIFY COLUMN `upgrade_money` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '升级条件，累计消费额';",
            "ALTER TABLE `ttpos_member_level` MODIFY COLUMN `upgrade_point` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '升级条件，累计积分';",
            "ALTER TABLE `ttpos_member_level` MODIFY COLUMN `discount` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '等级权益,百分比折扣,单位%, 如80%为打8折，discount值为0.8 ';",
            "ALTER TABLE `ttpos_member_point_log` MODIFY COLUMN `value` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '数值,负数:减积分 正数:加积分';",
            "ALTER TABLE `ttpos_member_recharge_order` MODIFY COLUMN `amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '交易金额=充值金额+手续费';",
            "ALTER TABLE `ttpos_member_recharge_order` MODIFY COLUMN `refund_money` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款金额，不大于amount';",
            "ALTER TABLE `ttpos_member_recharge_order` MODIFY COLUMN `charge_due` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '找零';",
            "ALTER TABLE `ttpos_member_recharge_order` MODIFY COLUMN `recharge_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '充值金额';",
            "ALTER TABLE `ttpos_member_recharge_order` MODIFY COLUMN `refund_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款充值金额，不大于recharge_amount';",
            "ALTER TABLE `ttpos_member_recharge_order` MODIFY COLUMN `gift_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠送金额';",
            "ALTER TABLE `ttpos_member_recharge_order` MODIFY COLUMN `gift_point` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠送积分';",
            "ALTER TABLE `ttpos_member_recharge_order` MODIFY COLUMN `balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '充值前会员余额';",
            "ALTER TABLE `ttpos_member_recharge_order` MODIFY COLUMN `balance_recharged` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '充值后会员余额';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `product_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品数量.订单中商品的总数量，商品A数量2，商品B数量1，则商品数量为3';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `product_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品金额';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `origin_product_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品原价,折前价，已含税';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `member_discount_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '会员折扣';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '订单金额';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `refund_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款金额';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `delivery_distance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '配送距离，单位km';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `delivery_fee_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '配送费';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `delivery_fee_distance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '距离费送费';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `delivery_fee_min_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '起步配送费';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `delivery_fee_base_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '基础配送费';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `delivery_fee_per_km` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '每公里配送费';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `rider_rating` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '骑手评分';",
            "ALTER TABLE `ttpos_member_sale_order` MODIFY COLUMN `remaining_distance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '剩余距离';",
            "ALTER TABLE `ttpos_payment_method` MODIFY COLUMN `fee_percent` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '手续费百分比,取值范围0-1';",
            "ALTER TABLE `ttpos_payment_order` MODIFY COLUMN `payment_fee_percent` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付手续费百分比,取值范围0-1';",
            "ALTER TABLE `ttpos_payment_order` MODIFY COLUMN `payment_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付金额';",
            "ALTER TABLE `ttpos_payment_order` MODIFY COLUMN `payment_commission_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付手续费,支付金额*支付手续费百分比';",
            "ALTER TABLE `ttpos_payment_order` MODIFY COLUMN `amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '实收金额，实收金额=支付金额+支付手续费';",
            "ALTER TABLE `ttpos_payment_order` MODIFY COLUMN `balance_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '主账户金额,用于反结账时退款';",
            "ALTER TABLE `ttpos_payment_order` MODIFY COLUMN `gift_balance_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠送帐户金额,用于反结账时退款';",
            "ALTER TABLE `ttpos_production_order_product` MODIFY COLUMN `num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品数量';",
            "ALTER TABLE `ttpos_production_order_product` MODIFY COLUMN `init_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '初始送厨数量，退菜后，init_num肯定大于num';",
            "ALTER TABLE `ttpos_product_bom` MODIFY COLUMN `purchase_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '采购单价';",
            "ALTER TABLE `ttpos_product_bom` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '价格';",
            "ALTER TABLE `ttpos_product_bom` MODIFY COLUMN `stock_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '库存数量';",
            "ALTER TABLE `ttpos_product_bom` MODIFY COLUMN `actual_sale_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '实际销量。每次卖出时,实际销量增加';",
            "ALTER TABLE `ttpos_product_bom_card` MODIFY COLUMN `num` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '加工份数';",
            "ALTER TABLE `ttpos_product_package` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '套餐价格';",
            "ALTER TABLE `ttpos_product_package` MODIFY COLUMN `actual_sale_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '实际销量。每次卖出时,实际销量增加';",
            "ALTER TABLE `ttpos_product_package_group_item` MODIFY COLUMN `num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '数量';",
            "ALTER TABLE `ttpos_product_sauce` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '价格';",
            "ALTER TABLE `ttpos_purchase_form` MODIFY COLUMN `num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总数量';",
            "ALTER TABLE `ttpos_purchase_form` MODIFY COLUMN `amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总金额';",
            "ALTER TABLE `ttpos_purchase_form_item` MODIFY COLUMN `estimate_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '预计单价';",
            "ALTER TABLE `ttpos_purchase_form_item` MODIFY COLUMN `estimate_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '预计金额';",
            "ALTER TABLE `ttpos_purchase_form_item` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '单价';",
            "ALTER TABLE `ttpos_purchase_form_item` MODIFY COLUMN `amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '金额';"
        ];

        // 分批执行SQL语句
        foreach ($allSqls as $index => $sql) {
            try {
                $this->execute($sql);
                echo "执行SQL Part2 " . ($index + 1) . "/" . count($allSqls) . " 成功\n";
            } catch (Exception $e) {
                // 记录错误但继续执行
                echo "Warning SQL Part2 " . ($index + 1) . ": " . $e->getMessage() . "\n";
            }
        }
    }
}
