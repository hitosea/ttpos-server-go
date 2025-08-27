<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateDecimal204Part4 extends Migrator
{
    /**
     * Change Method.
     *
     * 执行最后剩余的decimal字段类型修改 - 第四部分
     */
    public function change()
    {
        // 执行最后剩余的SQL语句
        $allSqls = [
            "ALTER TABLE `ttpos_sale_order_buffet_customer_type` MODIFY COLUMN `sale_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变';",
            "ALTER TABLE `ttpos_sale_order_buffet_customer_type` MODIFY COLUMN `sale_price_no_tax` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '销售价,未含税价格（折前）';",
            "ALTER TABLE `ttpos_sale_order_buffet_customer_type` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '最终单价（折后价），只进行自定义打折，不进行会员打折';",
            "ALTER TABLE `ttpos_sale_order_buffet_customer_type` MODIFY COLUMN `custom_discount_rate` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '自定义折扣率, 值为0-1之间(0-100%)';",
            "ALTER TABLE `ttpos_sale_order_buffet_customer_type` MODIFY COLUMN `custom_discount_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '自定义折扣金额（单人）。自定义折扣金额（单人）=自助餐顾客类型原价*自定义折扣率';",
            "ALTER TABLE `ttpos_sale_order_buffet_customer_type` MODIFY COLUMN `tax_rate` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '税率,值为0-1之间.加购时记录税率,结账时再重新核算';",
            "ALTER TABLE `ttpos_sale_order_buffet_customer_type` MODIFY COLUMN `service_tax_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务费税费（单人）,0-不收取税费；收取时，服务费税费=服务费*税率';",
            "ALTER TABLE `ttpos_sale_order_buffet_customer_type` MODIFY COLUMN `tax_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '自助餐顾客类型税费（单人）。自助餐顾客类型已含税时，税费=自助餐顾客类型原价*(1-1/(1+税率))；自助餐顾客类型未含税时，税费=自助餐顾客类型原价*税率';",
            "ALTER TABLE `ttpos_sale_order_buffet_customer_type` MODIFY COLUMN `service_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务费（单人）,0-固定服务费 大于0-按比例收服务费；自助餐顾客类型已含税时，服务费=(自助餐顾客类型原价-自助餐顾客类型税费)*服务费比例；自助餐顾客类型未含税时，服务费=自助餐顾客类型原价*服务费比例';",
            "ALTER TABLE `ttpos_sale_order_buffet_customer_type` MODIFY COLUMN `total_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '应收金额(单人)。商品已含税时，应收金额(单人)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费';",
            "ALTER TABLE `ttpos_sale_order_buffet_customer_type` MODIFY COLUMN `origin_total_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '原始应收金额(单人)。商品已含税时，应收金额(单人)=（原始单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=原始单价+服务费+总税费';",
            "ALTER TABLE `ttpos_sale_order_buffet_delay_product` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '价格,下单时固定不受后台改变，结账时再检查是否改变';",
            "ALTER TABLE `ttpos_sale_order_coupon` MODIFY COLUMN `coupon_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '优惠券抵扣金额，实际抵扣金额';",
            "ALTER TABLE `ttpos_sale_order_coupon` MODIFY COLUMN `coupon_origin_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '优惠券原始金额(面值)';",
            "ALTER TABLE `ttpos_sale_order_discount_strategy` MODIFY COLUMN `value` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '优惠策略值';",
            "ALTER TABLE `ttpos_sale_order_peak_time` MODIFY COLUMN `amount` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '订单金额';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品数量';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `unit_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '单位数量，用于套餐子商品';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `flavor_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '规格原价（单商品）,仅某规格商品的原价';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `sauce_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '小料价（单商品）,所有小料的价格之和';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `product_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '原始单价（单商品）,规格原价+小料价';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `sale_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '销售价（单商品，折前价）,当自定义价格时，销售价=自定义价格,否则销售价=原始单价';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `sale_price_no_tax` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '销售价,未含税价格（折前）';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `tax_rate` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '税率,单位%.加购时记录税率,结账时再重新核算';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `member_discount_rate` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '会员折扣率(0-100%)';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `member_card_discount_rate` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '会员卡折扣率(0-100%)';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `member_order_discount_rate` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '会员端商品价格上浮比例1%-300%';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `custom_discount_rate` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '自定义折扣率(0-100%)';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `member_discount_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '会员折扣后的价格（单商品）=销售价*会员折扣率*会员卡折扣率';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。销售价*折扣率';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `service_tax_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务费税费（单商品）,0-不收取税费；收取时，服务费税费=服务费*税率';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `tax_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品税费（单商品）。商品已含税时，税费=规格原价*(1-1/(1+税率))；商品未含税时，税费=原始单价*税率';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `service_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务费（单商品）,0-固定服务费 大于0-按比例收服务费；商品已含税时，服务费=(最终单价-商品税费)*服务费比例；商品未含税时，服务费=最终单价*服务费比例';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `total_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '应收金额(单商品)。商品已含税时，应收金额(单商品)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `origin_total_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '应收金额(单商品)。商品已含税时，应收金额(单商品)=(销售价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=销售价+服务费+总税费';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `discount_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '打折金额（单商品）=销售价-最终单价。校验：打折金额=会员折扣金额+自定义折扣金额';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `member_discount_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '会员折扣金额（单商品）=销售价*（1-会员折扣率*会员卡折扣率）';",
            "ALTER TABLE `ttpos_sale_order_product` MODIFY COLUMN `custom_discount_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '自定义折扣金额（单商品）。自定义折扣金额（单商品）=会员折扣后的价格（单商品）*(1-自定义折扣率) 。校验：自定义折扣金额（单商品）=销售价 - 最终单价（单商品）-会员折扣金额（单商品）；注意，不能这样算，自定义折扣金额（单商品）=销售价*(1-自定义折扣率)';",
            "ALTER TABLE `ttpos_sale_order_product_bom` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '单价,不随后台更新，记录加购时的价格。结账时要校验价格是否变动';",
            "ALTER TABLE `ttpos_staff_shift_log` MODIFY COLUMN `previous_shift_cash` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '上一班遗留备用金';",
            "ALTER TABLE `ttpos_staff_shift_log` MODIFY COLUMN `current_cash_total` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '当前钱箱现金总计';",
            "ALTER TABLE `ttpos_staff_shift_log` MODIFY COLUMN `total_income` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总收入';",
            "ALTER TABLE `ttpos_staff_shift_log` MODIFY COLUMN `cash_taken_out` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '本班取出现金';",
            "ALTER TABLE `ttpos_staff_shift_log` MODIFY COLUMN `cash_left` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '本班遗留备用金';",
            "ALTER TABLE `ttpos_staff_shift_log` MODIFY COLUMN `cash_income` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '本班收入现金';",
            "ALTER TABLE `ttpos_staff_shift_log` MODIFY COLUMN `total_business` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '本班营业总额(不包含退款)';",
            "ALTER TABLE `ttpos_staff_shift_log` MODIFY COLUMN `withdraw_cash` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '中途取出现金';",
            "ALTER TABLE `ttpos_staff_shift_log` MODIFY COLUMN `deposit_cash` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '中途存入现金';",
            "ALTER TABLE `ttpos_statistics_customer_type` MODIFY COLUMN `product_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '销售价,未含税价格（折前）';",
            "ALTER TABLE `ttpos_statistics_customer_type` MODIFY COLUMN `product_sale_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变';",
            "ALTER TABLE `ttpos_statistics_customer_type` MODIFY COLUMN `product_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品数量';",
            "ALTER TABLE `ttpos_statistics_customer_type` MODIFY COLUMN `tax_rate` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '税率';",
            "ALTER TABLE `ttpos_statistics_customer_type` MODIFY COLUMN `tax_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '税费';",
            "ALTER TABLE `ttpos_statistics_customer_type` MODIFY COLUMN `service_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务费';",
            "ALTER TABLE `ttpos_statistics_customer_type` MODIFY COLUMN `service_tax` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务税';",
            "ALTER TABLE `ttpos_statistics_customer_type` MODIFY COLUMN `give_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠菜数量';",
            "ALTER TABLE `ttpos_statistics_customer_type` MODIFY COLUMN `free_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '免单数量';",
            "ALTER TABLE `ttpos_statistics_customer_type` MODIFY COLUMN `refund_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款数量';",
            "ALTER TABLE `ttpos_statistics_delay` MODIFY COLUMN `product_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '销售价,未含税价格（折前）';",
            "ALTER TABLE `ttpos_statistics_delay` MODIFY COLUMN `product_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品数量';",
            "ALTER TABLE `ttpos_statistics_delay` MODIFY COLUMN `tax_rate` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '税率';",
            "ALTER TABLE `ttpos_statistics_delay` MODIFY COLUMN `tax_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '税费';",
            "ALTER TABLE `ttpos_statistics_delay` MODIFY COLUMN `service_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务费';",
            "ALTER TABLE `ttpos_statistics_delay` MODIFY COLUMN `service_tax` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务税';",
            "ALTER TABLE `ttpos_statistics_delay` MODIFY COLUMN `give_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠菜数量';",
            "ALTER TABLE `ttpos_statistics_delay` MODIFY COLUMN `free_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '免单数量';",
            "ALTER TABLE `ttpos_statistics_delay` MODIFY COLUMN `refund_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款数量';",
            "ALTER TABLE `ttpos_statistics_member` MODIFY COLUMN `recharge_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '充值金额';",
            "ALTER TABLE `ttpos_statistics_member` MODIFY COLUMN `give_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠送金额';",
            "ALTER TABLE `ttpos_statistics_member` MODIFY COLUMN `give_point` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠送积分';",
            "ALTER TABLE `ttpos_statistics_member` MODIFY COLUMN `payment_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付金额';",
            "ALTER TABLE `ttpos_statistics_member` MODIFY COLUMN `payment_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付手续费';",
            "ALTER TABLE `ttpos_statistics_member` MODIFY COLUMN `refund_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款金额';",
            "ALTER TABLE `ttpos_statistics_member` MODIFY COLUMN `refund_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款手续费';",
            "ALTER TABLE `ttpos_statistics_member_payment` MODIFY COLUMN `payment_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付金额';",
            "ALTER TABLE `ttpos_statistics_member_payment` MODIFY COLUMN `refund_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款金额';",
            "ALTER TABLE `ttpos_statistics_payment` MODIFY COLUMN `payment_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付金额';",
            "ALTER TABLE `ttpos_statistics_payment` MODIFY COLUMN `refund_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款金额';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `product_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品单价: 未含税';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `product_sale_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品销售价: 规格+加料';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `product_final_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品最终价';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `flavor_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品原价(规格价)';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `sauce_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '加料价格';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `product_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品数量';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `tax_rate` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '税率';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `tax_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '税费';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `service_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务费';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `service_tax` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务税';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `give_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠菜数量';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `free_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '免单数量';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `refund_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款数量';",
            "ALTER TABLE `ttpos_statistics_product` MODIFY COLUMN `member_order_discount_rate` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '会员端商品价格上浮比例1%-300%';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `product_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品原价: 不含税';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `product_origin_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '原商品金额';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `product_sale_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品销售价';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `product_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品数量';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `product_tax` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品税';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `service_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务费';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `service_tax` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '服务税';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `discount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '优惠折扣';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `discount_member` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '会员折扣';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `gift_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠菜金额';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `gift_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠菜数量';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `free_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '免单金额';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `free_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '免单数量';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `payment_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付金额';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `payment_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付手续费';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `payment_balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付余额';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `refund_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款金额';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `refund_payment_balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款支付余额';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `refund_tax` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款税额';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `no_refund_tax` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '不退税金额';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `extend_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '扩展价格';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `delivery_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '配送费';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `refund_service_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款服务费';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `refund_discount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款优惠折扣';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `refund_discount_member` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款会员折扣';",
            "ALTER TABLE `ttpos_statistics_sale` MODIFY COLUMN `refund_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '退款支付手续费';",
            "ALTER TABLE `ttpos_tax` MODIFY COLUMN `tax_rate` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '税率';",
            "ALTER TABLE `ttpos_warehouse_form` MODIFY COLUMN `num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '数量';",
            "ALTER TABLE `ttpos_warehouse_form_item` MODIFY COLUMN `num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '入库数量';",
            "ALTER TABLE `ttpos_warehouse_monthly_form` MODIFY COLUMN `stock` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '库存';",
            "ALTER TABLE `ttpos_warehouse_monthly_material_form` MODIFY COLUMN `stock` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '库存';",
            "ALTER TABLE `ttpos_warehouse_monthly_product_bom_form` MODIFY COLUMN `stock` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '库存';",
            "ALTER TABLE `ttpos_warehouse_out_form_item` MODIFY COLUMN `num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '数量';"
        ];

        // 分批执行SQL语句
        foreach ($allSqls as $index => $sql) {
            try {
                $this->execute($sql);
                echo "执行SQL Part4 " . ($index + 1) . "/" . count($allSqls) . " 成功\n";
            } catch (Exception $e) {
                // 记录错误但继续执行
                echo "Warning SQL Part4 " . ($index + 1) . ": " . $e->getMessage() . "\n";
            }
        }
    }
}
