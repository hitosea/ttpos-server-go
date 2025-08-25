<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateDecimal204 extends Migrator
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
        // 执行所有SQL语句来修改decimal字段类型为decimal(20,4)，默认值改为0.0000
        $allSqls = [
            "ALTER TABLE `ttpos_buffet_customer_type_price` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '价格';",
            "ALTER TABLE `ttpos_buffet_delay` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '价格';",
            "ALTER TABLE `ttpos_buffet_package` MODIFY COLUMN `actual_sale_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '实际销量。每次卖出时,实际销量增加';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_sales` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总销售额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_service_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总服务费';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_payment_commission_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总支付手续费';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_tax_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总税费';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_discount_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总优惠折扣';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_refund_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总退款';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_revenue` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总营业收入';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_actual_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '总实收金额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_recharge_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '充值金额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_gift_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠送金额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `previous_balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '上一班遗留备用金';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_off_cash_withdrawal` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '下班取出现金';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_cash_balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '本班遗留备用金';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `cash_deposit` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '中途存入现金';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `cash_withdrawal` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '中途取出现金';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_min_order_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '最小订单金额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_max_order_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '最大订单金额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_average_order_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '平均订单金额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_table_min_order_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '桌台最小订单金额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_table_max_order_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '桌台最大订单金额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_table_average_order_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '桌台人均消费金额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_scan_min_order_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '点餐最小订单金额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_scan_max_order_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '点餐最大订单金额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_scan_average_order_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '点餐平均订单金额';",
            "ALTER TABLE `ttpos_cashier_duty_detail` MODIFY COLUMN `total_gift_product_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠菜金额';",
            "ALTER TABLE `ttpos_cash_box` MODIFY COLUMN `balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '钱箱余额';",
            "ALTER TABLE `ttpos_cash_box` MODIFY COLUMN `frozen_balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '冻结金额。冻结金额不能使用，在前端显示为已扣除或已增加。冻结金额可为负数。钱箱余额=钱箱余额+冻结金额';",
            "ALTER TABLE `ttpos_cash_box` MODIFY COLUMN `previous_balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '上一班遗留备用金';",
            "ALTER TABLE `ttpos_cash_box` MODIFY COLUMN `cash_withdrawal` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '中途取出金额';",
            "ALTER TABLE `ttpos_cash_box` MODIFY COLUMN `cash_deposit` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '中途存入金额';",
            "ALTER TABLE `ttpos_cash_box_log` MODIFY COLUMN `amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '金额';",
            "ALTER TABLE `ttpos_h5_order` MODIFY COLUMN `member_discount_rate` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '会员折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变';",
            "ALTER TABLE `ttpos_h5_order` MODIFY COLUMN `member_card_discount_rate` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '会员卡折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变';",
            "ALTER TABLE `ttpos_h5_order` MODIFY COLUMN `custom_discount_rate` decimal(20,4) NOT NULL DEFAULT 1.0000 COMMENT '自定义折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变';",
            "ALTER TABLE `ttpos_h5_order` MODIFY COLUMN `product_total_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '商品总价。接单和拒单后从sale_order_product表获取，不再改变';",
            "ALTER TABLE `ttpos_h5_order` MODIFY COLUMN `total_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '订单金额. 订单金额=商品总价*折扣率。接单和拒单后从sale_order_product表获取，不再改变';",
            "ALTER TABLE `ttpos_h5_order_product` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '最终单价（折后价）。接单和拒单后从sale_order_product表获取，不再改变';",
            "ALTER TABLE `ttpos_h5_order_product` MODIFY COLUMN `sale_price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '销售价（折前价）。接单和拒单后从sale_order_product表获取，不再改变';",
            "ALTER TABLE `ttpos_h5_order_product` MODIFY COLUMN `num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '最终商品数量.接单和拒单后从sale_order_product表获取，不再改变';",
            "ALTER TABLE `ttpos_ll_payment_order` MODIFY COLUMN `order_amount` decimal(20,4) NULL DEFAULT 0.0000 COMMENT 'lianlian订单金额';",
            "ALTER TABLE `ttpos_ll_payment_order` MODIFY COLUMN `commission_fee` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '支付手续费,支付金额*支付手续费百分比';",
            "ALTER TABLE `ttpos_loss_report_form` MODIFY COLUMN `num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '数量';",
            "ALTER TABLE `ttpos_marketing_activity` MODIFY COLUMN `reward_value` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '奖励值';",
            "ALTER TABLE `ttpos_marketing_activity` MODIFY COLUMN `reward_condition_amount` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '奖励条件金额';",
            "ALTER TABLE `ttpos_marketing_activity_consumption` MODIFY COLUMN `consumption_amount` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '消费金额';",
            "ALTER TABLE `ttpos_marketing_activity_consumption` MODIFY COLUMN `reward_amount` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '奖励金额';",
            "ALTER TABLE `ttpos_marketing_activity_record` MODIFY COLUMN `reward_value` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '奖励值';",
            "ALTER TABLE `ttpos_marketing_coupon` MODIFY COLUMN `amount` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '优惠券金额';",
            "ALTER TABLE `ttpos_material` MODIFY COLUMN `valuation` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '估值率';",
            "ALTER TABLE `ttpos_material` MODIFY COLUMN `init_stock` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '期初库存';",
            "ALTER TABLE `ttpos_material` MODIFY COLUMN `price` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '采购单价';",
            "ALTER TABLE `ttpos_material` MODIFY COLUMN `stock_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '库存数量';",
            "ALTER TABLE `ttpos_material` MODIFY COLUMN `actual_sale_num` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '实际销量。每次卖出时,实际销量增加';",
            "ALTER TABLE `ttpos_material_unit` MODIFY COLUMN `conversion_rate` decimal(20,4) NULL DEFAULT 1.0000 COMMENT '转换率';",
            "ALTER TABLE `ttpos_member` MODIFY COLUMN `point` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '积分';",
            "ALTER TABLE `ttpos_member` MODIFY COLUMN `frozen_point` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '冻结积分。冻结积分不能使用，在前端显示为已扣除或已增加。冻结积分可为负数。积分余额=积分+冻结积分';",
            "ALTER TABLE `ttpos_member` MODIFY COLUMN `accumulated_get_point` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '累计获取积分';",
            "ALTER TABLE `ttpos_member` MODIFY COLUMN `accumulated_consumption_get_point` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '累计消费获取积分(只存消费赠送积分，不存充值与活动赠送积分)';",
            "ALTER TABLE `ttpos_member` MODIFY COLUMN `accumulated_consumption_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '累计消费金额';",
            "ALTER TABLE `ttpos_member` MODIFY COLUMN `balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '余额';",
            "ALTER TABLE `ttpos_member` MODIFY COLUMN `frozen_balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '冻结余额。冻结余额不能使用，在前端显示为已扣除或已增加。冻结余额可为负数。会员余额=余额+冻结余额';",
            "ALTER TABLE `ttpos_member` MODIFY COLUMN `gift_balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '赠送账户余额';",
            "ALTER TABLE `ttpos_member` MODIFY COLUMN `frozen_gift_balance` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '冻结赠送账户余额。冻结赠送账户余额不能使用，在前端显示为已扣除或已增加。冻结赠送账户余额可为负数。赠送账户余额=赠送账户余额+冻结赠送账户余额';",
            "ALTER TABLE `ttpos_member` MODIFY COLUMN `accumulated_recharge_amount` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '累计充值金额';",
            "ALTER TABLE `ttpos_member_balance_log` MODIFY COLUMN `money` decimal(20,4) NOT NULL DEFAULT 0.0000 COMMENT '变动金额,负数:减余额 正数:加余额。包含赠送余额';",
            "ALTER TABLE `ttpos_member_balance_log` MODIFY COLUMN `gift_money` decimal(20,4) NULL DEFAULT 0.0000 COMMENT '变动赠送金额';"
        ];

        // 分批执行SQL语句，避免一次执行太多
        foreach ($allSqls as $index => $sql) {
            try {
                $this->execute($sql);
                echo "执行SQL " . ($index + 1) . "/" . count($allSqls) . " 成功\n";
            } catch (Exception $e) {
                // 记录错误但继续执行
                echo "Warning SQL " . ($index + 1) . ": " . $e->getMessage() . "\n";
            }
        }
    }
}
