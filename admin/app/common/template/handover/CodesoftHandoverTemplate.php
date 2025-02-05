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
 * Codesoft 交班单模版
 */
class CodesoftHandoverTemplate extends BaseTemplate
{
    /**
     * 生成模版
     */
    public function create($printerType, $data, $shopName, $isPrePrint)
    {
        $template = PrinterTemplate::getTemplate(1);
        $is_balance = Supplier::where('shop_supplier_id', $data['shop_supplier_id'])->value('is_open_member') ?: 0;
        //
        $name = __('人');
        $width = 48;
        $leftWidth = 27;
        $centerWidth = 12;
        $rightWidth = 19;
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

        //
        $printer = new SunmiCloudPrinter(567);

        /* *
        * 模版二
        */
        if ($template == 2 || $template == 3) {
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->appendText("{$shopName}\n");
            $printer->lineFeed();
            $printer->setLineSpacing(80);
            $printer->setCharacterSize(2, 2);
            $printer->appendText(__("交班单"));
            $printer->setCharacterSize(1, 1);
            $printer->lineFeed();
            if ($this->lang != 'th') {
                $printer->lineFeed(1);
            }
            $printer->appendText($startTime . " " . __("至") . " " . $endTime);
            $printer->lineFeed();
            $printer->lineFeed();
            $printer->setPrintModes(false, false, false);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("当班编号"), '', $data['shift_no'], $width));
            $printer->lineFeed();
            $printer->appendText(printText(__("交班人"), '', $user->real_name, $width));
            $printer->lineFeed();
            // 营业数据
            $printer->setLineSpacing($this->lang == 'th' ? 30 : 90);
            $printer->appendText(printText(__("总销售额"), '', $this->getPriceAndUnit($businessData['all']['receivable_price']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("实收金额"), '', $this->getPriceAndUnit($businessData['all']['received_price']), $width));
            $printer->lineFeed(1);
            // 支付方式
            $printer->appendText("------------------------------------------------");
            $printer->setPrintModes(true, false, false);
            $printer->appendText(printText(__('支付方式'), __('订单数'), __('金额'), $width, 24 - (($this->lang == 'en' || $this->lang == 'th' || $this->lang == 'tr') ? 4 : 0), 20, 16));
            $printer->setPrintModes(false, false, false);
            $totalPayPrice = 0;
            foreach ($data['incomes'] as $income) {
                if ($income['pay_type'] !== -1) {
                    $printer->appendText(printText($income['pay_type_way'], $income['order_num'], $this->getPriceAndUnit($income['price']), $width, 26, 10, 18));
                    $printer->lineFeed($this->lang == 'th' ? 2 : 1);
                    $totalPayPrice += $income['price'];
                }
            }
            if ($totalPayPrice > 0) {
                $printer->appendText(printText(__("总金额"), '', $this->getPriceAndUnit($totalPayPrice), $width, 26, 10, 18));
                $printer->lineFeed($this->lang == 'th' ? 2 : 1);
            }
            // 其他费用
            $printer->appendText("------------------------------------------------");
            $printer->appendText(printText(__("原商品金额"), '', $this->getPriceAndUnit($businessData['all']['not_tax_total_product_price']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("支付手续费"), '', $this->getPriceAndUnit($businessData['all']['pay_fee_money']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("服务费"), '', $this->getPriceAndUnit($businessData['all']['service_money']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("税费"), '', $this->getPriceAndUnit($businessData['all']['consumption_tax_money']), $width));
            $printer->lineFeed(1);
            // 优惠折扣
            $printer->appendText("------------------------------------------------");
            $printer->appendText(printText(__("优惠折扣"), '', $this->getPriceAndUnit($businessData['all']['discount_money']), $width));
            $printer->lineFeed(1);
            if ($is_balance == 1 || $businessData['all']['user_discount_money'] > 0) {
                $printer->appendText(printText(__("会员折扣"), '', $this->getPriceAndUnit($businessData['all']['user_discount_money']), $width));
                $printer->lineFeed(1);
            }
            $printer->appendText(printText(__("免单金额"), '', $this->getPriceAndUnit($businessData['all']['free_order_price']), $width));
            $printer->lineFeed(2);
            // 退款
            $printer->appendText("------------------------------------------------");
            $printer->appendText(printText(__("退款金额"), '', $this->getPriceAndUnit($businessData['all']['refund_money']), $width));
            $printer->lineFeed($this->lang == 'th' ? 2 : 1);
            $printer->appendText("------------------------------------------------");
            // 异常信息
            if ($template == 3) {
                $printer->setupColumns(
                    [510, SunmiCloudPrinter::ALIGN_LEFT, 0],
                    [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
                );
                $printer->printInColumns(__("退菜次数"), helper::amountPermillage($abnormalData['refund_product_times'] ?? 0) . '');
                $printer->printInColumns(__("退款次数"), helper::amountPermillage($abnormalData['refund_times'] ?? 0) . '');
                $printer->printInColumns(__("反结账次数"), helper::amountPermillage($abnormalData['reverse_settle_times'] ?? 0) . '');
                $printer->printInColumns(__("赠菜次数"), helper::amountPermillage($abnormalData['product_free_times'] ?? 0) . '');
                $printer->printInColumns(__("免单次数"), helper::amountPermillage($abnormalData['free_order_times'] ?? 0) . '');
                $printer->printInColumns(__("转菜次数"), helper::amountPermillage($abnormalData['product_move_times'] ?? 0) . '');
                $printer->printInColumns(__("单品改价次数"), helper::amountPermillage($abnormalData['change_price_times'] ?? 0) . '');
                $printer->printInColumns(__("整单改价次数"), helper::amountPermillage($abnormalData['change_order_price_times'] ?? 0) . '');
                $printer->printInColumns(__("整单折扣次数"), helper::amountPermillage($abnormalData['discount_order_times'] ?? 0) . '');
                $printer->printInColumns(__("整单抹零次数"), helper::amountPermillage($abnormalData['round_order_times'] ?? 0) . '');
                $printer->appendText("------------------------------------------------");
            }
            // 会员充值
            if ($is_balance == 1 || $businessData['all']['recharge_amount'] > 0) {
                $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
                $printer->setPrintModes(true, false, false);
                $printer->appendText(__('会员数据'));
                $printer->setPrintModes(false, false, false);
                $printer->lineFeed(1);
                $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
                $printer->appendText(printText(__("充值金额"), '', $this->getPriceAndUnit($businessData['all']['recharge_amount']), $width));
                $printer->lineFeed(1);
                $printer->appendText(printText(__("赠送金额"), '', $this->getPriceAndUnit($businessData['all']['gift_money']), $width));
                $printer->lineFeed(1);
                $printer->appendText(printText(__("赠送积分"), '', helper::amountPermillage($businessData['all']['gift_points']) ?: '0', $width));
                $printer->lineFeed(1);
                $printer->appendText("------------------------------------------------");
            }
            // 合计
            $printer->lineFeed(1);
            $printer->appendText(printText(__("所有订单数"), '', floatval($businessData['all']['total_order_num']) ?: '0', $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("人数"), '', floatval($businessData['all']['total_people_num']) ?: '0', $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("平均订单金额"), '', $this->getPriceAndUnit($businessData['all']['avg_order_price']), $width));
            $printer->lineFeed(1);
            // 高峰时间
            $printer->appendText("------------------------------------------------");
            $printer->setPrintModes(true, false, false);
            $printer->appendText(printText(__('高峰时间'), __('订单数'), $this->lang == 'en' ? 'Amount' : __('订单金额'), $width, 24 - (($this->lang == 'en' || $this->lang == 'th' || $this->lang == 'tr') ? 4 : 0), 20, 18));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed();
            foreach ($businessData['all']['peak_hour_list'] as $peak) {
                if ($this->lang == 'th') {
                    $printer->lineFeed(1);
                }
                $printer->appendText(printText($peak['time_period'], $peak['num'], $this->getPriceAndUnit($peak['amount']), $width, 26, 10, 18));
                if ($this->lang == 'th') {
                    $printer->lineFeed(1);
                }
            }
            // 分类列表
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText("\n------------------------------------------------\n");
            $printer->setPrintModes(true, false, false);
            $printer->appendText(printText(__("分类"), __("数量"), __("小计"), $width, $leftWidth - ($this->lang == 'tr' ? 2 : 0)));
            $printer->setPrintModes(false, false, false);
            $printer->setLineSpacing(100);
            $printer->lineFeed();
            foreach ($categorys as $category) {
                $name = (new CategoryModel)->getNameTextAttr($category['name']);
                $printer->appendText(printText($name, $category['sales'], $this->getPriceAndUnit($category['prices']), $width, $leftWidth, $centerWidth, $rightWidth));
                $printer->lineFeed();
            }
            $printer->setLineSpacing($this->lang == 'th' ? 20 : 90);
            // 汇总
            $leftWidth = 35;
            $printer->appendText("------------------------------------------------");
            $printer->appendText(printText(__("上一班遗留备用金"), '', $this->getPriceAndUnit($previousShiftCash), $width, $leftWidth));
            $printer->lineFeed();
            $printer->appendText(printText(__("中途存入现金"), '', $this->getPriceAndUnit($depositCash), $width, $leftWidth));
            $printer->lineFeed();
            $printer->appendText(printText(__("中途取出现金"), '', $this->getPriceAndUnit($withdrawCash), $width, $leftWidth));
            $printer->lineFeed();
            $printer->appendText(printText(__("本班取出现金"), '', $this->getPriceAndUnit($cashTakenOut), $width, $leftWidth));
            $printer->lineFeed();
            $printer->appendText(printText(__("本班遗留备用金"), '', $this->getPriceAndUnit($cashLeft), $width, $leftWidth));
        }
        /* *
        * 模版一
        */ 
        else {
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->appendText("{$shopName}\n");
            $printer->lineFeed();
            $printer->setLineSpacing(80);
            $printer->setCharacterSize(2, 2);
            $printer->appendText(__("交班单"));
            $printer->setCharacterSize(1, 1);
            $printer->lineFeed();
            $printer->lineFeed();
            $printer->setPrintModes(false, false, false);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("当班编号"), '', $data['shift_no'], $width));
            $printer->lineFeed();
            $printer->appendText(printText(__("交班人"), '', $user->real_name, $width));
            $printer->lineFeed();
            $printer->appendText(printText(__("当班时间"), '', $startTime . " " . __("至"), $width));
            $printer->lineFeed();
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
            $printer->appendText($endTime);
            $printer->lineFeed();
            // 营业数据
            $printer->setLineSpacing($this->lang == 'th' ? 20 : 90);
            $printer->appendText(printText(__("总销售额"), '', $this->getPriceAndUnit($businessData['all']['receivable_price']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("原商品金额"), '', $this->getPriceAndUnit($businessData['all']['not_tax_total_product_price']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("支付手续费"), '', $this->getPriceAndUnit($businessData['all']['pay_fee_money']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("服务费"), '', $this->getPriceAndUnit($businessData['all']['service_money']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("税费"), '', $this->getPriceAndUnit($businessData['all']['consumption_tax_money']), $width));
            $printer->lineFeed(1);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("商品数量"), '', $businessData['all']['product_num'], $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("优惠折扣"), '', $this->getPriceAndUnit($businessData['all']['discount_money']), $width));
            $printer->lineFeed(1);
            if ($is_balance == 1 || $businessData['all']['user_discount_money'] > 0) {
                $printer->appendText(printText(__("会员折扣"), '', $this->getPriceAndUnit($businessData['all']['user_discount_money']), $width));
                $printer->lineFeed(1);
            }
            $printer->appendText(printText(__("免单金额"), '', $this->getPriceAndUnit($businessData['all']['free_order_price']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("退款金额"), '', $this->getPriceAndUnit($businessData['all']['refund_money']), $width));
            $printer->lineFeed(1);
            $printer->setPrintModes(true, true, false);
            $printer->setCharacterSize(2, 1);
            $printer->appendText(printText(__("实收金额"), '', $this->getPriceAndUnit($businessData['all']['received_price']), $width));
            $printer->setCharacterSize(1, 1);
            $printer->setPrintModes(false, false, false);
            $printer->appendText("\n------------------------------------------------");
            $printer->lineFeed(1);
            // 税收百分比对象列表
            foreach ($businessData['all']['percentage_list'] as $key => $percentage) {
                $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
                $printer->setPrintModes(true, false, false);
                if ($this->lang == 'ja') {
                    $printer->appendText($percentage['tax_rate'] . '%' . __('的对象'));
                } else {
                    $printer->appendText('VAT (' . $percentage['tax_rate'] . '%)');
                }
                $printer->setPrintModes(false, false, false);
                $printer->lineFeed(1);
                $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
                $printer->appendText(printText(__("合计"), '', $this->getPriceAndUnit($percentage['total_price']) . '', $width));
                $printer->lineFeed(1);
                $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
                if ($this->lang == 'ja') {
                    $printer->appendText("(" . __('其中消费税') . '' . $this->getPriceAndUnit($percentage['consumption_tax']) . ")");
                } else {
                    $printer->appendText("(" . __('其中VAT') . '' . $this->getPriceAndUnit($percentage['consumption_tax']) . ")");
                }
                $printer->lineFeed(1);
            }
            $printer->appendText("------------------------------------------------");
            // 合计
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('合计'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("所有订单数"), '', floatval($businessData['all']['total_order_num']) ?: '0', $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("桌数"), '', floatval($businessData['all']['total_table_num']) ?: '0', $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("人数"), '', floatval($businessData['all']['total_people_num']) ?: '0', $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("最小/大订单金额"), '', $this->getPriceAndUnit($businessData['all']['min_order_price']) . '/' . $this->getPriceAndUnit($businessData['all']['max_order_price']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("平均订单金额"), '', $this->getPriceAndUnit($businessData['all']['avg_order_price']), $width));
            $printer->lineFeed(1);
            // 桌台方式
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('桌台方式'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("订单数（桌数）"), '', floatval($businessData['all']['table_order_num']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("人数"), '', floatval($businessData['all']['table_people_num']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("最小/大订单金额"), '', $this->getPriceAndUnit($businessData['all']['table_min_order_price']) . '/' . $this->getPriceAndUnit($businessData['all']['table_max_order_price']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("平均订单金额"), '', $this->getPriceAndUnit($businessData['all']['table_avg_order_price']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("人均"), '', $this->getPriceAndUnit($businessData['all']['table_people_avg']), $width));
            $printer->lineFeed(1);
            // 收银方式
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('点餐方式'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("订单数"), '', floatval($businessData['all']['cashier_order_num']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("最小/大订单金额"), '', $this->getPriceAndUnit($businessData['all']['cashier_min_order_price']) . '/' . $this->getPriceAndUnit($businessData['all']['cashier_max_order_price']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("平均订单金额"), '', $this->getPriceAndUnit($businessData['all']['cashier_avg_order_price']), $width));
            $printer->lineFeed(1);
            // 支付方式
            $printer->appendText("------------------------------------------------");
            $printer->setPrintModes(true, false, false);
            $printer->appendText(printText(__('支付方式'), __('订单数'), __('金额'), $width, 24 - (($this->lang == 'en' || $this->lang == 'th' || $this->lang == 'tr') ? 4 : 0), 20, 16));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $totalPayPrice = 0;
            foreach ($data['incomes'] as $income) {
                if ($income['pay_type'] !== -1) {
                    $printer->appendText(printText($income['pay_type_way'], $income['order_num'], $this->getPriceAndUnit($income['price']), $width, 26, 10, 18));
                    $printer->lineFeed($this->lang == 'th' ? 2 : 1);
                    $totalPayPrice += $income['price'];
                }
            }
            if ($totalPayPrice > 0) {
                $printer->appendText(printText(__("总金额"), '', $this->getPriceAndUnit($totalPayPrice), $width, 26, 10, 18));
                $printer->lineFeed($this->lang == 'th' ? 2 : 1);
            }
            // 高峰时间
            $printer->appendText("------------------------------------------------");
            $printer->setPrintModes(true, false, false);
            $printer->appendText(printText(__('高峰时间'), __('订单数'), $this->lang == 'en' ? 'Amount' : __('订单金额'), $width, 24 - (($this->lang == 'en' || $this->lang == 'th' || $this->lang == 'tr') ? 4 : 0), 20, 18));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed();
            foreach ($businessData['all']['peak_hour_list'] as $peak) {
                if ($this->lang == 'th') {
                    $printer->lineFeed(1);
                }
                $printer->appendText(printText($peak['time_period'], $peak['num'], $this->getPriceAndUnit($peak['amount']), $width, 26, 10, 18));
                if ($this->lang == 'th') {
                    $printer->lineFeed(1);
                }
            }
            // 分类列表
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText("\n------------------------------------------------\n");
            $printer->setPrintModes(true, false, false);
            $printer->appendText(printText(__("分类"), __("数量"), __("小计"), $width, $leftWidth - ($this->lang == 'tr' ? 2 : 0)));
            $printer->setPrintModes(false, false, false);
            $printer->setLineSpacing(100);
            $printer->lineFeed();
            foreach ($categorys as $category) {
                $name = (new CategoryModel)->getNameTextAttr($category['name']);
                $printer->appendText(printText($name, $category['sales'], $this->getPriceAndUnit($category['prices']), $width, $leftWidth, $centerWidth, $rightWidth));
                $printer->lineFeed();
            }
            $printer->setLineSpacing($this->lang == 'th' ? 20 : 90);
            // 汇总
            $leftWidth = 35;
            $printer->appendText("------------------------------------------------\n");
            $printer->appendText(printText(__("本班实收金额"), '',  $this->getPriceAndUnit($businessData['all']['received_price']), $width, $leftWidth));
            $printer->lineFeed();
            $printer->appendText(printText(__("上一班遗留备用金"), '', $this->getPriceAndUnit($previousShiftCash), $width, $leftWidth));
            $printer->lineFeed();
            $printer->appendText(printText(__("中途存入现金"), '', $this->getPriceAndUnit($depositCash), $width, $leftWidth));
            $printer->lineFeed();
            $printer->appendText(printText(__("中途取出现金"), '', $this->getPriceAndUnit($withdrawCash), $width, $leftWidth));
            $printer->lineFeed();
            $printer->appendText(printText(__("本班取出现金"), '', $this->getPriceAndUnit($cashTakenOut), $width, $leftWidth));
            $printer->lineFeed();
            $printer->appendText(printText(__("本班遗留备用金"), '', $this->getPriceAndUnit($cashLeft), $width, $leftWidth));
        }
        //
        $printer->setLineSpacing(90);
        $printer->lineFeed(2);
        $printer->printAndExitPageMode();
        $printer->lineFeed(4);
        $printer->cutPaper(true);
        // 打开钱箱
        if (!$isPrePrint) {
            $printer->appendText("\x10\x14\x01\x00\x01");
        }
        //
        return $printer->orderData;
    }
}
