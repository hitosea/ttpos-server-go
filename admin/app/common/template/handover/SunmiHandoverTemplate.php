<?php

namespace app\common\template\handover;

use help\DateHelp;
use app\common\library\helper;
use app\common\model\shop\User;
use app\common\template\BaseTemplate;
use app\common\model\shop\UserShiftLog;
use app\common\model\supplier\Supplier;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\model\settings\PrinterTemplate;
use app\cashier\model\order\Order as CashierOrderModel;
use app\common\library\printer\party\SunmiCloudPrinter;
use app\common\model\product\Category as CategoryModel;

/**
 * Sunmi 交班单模版
 */
class SunmiHandoverTemplate extends BaseTemplate
{
    /**
     * 生成模版
     */
    public function create($data, $shopName, $printerType, $isPrePrint)
    {
        $template = PrinterTemplate::getTemplate(1);
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
        $user = User::where('shop_user_id', $data['shift_user_id'])->find();
        $categorys = (new UserShiftLog)->getSalesInfo($data['shift_user_id'], $data['shop_supplier_id'], $startTime, $endTime);
        $businessData = (new CashierOrderModel)->businessData([
            'mode' => 0,
            'shop_supplier_id' => $data['shop_supplier_id'],
            'cashier_id' => $data['shift_user_id'],
            'time' => [date('Y-m-d H:i:s', $startTime), date('Y-m-d H:i:s', $endTime)]
        ]);
        $abnormalData = $data['abnormal'] ?? [];

        // 佛历
        $startTime = date('Y/m/d H:i:s', $startTime);
        $endTime = date('Y/m/d H:i:s', $endTime);
        if ($this->defaultCalendar == '3') {
            $startTime = DateHelp::changeBuddhistCalendar($startTime);
            $endTime = DateHelp::changeBuddhistCalendar($endTime);
        }

        // 是否自己打印
        $isOneself = $printerType != PrinterTypeEnum::SUNMI_LAN && $printerType != PrinterTypeEnum::SUNMI_CLOUD;
        $lineSpacing = $printerType == PrinterTypeEnum::SUNMI_LAN || $printerType == PrinterTypeEnum::SUNMI_CLOUD ? 20 : 20;
        $printer = new SunmiCloudPrinter(567);
        $printer->lineFeed();
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        $printer->appendText("{$shopName}\n");

        /* *
        * 模版二
        */
        if ($template == 2 || $template == 3) {
            // 交班信息
            $printer->setPrintModes(true, true, false);
            $printer->setLineSpacing(34);
            $printer->lineFeed();
            $printer->appendText(__("交班单") . "\n");
            $printer->lineFeed();
            $printer->setPrintModes(false, false, false);
            $printer->appendText($startTime . " " . __("至") . " " . $endTime);
            $printer->lineFeed();
            $printer->lineFeed();
            $printer->setLineSpacing($lineSpacing);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setupColumns(
                [220, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->printInColumns(__("当班编号"), $data['shift_no']);
            $printer->lineFeed();
            $printer->printInColumns(__("交班人"), $user->real_name);
            $printer->lineFeed();
            // 营业数据
            $printer->setupColumns(
                [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->printInColumns(__("总销售额"), $this->getPriceAndUnit($businessData['all']['receivable_price']));
            $printer->lineFeed();
            $printer->printInColumns(__("实收金额"), $this->getPriceAndUnit($businessData['all']['received_price']));
            $printer->lineFeed();
            // 支付方式
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2);
            $printer->setupColumns(
                [($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? 240 : 280, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? ($this->lang == 'en' ? 200 : 180) : ($this->lang == 'my' ? 180 : 120), SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, $this->lang == 'my' ? SunmiCloudPrinter::ALIGN_LEFT : SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->setPrintModes(true, false, false);
            $printer->printInColumns(__("支付方式"), __('订单数'), __('金额'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setupColumns(
                [300, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [120, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $totalPayPrice = 0;
            foreach ($businessData['incomes'] as $key => $income) {
                if ($income['pay_type'] !== -1) {
                    $printer->printInColumns($income['pay_type_way'] . '',  $income['order_num'] . '',  $this->getPriceAndUnit($income['price']));
                    $printer->lineFeed();
                    $totalPayPrice += $income['price'];
                }
            }
            if ($totalPayPrice > 0) {
                $printer->printInColumns(__("总金额"), '',  $this->getPriceAndUnit($totalPayPrice));
                $printer->lineFeed();
            }
            // 其他费用
            $printer->setupColumns(
                [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2);
            $printer->setupColumns(
                [$this->lang == 'my' ? 400 : 320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->printInColumns(__("原商品金额"), $this->getPriceAndUnit($businessData['all']['not_tax_total_product_price']));
            $printer->lineFeed();
            $printer->printInColumns(__("支付手续费"), $this->getPriceAndUnit($businessData['all']['pay_fee_money']));
            $printer->lineFeed();
            $printer->setupColumns(
                [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->printInColumns(__("服务费"), $this->getPriceAndUnit($businessData['all']['service_money']));
            $printer->lineFeed();
            $printer->printInColumns(__("税费"), $this->getPriceAndUnit($businessData['all']['consumption_tax_money']));
            $printer->lineFeed();
            // 优惠折扣
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2);
            $printer->printInColumns(__("优惠折扣"), $this->getPriceAndUnit($businessData['all']['discount_money']));
            $printer->lineFeed();
            if ($is_balance == 1 || $businessData['all']['user_discount_money'] > 0) {
                $printer->printInColumns(__("会员折扣"), $this->getPriceAndUnit($businessData['all']['user_discount_money']));
                $printer->lineFeed();
            }
            $printer->printInColumns(__("免单金额"), $this->getPriceAndUnit($businessData['all']['free_order_price']));
            $printer->lineFeed();
            // 退款
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2);
            $printer->printInColumns(__("退款金额"), $this->getPriceAndUnit($businessData['all']['refund_money']));
            $printer->lineFeed();
            $printer->appendText("------------------------------------------------");
            // 异常信息
            if ($template == 3) {
                $printer->setupColumns(
                    [510, SunmiCloudPrinter::ALIGN_LEFT, 0],
                    [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
                );
                $printer->lineFeed(2);
                $printer->printInColumns(__("退菜次数"), helper::amountPermillage($abnormalData['refund_product_times'] ?? 0) . '');
                $printer->lineFeed(1);
                $printer->printInColumns(__("退款次数"), helper::amountPermillage($abnormalData['refund_times'] ?? 0) . '');
                $printer->lineFeed(1);
                $printer->printInColumns(__("反结账次数"), helper::amountPermillage($abnormalData['reverse_settle_times'] ?? 0) . '');
                $printer->lineFeed(1);
                $printer->printInColumns(__("赠菜次数"), helper::amountPermillage($abnormalData['product_free_times'] ?? 0) . '');
                $printer->lineFeed(1);
                $printer->printInColumns(__("免单次数"), helper::amountPermillage($abnormalData['free_order_times'] ?? 0) . '');
                $printer->lineFeed(1);
                $printer->printInColumns(__("转菜次数"), helper::amountPermillage($abnormalData['product_move_times'] ?? 0) . '');
                if ($this->lang != 'my' || !$isOneself) {
                    $printer->lineFeed(1);
                }
                if ($this->lang == 'my') {
                    $printer->setLineSpacing(50);
                }
                $printer->printInColumns(__("单品改价次数"), helper::amountPermillage($abnormalData['change_price_times'] ?? 0) . '');
                $printer->setLineSpacing($lineSpacing);
                if ($this->lang != 'my' || !$isOneself) {
                    $printer->lineFeed(1);
                }
                if ($this->lang == 'my') {
                    $printer->setLineSpacing(50);
                }
                $printer->printInColumns(__("整单改价次数"), helper::amountPermillage($abnormalData['change_order_price_times'] ?? 0) . '');
                $printer->setLineSpacing($lineSpacing);
                if ($this->lang != 'my' || !$isOneself) {
                    $printer->lineFeed(1);
                }
                if ($this->lang == 'my') {
                    $printer->setLineSpacing(50);
                }
                $printer->printInColumns(__("整单折扣次数"), helper::amountPermillage($abnormalData['discount_order_times'] ?? 0) . '');
                $printer->setLineSpacing($lineSpacing);
                if ($this->lang != 'my' || !$isOneself) {
                    $printer->lineFeed(1);
                }
                if ($this->lang == 'my') {
                    $printer->setLineSpacing(50);
                }
                $printer->printInColumns(__("整单抹零次数"), helper::amountPermillage($abnormalData['round_order_times'] ?? 0) . '');
                $printer->setLineSpacing($lineSpacing);
                $printer->lineFeed(1);
                $printer->setLineSpacing($lineSpacing);
                $printer->appendText("------------------------------------------------");
            }
            // 
            $printer->setupColumns(
                [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            // 会员充值
            if ($is_balance == 1 || $businessData['all']['recharge_amount'] > 0) {
                $printer->lineFeed(2); 
                $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
                $printer->setPrintModes(true, false, false);
                $printer->appendText(__('会员数据'));
                $printer->setPrintModes(false, false, false);
                $printer->lineFeed(1);
                $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
                $printer->lineFeed(1);
                $printer->printInColumns(__("充值金额"), $this->getPriceAndUnit($businessData['all']['recharge_amount']));
                $printer->lineFeed(1);
                $printer->printInColumns(__("赠送金额"), $this->getPriceAndUnit($businessData['all']['gift_money']));
                $printer->lineFeed(1);
                $printer->printInColumns(__("赠送积分"), helper::amountPermillage($businessData['all']['gift_points']) ?: '0');
                $printer->lineFeed(1);
                $printer->appendText("------------------------------------------------");
            }
            // 合计
            $printer->lineFeed(2);
            $printer->printInColumns(__("所有订单数"), $businessData['all']['total_order_num'] . '');
            $printer->lineFeed();
            $printer->printInColumns(__("人数"), $businessData['all']['total_people_num'] . '');
            $printer->lineFeed();
            $printer->printInColumns(__("平均订单金额"), $this->getPriceAndUnit($businessData['all']['avg_order_price']));
            $printer->lineFeed();
            // 高峰时间
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2);
            $printer->setPrintModes(true, false, false);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setupColumns(
                [($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? ($this->lang == 'th' ? 180 : 240) : 280, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? ($this->lang == 'en' ? 200 : 180) : ($this->lang == 'my' ? 180 : 120), SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, $this->lang == 'my' ? SunmiCloudPrinter::ALIGN_LEFT : SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->printInColumns(__('高峰时间'), __('订单数'), $this->lang == 'en' ? 'Amount' : __('订单金额'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setupColumns(
                [270, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [96, SunmiCloudPrinter::ALIGN_CENTER, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            foreach ($businessData['all']['peak_hour_list'] as $key => $peak) {
                $printer->printInColumns($peak['time_period'] . '', $peak['num'] . '', $this->getPriceAndUnit($peak['amount']));
                $printer->lineFeed();
            }
            // 分类列表
            $printer->appendText("------------------------------------------------\n");
            $printer->lineFeed();
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setupColumns(
                [300, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [96, SunmiCloudPrinter::ALIGN_CENTER, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->setPrintModes(true, false, false);
            $printer->printInColumns(__("分类"), __("数量"), __("小计"));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed();
            foreach ($categorys as $category) {
                $name = (new CategoryModel)->getNameTextAttr($category['name']);
                $printer->printInColumns($name, "{$category['sales']}",  $this->getPriceAndUnit($category['prices']));
                $printer->lineFeed();
            }
            // 汇总
            $printer->setupColumns(
                [$this->lang == 'en' ? 430 : 400, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->appendText("------------------------------------------------\n");
            $printer->lineFeed();
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("上一班遗留备用金"), $this->getPriceAndUnit($previousShiftCash));
            $printer->setLineSpacing($lineSpacing);
            $printer->lineFeed();
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("中途存入现金"), $this->getPriceAndUnit($depositCash));
            $printer->setLineSpacing($lineSpacing);
            $printer->lineFeed();
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("中途取出现金"), $this->getPriceAndUnit($withdrawCash));
            $printer->setLineSpacing($lineSpacing);
            $printer->lineFeed();
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("本班取出现金"), $this->getPriceAndUnit($cashTakenOut));
            $printer->setLineSpacing($lineSpacing);
            $printer->lineFeed();
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("本班遗留备用金"), $this->getPriceAndUnit($cashLeft));
            $printer->setLineSpacing($lineSpacing);
        }
        /* *
        * 模版一
        */ 
        else {
            $printer->lineFeed();
            $printer->setPrintModes(true, true, false);
            $printer->setLineSpacing(50);
            $printer->appendText(__("交班单") . "\n");
            $printer->lineFeed();
            // 交班信息
            $printer->setLineSpacing($lineSpacing);
            $printer->setPrintModes(false, false, false);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setupColumns(
                [220, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->printInColumns(__("当班编号"), $data['shift_no']);
            $printer->lineFeed();
            $printer->printInColumns(__("交班人"), $user->real_name);
            $printer->lineFeed();
            $printer->printInColumns(__("当班时间"), $startTime . " " . __("至"));
            $printer->setLineSpacing($lineSpacing - 6);
            $printer->lineFeed();
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
            $printer->appendText($endTime);
            $printer->setLineSpacing($lineSpacing);
            $printer->lineFeed(2);
            $printer->setupColumns(
                [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            // 营业数据
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->printInColumns(__("总销售额"), $this->getPriceAndUnit($businessData['all']['receivable_price']));
            $printer->lineFeed();
            $printer->printInColumns(__("原商品金额"), $this->getPriceAndUnit($businessData['all']['not_tax_total_product_price']));
            $printer->lineFeed();
            $printer->setupColumns(
                [$this->lang == 'my' ? 400 : 320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->printInColumns(__("支付手续费"), $this->getPriceAndUnit($businessData['all']['pay_fee_money']));
            $printer->lineFeed();
            $printer->setupColumns(
                [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->printInColumns(__("服务费"), $this->getPriceAndUnit($businessData['all']['service_money']));
            $printer->lineFeed();
            $printer->printInColumns(__("税费"), $this->getPriceAndUnit($businessData['all']['consumption_tax_money']));
            $printer->lineFeed();
            $printer->printInColumns(__("商品数量"), $businessData['all']['product_num'] . '');
            $printer->lineFeed();
            $printer->printInColumns(__("优惠折扣"), $this->getPriceAndUnit($businessData['all']['discount_money']));
            $printer->lineFeed();
            if ($is_balance == 1 || $businessData['all']['user_discount_money'] > 0) {
                $printer->printInColumns(__("会员折扣"), $this->getPriceAndUnit($businessData['all']['user_discount_money']));
                $printer->lineFeed();
            }
            $printer->printInColumns(__("退款金额"), $this->getPriceAndUnit($businessData['all']['refund_money']));
            $printer->lineFeed();
            $printer->printInColumns(__("免单金额"), $this->getPriceAndUnit($businessData['all']['free_order_price']));
            $printer->lineFeed();
            $printer->setPrintModes(true, false, false);
            $printer->setCharacterSize(2, 1);
            $printer->printInColumns(__("实收金额"), $this->getPriceAndUnit($businessData['all']['received_price']));
            $printer->setCharacterSize(1, 1);
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed();
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2);
            // 税收百分比对象列表
            foreach ($businessData['all']['percentage_list'] as $percentage) {
                $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
                $printer->setPrintModes(true, false, false);
                if ($this->lang == 'ja') {
                    $printer->appendText($percentage['tax_rate'] . '%' . __('的对象'));
                } else {
                    $printer->appendText('VAT (' . $percentage['tax_rate'] . '%)');
                }
                $printer->setPrintModes(false, false, false);
                $printer->lineFeed(2);
                $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
                $printer->printInColumns(__("合计"), $this->getPriceAndUnit($percentage['total_price']));
                $printer->lineFeed(1);
                $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
                if ($this->lang == 'ja') {
                    $printer->appendText("(" . __('其中消费税') . '' . $this->getPriceAndUnit($percentage['consumption_tax']) . ")");
                } else {
                    $printer->appendText("(" . __('其中VAT') . '' . $this->getPriceAndUnit($percentage['consumption_tax']) . ")");
                }
                $printer->lineFeed(2);
            }
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2);
            // 合计
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('合计'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(2);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->printInColumns(__("所有订单数"), $businessData['all']['total_order_num'] . '');
            $printer->lineFeed(1);
            $printer->printInColumns(__("桌数"), $businessData['all']['total_table_num'] . '');
            $printer->lineFeed(1);
            $printer->printInColumns(__("人数"), $businessData['all']['total_people_num'] . '');
            $printer->lineFeed(1);
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("最小/大订单金额"), $this->getPriceAndUnit($businessData['all']['min_order_price']) . '/' . $this->getPriceAndUnit($businessData['all']['max_order_price']));
            if ($this->lang == 'my') {
                $printer->setLineSpacing($lineSpacing);
            }
            $printer->lineFeed(1);
            $printer->printInColumns(__("平均订单金额"), $this->getPriceAndUnit($businessData['all']['avg_order_price']));
            $printer->lineFeed(2);
            // 桌台方式
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('桌台方式'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(2);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setupColumns(
                [380, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->printInColumns(__("订单数（桌数）"), $businessData['all']['table_order_num'] . '');
            $printer->setupColumns(
                [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->lineFeed(1);
            $printer->printInColumns(__("人数"), $businessData['all']['table_people_num'] . '');
            $printer->lineFeed(1);
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("最小/大订单金额"), $this->getPriceAndUnit($businessData['all']['table_min_order_price']) . '/' . $this->getPriceAndUnit($businessData['all']['table_max_order_price']));
            if ($this->lang == 'my') {
                $printer->setLineSpacing($lineSpacing);
            }
            $printer->lineFeed(1);
            $printer->printInColumns(__("平均订单金额"), $this->getPriceAndUnit($businessData['all']['table_avg_order_price']));
            $printer->lineFeed(1);
            $printer->printInColumns(__("人均"), $this->getPriceAndUnit($businessData['all']['table_people_avg']));
            $printer->lineFeed(2);
            // 收银方式
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('点餐方式'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(2);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->printInColumns(__("订单数"), $businessData['all']['cashier_order_num'] . '');
            $printer->lineFeed(1);
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("最小/大订单金额"), $this->getPriceAndUnit($businessData['all']['cashier_min_order_price']) . '/' . $this->getPriceAndUnit($businessData['all']['cashier_max_order_price']));
            if ($this->lang == 'my') {
                $printer->setLineSpacing($lineSpacing);
            }
            $printer->lineFeed(1);
            $printer->printInColumns(__("平均订单金额"), $this->getPriceAndUnit($businessData['all']['cashier_avg_order_price']));
            $printer->lineFeed(1);
            // 支付方式
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2);
            $printer->setupColumns(
                [($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? 240 : 280, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? ($this->lang == 'en' ? 200 : 180) : ($this->lang == 'my' ? 180 : 120), SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, $this->lang == 'my' ? SunmiCloudPrinter::ALIGN_LEFT : SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->setPrintModes(true, false, false);
            $printer->printInColumns(__("支付方式"), __('订单数'), __('金额'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setupColumns(
                [300, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [120, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $totalPayPrice = 0;
            foreach ($businessData['incomes'] as $key => $income) {
                if ($income['pay_type'] !== -1) {
                    $printer->printInColumns($income['pay_type_way'] . '',  $income['order_num'] . '',  $this->getPriceAndUnit($income['price']));
                    $printer->lineFeed();
                    $totalPayPrice += $income['price'];
                }
            }
            if ($totalPayPrice > 0) {
                $printer->printInColumns(__("总金额"), '',  $this->getPriceAndUnit($totalPayPrice));
                $printer->lineFeed();
            }
            // 高峰时间
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2);
            $printer->setPrintModes(true, false, false);
            $printer->setupColumns(
                [($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? ($this->lang == 'th' ? 180 : 240) : 280, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [($this->lang == 'tr' || $this->lang == 'th' || $this->lang == 'en') ? ($this->lang == 'en' ? 200 : 180) : ($this->lang == 'my' ? 180 : 120), SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, $this->lang == 'my' ? SunmiCloudPrinter::ALIGN_LEFT : SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->printInColumns(__('高峰时间'), __('订单数'), $this->lang == 'en' ? 'Amount' : __('订单金额'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setupColumns(
                [270, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [96, SunmiCloudPrinter::ALIGN_CENTER, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            foreach ($businessData['all']['peak_hour_list'] as $key => $peak) {
                $printer->printInColumns($peak['time_period'] . '', $peak['num'] . '', $this->getPriceAndUnit($peak['amount']));
                $printer->lineFeed();
            }
            // 分类
            $printer->appendText("------------------------------------------------\n");
            $printer->lineFeed();
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setupColumns(
                [300, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [96, SunmiCloudPrinter::ALIGN_CENTER, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->setPrintModes(true, false, false);
            $printer->printInColumns(__("分类"), __("数量"), __("小计"));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed();
            foreach ($categorys as $category) {
                $name = (new CategoryModel)->getNameTextAttr($category['name']);
                $printer->printInColumns($name, "{$category['sales']}",  $this->getPriceAndUnit($category['prices']));
                $printer->lineFeed();
            }
            // 汇总
            $printer->setupColumns(
                [400, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->appendText("------------------------------------------------\n");
            $printer->lineFeed();
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns($this->lang == 'en' ? "Actual amount received this shift" : __("本班实收金额"), $this->getPriceAndUnit($businessData['all']['received_price']));
            $printer->setLineSpacing($lineSpacing);
            $printer->lineFeed();
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("上一班遗留备用金"), $this->getPriceAndUnit($previousShiftCash));
            $printer->setLineSpacing($lineSpacing);
            $printer->lineFeed();
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("中途存入现金"), $this->getPriceAndUnit($depositCash));
            $printer->setLineSpacing($lineSpacing);
            $printer->lineFeed();
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("中途取出现金"), $this->getPriceAndUnit($withdrawCash));
            $printer->setLineSpacing($lineSpacing);
            $printer->lineFeed();
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("本班取出现金"), $this->getPriceAndUnit($cashTakenOut));
            $printer->setLineSpacing($lineSpacing);
            $printer->lineFeed();
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("本班遗留备用金"), $this->getPriceAndUnit($cashLeft));
            $printer->setLineSpacing($lineSpacing);
        }
        //
        $printer->lineFeed(4);
        $printer->printAndExitPageMode();
        $printer->lineFeed(4);
        $printer->cutPaper(false);
        // 打开钱箱
        if (!$isPrePrint) {
            $printer->appendText(chr(27) . chr(112) . chr(0) . chr(25) . chr(250));
        }
        //
        return $printer->orderData;
    }
}
