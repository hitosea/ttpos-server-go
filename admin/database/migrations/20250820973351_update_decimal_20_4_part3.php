<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateDecimal204Part3 extends Migrator
{
    /**
     * Change Method.
     *
     * 继续执行剩余的decimal字段类型修改 - 第三部分
     */
    public function change()
    {
        // 继续执行剩余的SQL语句
        $allSqls = [
            "ALTER TABLE `ttpos_purchase_receipt_order` MODIFY COLUMN `num` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '物资数量，每种物品算一个';",
            "ALTER TABLE `ttpos_purchase_receipt_order_item` MODIFY COLUMN `num` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '收货数量';",
            "ALTER TABLE `ttpos_purchase_receipt_order_item` MODIFY COLUMN `unit_conversion_rate` decimal(20,4) NULL DEFAULT 1.0000 COMMENT '单位转换率。收货数量*转换率=基准单位收货数量';",
            "ALTER TABLE `ttpos_refund_order` MODIFY COLUMN `amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款金额';",
            "ALTER TABLE `ttpos_related_material` MODIFY COLUMN `num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '材料用量,可小数';",
            "ALTER TABLE `ttpos_related_material` MODIFY COLUMN `base_unit_conversion_rate` decimal(20,4) NULL DEFAULT 1.0000 COMMENT '基准单位转换率。用量*转换率=基准单位用量';",
            "ALTER TABLE `ttpos_return_order` MODIFY COLUMN `refund_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款金额,包括税额';",
            "ALTER TABLE `ttpos_return_order` MODIFY COLUMN `refund_tax_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款税额';",
            "ALTER TABLE `ttpos_return_order_amount` MODIFY COLUMN `amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款金额';",
            "ALTER TABLE `ttpos_return_order_product` MODIFY COLUMN `product_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品单价';",
            "ALTER TABLE `ttpos_return_order_product` MODIFY COLUMN `tax_rate` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '税率,根据结账时税率计算';",
            "ALTER TABLE `ttpos_return_order_product` MODIFY COLUMN `num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品数量,退货的商品数量';",
            "ALTER TABLE `ttpos_return_order_product` MODIFY COLUMN `product_discount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品折扣';",
            "ALTER TABLE `ttpos_return_order_product` MODIFY COLUMN `product_total_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品总金额（退款总金额）';",
            "ALTER TABLE `ttpos_sale_bill` MODIFY COLUMN `amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '订单金额(折后价),关联销售订单的总金额之和';",
            "ALTER TABLE `ttpos_sale_bill` MODIFY COLUMN `origin_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '订单金额(折前价)。商品未含税时，订单金额(折前价)=商品金额+服务费+税费。商品已含税时，订单金额(折前价)=商品金额（含商品消费税）+服务费+税费（只有服务费税）';",
            "ALTER TABLE `ttpos_sale_bill` MODIFY COLUMN `product_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品金额,关联销售订单的商品金额之和';",
            "ALTER TABLE `ttpos_sale_bill` MODIFY COLUMN `service_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务费,关联销售订单的服务费之和';",
            "ALTER TABLE `ttpos_sale_bill` MODIFY COLUMN `tax_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '税费,关联销售订单的税费之和';",
            "ALTER TABLE `ttpos_sale_bill` MODIFY COLUMN `custom_discount_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '自定义折扣费用,关联销售订单的会员折扣费用之和';",
            "ALTER TABLE `ttpos_sale_bill` MODIFY COLUMN `member_discount_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '会员折扣费用,关联销售订单的会员折扣费用之和';",
            "ALTER TABLE `ttpos_sale_bill` MODIFY COLUMN `gift_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠菜金额,关联销售订单的赠菜金额之和';",
            "ALTER TABLE `ttpos_sale_bill` MODIFY COLUMN `free_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '免单金额,关联销售订单的免单金额之和';",
            "ALTER TABLE `ttpos_sale_bill` MODIFY COLUMN `payment_commission_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付手续费,多次支付的支付手续费之和';",
            "ALTER TABLE `ttpos_sale_bill` MODIFY COLUMN `payment_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付金额,支付金额-订单总金额=支付手续费';",
            "ALTER TABLE `ttpos_sale_bill` MODIFY COLUMN `product_original_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '原始商品金额。 商品原始金额=(订单.原始商品金额)之和。';",
            "ALTER TABLE `ttpos_sale_bill_setting` MODIFY COLUMN `service_fee_value` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例';",
            "ALTER TABLE `ttpos_sale_bill_setting` MODIFY COLUMN `points_exchange_rate` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '积分抵扣汇率,1积分抵扣多少元';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `member_discount_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总会员折扣金额。总会员折扣金额=(订单商品.会员折扣金额)之和';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `custom_discount_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总自定义折扣金额。总自定义折扣金额=(订单商品.自定义折扣金额)之和';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `zero_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '优惠折扣抹零金额。';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `product_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品金额，订单商品的最终单价(折后价)之和。商品已含税时，该金额包括了税费。当商品未含税时，该金额不包括税费';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `product_original_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '原始商品金额(折前价)。 商品原始金额=订单商品的销售价(折前价)之和。';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `service_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务费固定服务费时，服务费=固定服务费；按比例收服务费时，服务费=(订单商品.总服务费)之和';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `tax_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '税费。税费=(订单商品.总税费)之和';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '应收金额。商品未含税时，总金额=商品金额+服务费+税费。商品已含税时，总金额=商品金额（含商品消费税）+服务费+税费（只有服务费税）';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `origin_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '原始应收金额。原始应收金额=商品金额+服务费+消费税。商品未含税时，原始应收金额=商品金额+服务费+消费税（商品消费税税费+服务费税费）。商品已含税时，原始应收金额=商品金额（包含商品消费税税费）+服务费+服务费税费。';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `member_discount_rate` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '会员折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `member_card_discount_rate` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '会员卡折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `custom_discount_rate` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '自定义折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `custom_amount` decimal(20,4) NOT NULL DEFAULT -1.0000 COMMENT '整单改价金额。改价后，应收金额=整单改价金额，前端优先显示改价后的金额，改价金额不能为负数。当为-1时，表示不改价，显示amount改收金额';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `pay_points` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '抵扣积分,用了多少积分进行抵扣';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `pay_points_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '抵扣金额,积分 抵扣了多少金额';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `points_exchange_rate` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '积分抵扣汇率,1积分抵扣多少元';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `coupon_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '优惠券抵扣金额,抵扣了多少金额';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `payment_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '已支付金额,关联付款单的支付金额之和。';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `change_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '找零金额,结账完成后才记录';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `zero_checkout_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '结账抹零金额。';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `final_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '最终应收金额。最终应收金额=应收金额+手续费-结账抹零金额';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `payment_commission_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付手续费,关联付款单的支付手续费之和';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `gift_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠菜金额,(销售订单赠菜商品.总最终单价)之和';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `gift_points` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠送积分. 赠送积分=应收金额amount*积分赠送比例.';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `gift_points_rate` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠送积分比例. 取值范围0-1。结账后记录，不受后台改变';",
            "ALTER TABLE `ttpos_sale_order` MODIFY COLUMN `member_balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '会员余额.会员消费本单后剩余的余额';"
        ];

        // 分批执行SQL语句
        foreach ($allSqls as $index => $sql) {
            try {
                $this->execute($sql);
                echo "执行SQL Part3 " . ($index + 1) . "/" . count($allSqls) . " 成功\n";
            } catch (Exception $e) {
                // 记录错误但继续执行
                echo "Warning SQL Part3 " . ($index + 1) . ": " . $e->getMessage() . "\n";
            }
        }
    }
}
