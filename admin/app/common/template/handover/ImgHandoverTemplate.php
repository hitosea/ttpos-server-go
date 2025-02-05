<?php

namespace app\common\template\handover;

use help\DateHelp;
use base\imgs\ImgFont;
use app\common\library\helper;
use app\common\model\shop\User;
use app\common\template\BaseTemplate;
use app\common\model\shop\UserShiftLog;
use app\common\model\supplier\Supplier;
use app\common\model\settings\PrinterTemplate;
use app\cashier\model\order\Order as CashierOrderModel;
use app\common\library\printer\party\SunmiCloudPrinter;
use app\common\model\product\Category as CategoryModel;

/**
 * 图片打印 交班单模版
 */
class ImgHandoverTemplate extends BaseTemplate
{
    /**
     * 生成模版
     */
    public function create($data, $shopName, $printerType, $isPrePrint)
    {
        $template = PrinterTemplate::getTemplate(1);
        //
        $is_balance = Supplier::where('shop_supplier_id', $data['shop_supplier_id'])->value('is_open_member') ?: 0;
        //
        $name = __('人');
        $startTime = $data['shift_start_time'];
        $endTime = $data['shift_end_time'];
        $previousShiftCash = Helper::number2($data['previous_shift_cash'] ?? 0);
        $withdrawCash = Helper::number2($data['withdraw_cash'] ?? 0);
        $depositCash = Helper::number2($data['deposit_cash'] ?? 0);
        $cashTakenOut = Helper::number2($data['cash_taken_out'] ?? 0);
        $cashLeft = Helper::number2($data['cash_left'] ?? 0);
        //
        $user = User::where('shop_user_id', $data['shift_user_id'])->find();
        $categorys = (new UserShiftLog)->getSalesInfo($data['shift_user_id'], $data['shop_supplier_id'], $startTime, $endTime);
        $businessData = (new CashierOrderModel)->businessData([
            'mode' => 0,
            'shop_supplier_id' => $data['shop_supplier_id'],
            'cashier_id' => $data['shift_user_id'],
            'time' => [date('Y-m-d H:i:s', $startTime), date('Y-m-d H:i:s', $endTime)]
        ]);
        // 佛历
        $startTime = date('Y/m/d H:i:s', $startTime);
        $endTime = date('Y/m/d H:i:s', $endTime);
        if ($this->defaultCalendar == '3') {
            $startTime = DateHelp::changeBuddhistCalendar($startTime);
            $endTime = DateHelp::changeBuddhistCalendar($endTime);
        }
      
        // 
        $abnormalData = $data['abnormal'] ?? [];

        /* *
        * 模版头
        */
        $printer = new ImgFont(568);
        $printer->setTextLineHeight(45);
        $printer->setImagePadding(0);
        $printer->setAlignment(ImgFont::ALIGN_CENTER);
        $printer->appendText("{$shopName}");
        $printer->lineFeed(1, 58);
        /* *
        * 模版二
        */
        if ($template == 2 || $template == 3) {
            $printer->setFontSize(28);
            $printer->appendText(__("交班单"));
            $printer->setFontSize(20);
            $printer->lineFeed();
            $printer->appendText($startTime . " " . __("至") . " " . $endTime);
            $printer->lineFeed();
            $printer->lineFeed(1, 24);
            $printer->printInColumns(
                [__("当班编号"), 350, ImgFont::ALIGN_LEFT],
                [$data['shift_no'], 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("交班人"), 350, ImgFont::ALIGN_LEFT],
                [$user->real_name, 0, ImgFont::ALIGN_RIGHT],
            );
            // 营业数据
            $printer->printInColumns(
                [__("总销售额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['receivable_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->setTextLineHeight(38);
            $printer->printInColumns(
                [__("实收金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['received_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(45);
            // 支付方式
            $printer->printInColumns(
                [__("支付方式"), ($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? 220 : 280, ImgFont::ALIGN_LEFT, 2],
                [__("订单数"), ($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? ($this->lang == 'en' ? 240 : 180) : ($this->lang == 'my' ? 180 : 120), ImgFont::ALIGN_CENTER, 2],
                [__('金额'), 0, ImgFont::ALIGN_RIGHT, 2],
            );
            $totalPayPrice = 0;
            foreach ($data['incomes'] as $key => $income) {
                if ($income['pay_type'] !== -1) {
                    if ($key == count($data['incomes']) - 1) {
                        $printer->setTextLineHeight(34);
                    }
                    $printer->printInColumns(
                        [$income['pay_type_way'], 280, ImgFont::ALIGN_LEFT, 1],
                        [$income['order_num'], 120, ImgFont::ALIGN_CENTER, 1],
                        [$this->getPriceAndUnit($income['price']), 0, ImgFont::ALIGN_RIGHT, 1],
                    );
                    $totalPayPrice += $income['price'];
                }
            }
            if ($totalPayPrice > 0) {
                $printer->lineFeed(1, 10);
                $printer->printInColumns(
                    [__("总金额"), 280, ImgFont::ALIGN_LEFT, 1],
                    ['', 120, ImgFont::ALIGN_CENTER, 1],
                    [$this->getPriceAndUnit($totalPayPrice), 0, ImgFont::ALIGN_RIGHT, 1],
                );
            }
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(45);
            // 其他费用
            $printer->printInColumns(
                [__("原商品金额"), 320, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($businessData['all']['not_tax_total_product_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("支付手续费"), 320, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($businessData['all']['pay_fee_money']), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->printInColumns(
                [__("服务费"), 320, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($businessData['all']['service_money']), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->setTextLineHeight(34);
            $printer->printInColumns(
                [__("税费"), 320, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($businessData['all']['consumption_tax_money']), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            // 优惠折扣
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(45);
            if (!($is_balance == 1 || $businessData['all']['user_discount_money'] > 0)) {
                $printer->setTextLineHeight(34);
            }
            $printer->printInColumns(
                [__("优惠折扣"), 320, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($businessData['all']['discount_money']), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            if ($is_balance == 1 || $businessData['all']['user_discount_money'] > 0) {
                $printer->printInColumns(
                    [__("会员折扣"), 320, ImgFont::ALIGN_LEFT, 1],
                    [$this->getPriceAndUnit($businessData['all']['user_discount_money']), 0, ImgFont::ALIGN_RIGHT, 1],
                );
            }
            $printer->setTextLineHeight(34);
            $printer->printInColumns(
                [__("免单金额"), 320, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($businessData['all']['free_order_price']), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            // 退款
            $printer->setTextLineHeight(40);
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(34);
            $printer->printInColumns(
                [__("退款金额"), 320, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($businessData['all']['refund_money']), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->setTextLineHeight(36);
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(45);
            // 异常信息
            if ($template == 3) {
                if ($this->lang == 'my') {
                    $printer->setTextLineHeight(50);
                }
                $printer->printInColumns(
                    [__("退菜次数"), 480, ImgFont::ALIGN_LEFT, 1],
                    [helper::amountPermillage($abnormalData['refund_product_times'] ?? 0), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->printInColumns(
                    [__("退款次数"), 480, ImgFont::ALIGN_LEFT, 1],
                    [helper::amountPermillage($abnormalData['refund_times'] ?? 0), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->printInColumns(
                    [__("反结账次数"), 480, ImgFont::ALIGN_LEFT, 1],
                    [helper::amountPermillage($abnormalData['reverse_settle_times'] ?? 0), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->printInColumns(
                    [__("赠菜次数"), 480, ImgFont::ALIGN_LEFT, 1],
                    [helper::amountPermillage($abnormalData['product_free_times'] ?? 0), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->printInColumns(
                    [__("免单次数"), 480, ImgFont::ALIGN_LEFT, 1],
                    [helper::amountPermillage($abnormalData['free_order_times'] ?? 0), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->printInColumns(
                    [__("转菜次数"), 480, ImgFont::ALIGN_LEFT, 1],
                    [helper::amountPermillage($abnormalData['product_move_times'] ?? 0), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->printInColumns(
                    [__("单品改价次数"), 540, ImgFont::ALIGN_LEFT, 1],
                    [helper::amountPermillage($abnormalData['change_price_times'] ?? 0), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->printInColumns(
                    [__("整单改价次数"), 540, ImgFont::ALIGN_LEFT, 1],
                    [helper::amountPermillage($abnormalData['change_order_price_times'] ?? 0), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->printInColumns(
                    [__("整单折扣次数"), 500, ImgFont::ALIGN_LEFT, 1],
                    [helper::amountPermillage($abnormalData['discount_order_times'] ?? 0), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->setTextLineHeight($this->lang == 'my' ? 50 : 36);
                $printer->printInColumns(
                    [__("整单抹零次数"), 500, ImgFont::ALIGN_LEFT, 1],
                    [helper::amountPermillage($abnormalData['round_order_times'] ?? 0), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->appendSplitline(true);
                $printer->setTextLineHeight(45);
            }
            // 会员充值
            if ($is_balance == 1 || $businessData['all']['recharge_amount'] > 0) {
                $printer->setAlignment(ImgFont::ALIGN_CENTER);
                $printer->setFontWeight(2);
                $printer->appendText(__('会员数据'));
                $printer->setFontWeight(1);
                $printer->lineFeed(1);
                $printer->setAlignment(ImgFont::ALIGN_LEFT);
                // 
                $printer->printInColumns(
                    [__("充值金额"), 320, ImgFont::ALIGN_LEFT, 1],
                    [$this->getPriceAndUnit($businessData['all']['recharge_amount']), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->printInColumns(
                    [__("赠送金额"), 320, ImgFont::ALIGN_LEFT, 1],
                    [$this->getPriceAndUnit($businessData['all']['gift_money']), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->setTextLineHeight(36);
                $printer->printInColumns(
                    [__("赠送积分"), 320, ImgFont::ALIGN_LEFT, 1],
                    [helper::amountPermillage($businessData['all']['gift_points']), 0, ImgFont::ALIGN_RIGHT, 1],
                );
                $printer->appendSplitline(true);
                $printer->setTextLineHeight(45);
            }
            // 
            $printer->setTextLineHeight(45);
            // 合计
            $printer->printInColumns(
                [__("所有订单数"), 320, ImgFont::ALIGN_LEFT, 1],
                [helper::amountPermillage($businessData['all']['total_order_num']), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->printInColumns(
                [__("人数"), 320, ImgFont::ALIGN_LEFT, 1],
                [helper::amountPermillage($businessData['all']['total_people_num']), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->setTextLineHeight(34);
            $printer->printInColumns(
                [__("平均订单金额"), 320, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($businessData['all']['avg_order_price']), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            // 高峰时间
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(45);
            $printer->printInColumns(
                [__("高峰时间"), ($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? 240 : 300, ImgFont::ALIGN_LEFT, 1],
                [__('订单数'), ($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? ($this->lang == 'en' ? 220 : 180) : ($this->lang == 'my' ? 180 : 120), ImgFont::ALIGN_LEFT, 1],
                [$this->lang == 'en' ? 'Amount' : __('订单金额'), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            foreach ($businessData['all']['peak_hour_list'] as $key => $peak) {
                if ($key == count($businessData['all']['peak_hour_list']) - 1) {
                    $printer->setTextLineHeight(34);
                }
                $printer->printInColumns(
                    [$peak['time_period'], 280, ImgFont::ALIGN_LEFT, 1],
                    [$peak['num'], 120, ImgFont::ALIGN_CENTER, 1],
                    [$this->getPriceAndUnit($peak['amount']), 0, ImgFont::ALIGN_RIGHT, 1],
                );
            }
            // 分类列表
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(45);
            $printer->printInColumns(
                [__("分类"), 280, ImgFont::ALIGN_LEFT, 1],
                [__('数量'), 120, ImgFont::ALIGN_CENTER, 1],
                [__('小计'), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            foreach ($categorys as $key=>$category) {
                if ($key == count($categorys) - 1) {
                    $printer->setTextLineHeight(34);
                }
                $name = (new CategoryModel)->getNameTextAttr($category['name']);
                $printer->printInColumns(
                    [$name, 280, ImgFont::ALIGN_LEFT, 1],
                    [$category['sales'], 120, ImgFont::ALIGN_CENTER, 1],
                    [$this->getPriceAndUnit($category['prices']), 0, ImgFont::ALIGN_RIGHT, 1],
                );
            }
            // 汇总
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(45);
            $printer->printInColumns(
                [__("上一班遗留备用金"), 380, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($previousShiftCash), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->printInColumns(
                [__("中途存入现金"), 360, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($depositCash), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->printInColumns(
                [__("中途取出现金"), 360, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($withdrawCash), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->printInColumns(
                [__("本班取出现金"), 360, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($cashTakenOut), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->printInColumns(
                [__("本班遗留备用金"), 360, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($cashLeft), 0, ImgFont::ALIGN_RIGHT, 1],
            );
        }
        /* *
        * 模版 一
        */ 
        else {
            $printer->setFontSize(28);
            $printer->appendText(__("交班单"));
            $printer->setFontSize(20);
            $printer->lineFeed();
            $printer->lineFeed(1, 24);
            $printer->printInColumns(
                [__("当班编号"), 350, ImgFont::ALIGN_LEFT],
                [$data['shift_no'], 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("交班人"), 350, ImgFont::ALIGN_LEFT],
                [$user->real_name, 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->appendText(__("当班时间"));
            $printer->setAlignment(ImgFont::ALIGN_RIGHT);
            $printer->appendText($startTime . " " . __("至"));
            $printer->lineFeed(1, 32);
            $printer->appendText($endTime);
            $printer->lineFeed();
            // 营业数据
            $printer->printInColumns(
                [__("总销售额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['receivable_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("原商品金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['not_tax_total_product_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("支付手续费"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['pay_fee_money']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("服务费"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['service_money']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("税费"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['consumption_tax_money']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("商品数量"), 350, ImgFont::ALIGN_LEFT],
                [helper::amountPermillage($businessData['all']['product_num']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("优惠折扣"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['discount_money']), 0, ImgFont::ALIGN_RIGHT],
            );
            if ($is_balance == 1 || $businessData['all']['user_discount_money'] > 0) {
                $printer->printInColumns(
                    [__("会员折扣"), 350, ImgFont::ALIGN_LEFT],
                    [$this->getPriceAndUnit($businessData['all']['user_discount_money']), 0, ImgFont::ALIGN_RIGHT],
                );
            }
            $printer->printInColumns(
                [__("退款金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['refund_money']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("免单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['free_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->lineFeed(1, 10);
            if ($this->lang == 'my') {
                $printer->setTextLineHeight(55);
            }
            $printer->printInColumns(
                [__("实收金额"), 350, ImgFont::ALIGN_LEFT, 1, 28],
                [$this->getPriceAndUnit($businessData['all']['received_price']), 0, ImgFont::ALIGN_RIGHT, 1, 28, 38],
            );
            // 税收百分比对象列表
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(45);
            foreach ($businessData['all']['percentage_list'] as $key => $percentage) {
                $printer->setAlignment(ImgFont::ALIGN_LEFT);
                $printer->setFontWeight(2);
                if ($this->lang == 'ja') {
                    $printer->appendText($percentage['tax_rate'] . '%' . __('的对象'));
                } else {
                    $printer->appendText('VAT (' . $percentage['tax_rate'] . '%)');
                }
                $printer->setFontWeight(1);
                $printer->lineFeed(1);
                $printer->printInColumns(
                    [__("合计"), 350, ImgFont::ALIGN_LEFT],
                    [$this->getPriceAndUnit($percentage['total_price']), 0, ImgFont::ALIGN_RIGHT],
                );
                $printer->setAlignment(ImgFont::ALIGN_RIGHT);
                if ($this->lang == 'ja') {
                    $printer->appendText("(" . __('其中消费税') . '' . $this->getPriceAndUnit($percentage['consumption_tax']) . ")");
                } else {
                    $printer->appendText("(" . __('其中VAT') . '' . $this->getPriceAndUnit($percentage['consumption_tax']) . ")");
                }
                $printer->lineFeed(1);
            }
            // 合计
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(45);
            $printer->setAlignment(ImgFont::ALIGN_CENTER);
            $printer->setFontWeight(2);
            $printer->appendText(__('合计'));
            $printer->setFontWeight(1);
            $printer->lineFeed(1);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->printInColumns(
                [__("所有订单数"), 350, ImgFont::ALIGN_LEFT],
                [floatval($businessData['all']['total_order_num']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("桌数"), 350, ImgFont::ALIGN_LEFT],
                [floatval($businessData['all']['total_table_num']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("人数"), 350, ImgFont::ALIGN_LEFT],
                [floatval($businessData['all']['total_people_num']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("最小/大订单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['min_order_price']) . '/' . $this->getPriceAndUnit($businessData['all']['max_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("平均订单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['avg_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            // 桌台方式
            $printer->lineFeed(1, 10);
            $printer->setAlignment(ImgFont::ALIGN_CENTER);
            $printer->setFontWeight(2);
            $printer->appendText(__('桌台方式'));
            $printer->setFontWeight(1);
            $printer->lineFeed(1);
            $printer->printInColumns(
                [__("订单数（桌数）"), 400, ImgFont::ALIGN_LEFT],
                [floatval($businessData['all']['table_order_num']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("人数"), 350, ImgFont::ALIGN_LEFT],
                [floatval($businessData['all']['table_people_num']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("最小/大订单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['table_min_order_price']) . '/' . $this->getPriceAndUnit($businessData['all']['table_max_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("平均订单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['table_avg_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("人均"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['table_people_avg']), 0, ImgFont::ALIGN_RIGHT],
            );
            // 收银方式
            $printer->lineFeed(1, 10);
            $printer->setAlignment(ImgFont::ALIGN_CENTER);
            $printer->setFontWeight(2);
            $printer->appendText(__('点餐方式'));
            $printer->setFontWeight(1);
            $printer->lineFeed(1);
            $printer->printInColumns(
                [__("订单数"), 350, ImgFont::ALIGN_LEFT],
                [floatval($businessData['all']['cashier_order_num']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("最小/大订单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['cashier_min_order_price']) . '/' . $this->getPriceAndUnit($businessData['all']['cashier_max_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("平均订单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($businessData['all']['cashier_avg_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            // 支付方式
            $printer->appendSplitline(true);
            $printer->printInColumns(
                [__("支付方式"), ($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? 240 : 280, ImgFont::ALIGN_LEFT, 2],
                [__("订单数"), ($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? ($this->lang == 'en' ? 220 : 180) : ($this->lang == 'my' ? 180 : 120), ImgFont::ALIGN_CENTER, 2],
                [__('金额'), 0, ImgFont::ALIGN_RIGHT, 2],
            );
            $totalPayPrice = 0;
            foreach ($data['incomes'] as $key => $income) {
                if ($income['pay_type'] !== -1) {
                    if ($key == count($data['incomes']) - 1) {
                        $printer->setTextLineHeight(34);
                    }
                    $printer->printInColumns(
                        [$income['pay_type_way'], 280, ImgFont::ALIGN_LEFT, 1],
                        [$income['order_num'], 120, ImgFont::ALIGN_CENTER, 1],
                        [$this->getPriceAndUnit($income['price']), 0, ImgFont::ALIGN_RIGHT, 1],
                    );
                    $totalPayPrice += $income['price'];
                }
            }
            if ($totalPayPrice > 0) {
                $printer->lineFeed(1, 10);
                $printer->printInColumns(
                    [__("总金额"), 280, ImgFont::ALIGN_LEFT, 1],
                    ['', 120, ImgFont::ALIGN_CENTER, 1],
                    [$this->getPriceAndUnit($totalPayPrice), 0, ImgFont::ALIGN_RIGHT, 1],
                );
            }
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(45);
            // 高峰时间
            $printer->printInColumns(
                [__("高峰时间"), ($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? 240 : 300, ImgFont::ALIGN_LEFT, 1],
                [__('订单数'), ($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? ($this->lang == 'en' ? 220 : 180) : ($this->lang == 'my' ? 180 : 120), ImgFont::ALIGN_LEFT, 1],
                [$this->lang == 'en' ? 'Amount' : __('订单金额'), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            foreach ($businessData['all']['peak_hour_list'] as $key => $peak) {
                if ($key == count($businessData['all']['peak_hour_list']) - 1) {
                    $printer->setTextLineHeight(34);
                }
                $printer->printInColumns(
                    [$peak['time_period'], 280, ImgFont::ALIGN_LEFT, 1],
                    [$peak['num'], 120, ImgFont::ALIGN_CENTER, 1],
                    [$this->getPriceAndUnit($peak['amount']), 0, ImgFont::ALIGN_RIGHT, 1],
                );
            }
            // 分类列表
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(45);
            $printer->printInColumns(
                [__("分类"), 280, ImgFont::ALIGN_LEFT, 1],
                [__('数量'), 120, ImgFont::ALIGN_CENTER, 1],
                [__('小计'), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            foreach ($categorys as $key=>$category) {
                if ($key == count($categorys) - 1) {
                    $printer->setTextLineHeight(34);
                }
                $name = (new CategoryModel)->getNameTextAttr($category['name']);
                $printer->printInColumns(
                    [$name, 280, ImgFont::ALIGN_LEFT, 1],
                    [$category['sales'], 120, ImgFont::ALIGN_CENTER, 1],
                    [$this->getPriceAndUnit($category['prices']), 0, ImgFont::ALIGN_RIGHT, 1],
                );
            }
            // 汇总
            $printer->appendSplitline(true);
            $printer->setTextLineHeight(45);
            $printer->printInColumns(
                [__("本班实收金额"), 380, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($businessData['all']['received_price']), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->printInColumns(
                [__("上一班遗留备用金"), 380, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($previousShiftCash), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->printInColumns(
                [__("中途存入现金"), 360, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($depositCash), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->printInColumns(
                [__("中途取出现金"), 360, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($withdrawCash), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->printInColumns(
                [__("本班取出现金"), 360, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($cashTakenOut), 0, ImgFont::ALIGN_RIGHT, 1],
            );
            $printer->printInColumns(
                [__("本班遗留备用金"), 360, ImgFont::ALIGN_LEFT, 1],
                [$this->getPriceAndUnit($cashLeft), 0, ImgFont::ALIGN_RIGHT, 1],
            );
        }
        // 
        $printer->lineFeed(3);
        // 
        $openMoneybox = false;
        if (!$isPrePrint) {
            $openMoneybox =  $this->isSunmi ? 2 : true;
        }
        return $printer->save('', !$this->isSunmi, $openMoneybox);
    }
}
