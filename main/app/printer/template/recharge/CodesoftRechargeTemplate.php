<?php

namespace app\common\template\recharge;

use help\DateHelp;
use app\common\library\helper;
use app\common\template\BaseTemplate;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\model\order\UserRechargeOrder;
use app\common\library\printer\party\SunmiCloudPrinter;

/**
 * Codesoft 充值单模版
 */
class CodesoftRechargeTemplate extends BaseTemplate
{
    /**
     * 生成模版
     */
    public function create(UserRechargeOrder $order, $printType)
    {
        $settingStore = $this->setting[SettingEnum::STORE]['values'];
        $shopName = $settingStore['name'] ?? '';

        // 余额日志
        $balanceLog = $order->balanceLog()->where('scene', 10)->find();

        // 日历
        $payTime = date('Y/m/d H:i:s', strtotime($order['update_time']));
        if ($this->defaultCalendar == '3') {
            $payTime = DateHelp::changeBuddhistCalendar($payTime);
        }

        /* *
        * 打印模版
        */
        $width = 48;
        $printer = new SunmiCloudPrinter(567);
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
        $printer->appendText(__("充值单"));
        $printer->lineFeed(1);
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        $printer->setPrintModes(true, true, false);
        $printer->setCharacterSize(2, 2);
        $printer->appendText("{$shopName}\n");
        $printer->setPrintModes(false, false, false);
        $printer->setCharacterSize(1, 1);
        $printer->lineFeed(1);
        // 
        $printer->setLineSpacing(90);
        $printer->appendText(printText(__("收银员"), '', $order->cashier?->real_name ?: $order->cashier?->user_name ?: '', $width));
        $printer->appendText(printText(__("时间"), '', $payTime, $width));
        $printer->appendText(printText(__("充值前"), '', $this->getPriceAndUnit($balanceLog?->before_money ?: 0), $width));
        $printer->appendText(printText(__("本次充值"), '', $this->getPriceAndUnit($order['recharge_money']), $width));
        $printer->appendText(printText(__("赠送金额"), '', $this->getPriceAndUnit($order['gift_money']), $width));
        $printer->appendText(printText(__("赠送积分"), '', helper::amountPermillage($order['gift_point']) . '', $width));
        $printer->appendText(printText(__("充值后"), '', $this->getPriceAndUnit($balanceLog?->balanceLog ?: 0), $width));
        // 
        $printer->setLineSpacing(($this->currencyUnit == '$' && $this->lang == 'en') ? 70 :  20);
        // 退款
        if ($order['refund_money'] && $order['refund_money'] > 0) {
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(1);
            $printer->appendText(printText(__("退款"), '', $this->getPriceAndUnit($order['refund_money']), $width));
        }
        // 合计應收：
        $printer->appendText("------------------------------------------------\n");
        $printer->setCharacterSize(2, 1);
        $printer->setPrintModes(true, true, false);
        $printer->appendText(printText(__("合计应收："), '', $this->getPriceAndUnit($order['actual_price']), $width));
        $printer->setPrintModes(false, false, false);
        $printer->setCharacterSize(1, 1);
        $printer->setLineSpacing(70);
        $printer->appendText("------------------------------------------------\n");
        $printer->setLineSpacing(90);
        // 支付方式
        $payTypes = $order->payType()->select()->append(['pay_type'])->toArray();
        foreach ($payTypes as $payType) {
            $printer->appendText(printText(__("支付方式"), '', $payType['pay_type']['text'], $width, 20, 0, 28) . "\n");
            $printer->appendText(printText(__("实收金额"), '', $this->getPriceAndUnit($payType['price']), $width, 34));
            if ($order->is_merge != 1 && $payType['pay_type']['value'] == OrderPayTypeEnum::CASH) {
                $printer->appendText(printText(__("找零"), '', helper::amountPermillage($order['change_due'] ?? 0) . '', $width, 34));
            }
        }
        // Print and exit page mode
        $printer->restoreDefaultLineSpacing();
        $printer->lineFeed();
        $printer->printAndExitPageMode();
        $printer->lineFeed(6);
        $printer->cutPaper(true);
        // 
        return $printer->orderData;
    }
}
