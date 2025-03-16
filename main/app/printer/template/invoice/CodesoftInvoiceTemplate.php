<?php

namespace app\common\template\invoice;

use help\DateHelp;
use app\common\library\helper;
use app\common\template\BaseTemplate;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\model\settings\PrinterTemplate;
use app\common\template\bill\CodesoftBillTemplate;
use app\common\model\settings\Setting as SettingModel;
use app\common\library\printer\party\SunmiCloudPrinter;

/**
 * Codesoft 发票模版
 */
class CodesoftInvoiceTemplate extends BaseTemplate
{
    /**
     * 生成模版
     */
    public function create($order, $printType, $isCashierPrinter)
    {
        $template = PrinterTemplate::getTemplate(7);
        $settingStore = $this->setting[SettingEnum::STORE]['values'];
        $shopName = $settingStore['name'] ?? '';

        /* *
        * 模版2
        */
        if ($template == 2) {
            $order->_shop_name = $shopName;
            $order->_template = 3;
            $order->_title = __("发票");
            return (new CodesoftBillTemplate($this->setting, $this->allSourceProductList, $this->isSunmi))->create($order, $printType);
        }

        /* *
        * 模版 1
        */
        $width = 48;
        $leftWidth = 28;
        $isXinyeWifi = $printType == PrinterTypeEnum::XPRINTER_WIFI;
        $lineSpacing = $isXinyeWifi ? "\x1B\x33\x38" : "\x1B\x33\x32";
        //
        $settingStore = $this->setting[SettingEnum::STORE]['values'];
        $settingPrinterConfig = $this->setting[SettingEnum::PRINTER]['values'] ?? [];
        $consumptionTax = $settingPrinterConfig['consumption_tax'] ?? 1;
        $settingCloud = SettingModel::getCloudBasic();
        $brandName = $settingCloud['base']['brand_name'] ?? $settingStore['brand_name'] ?? '';
        //
        $shopName = $settingStore['name'] ?? '';
        $address = $settingStore['address'] ?? '';
        $company = $settingStore['company'] ?? '';
        $phone = $settingStore['phone'] ?? '';
        $taxNumber = $settingStore['tax_number'] ?? '';
        //
        $percentageList = $order->getPercentageList();
        // 佛历
        $payTime = date('Y/m/d H:i:s', $order->pay_time);
        if ($this->defaultCalendar == '3') {
            $payTime = DateHelp::changeBuddhistCalendar($payTime);
        }
        //
        $printer = new SunmiCloudPrinter(567);
        $printer->restoreDefaultLineSpacing();
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        $printer->setCharacterSize(2, 1);
        $printer->appendText("{$shopName}");
        $printer->setCharacterSize(1, 1);
        $printer->setLineSpacing($isCashierPrinter ? 60 : 80);
        $printer->lineFeed();
        if ($isXinyeWifi) {
            $printer->lineFeed();
        }
        $printer->appendText("\x1B\x33\x48");
        $printer->appendText(__("非常感谢您今天的到来，我们期待您的再次光临"));
        $printer->appendText($lineSpacing);
        $printer->lineFeed(2);
        $printer->setCharacterSize(2, 2);
        $printer->appendText(__("发票"));
        $printer->setCharacterSize(1, 1);
        $printer->lineFeed(1);
        $printer->appendText($isXinyeWifi ? "\x1B\x33\x38" : "\x1B\x33\x18");
        $printer->lineFeed(1);
        $printer->appendText($lineSpacing);
        $printer->appendText($payTime);
        $printer->setLineSpacing(60);
        $printer->lineFeed(2);
        //
        if ($this->lang == 'ja') {
            $printer->setLineSpacing(52);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
            $printer->appendText(__("先生/小姐"));
            $printer->lineFeed();
        }
        $printer->appendText("------------------------------------------------\n");
        $printer->setLineSpacing(18);
        if ($isXinyeWifi) {
            $printer->setLineSpacing(40);
        }
        $printer->lineFeed();
        $printer->setCharacterSize(2, 2);
        $printer->appendText(printText(__("合计"), '', $this->getPriceAndUnit($order->actual_receive_price), $width - 24));
        $printer->setCharacterSize(1, 1);
        $printer->lineFeed(2);
        $printer->appendText(printText('(' . __("其中服务费"), '', $this->getPriceAndUnit($order->service_money ?: 0) . ')', $width, $leftWidth));
        $printer->lineFeed(2);
        if ($consumptionTax != 4) {
            $printer->appendText(printText('(' . __("其中VAT"), '', $this->getPriceAndUnit($order->consumption_tax_money ?: 0) . ')', $width, $leftWidth));
            $printer->lineFeed(2);
        }
        if ($isXinyeWifi) {
            $printer->setLineSpacing(60);
        }
        $printer->appendText("------------------------------------------------\n");
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
        // 消费税
        if ($consumptionTax != 4) {
            $printer->setLineSpacing($isXinyeWifi ? 24 : 12);
            $printer->lineFeed();
            $printer->appendText($lineSpacing);
            $printer->appendText(__("仅作为餐饮费收取以上金额"));
            $printer->lineFeed(2);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
            $printer->appendText(__("合计 (其中VAT)"));
            $printer->lineFeed();
            $printer->setLineSpacing($isXinyeWifi ? 32 : 22);
            $printer->lineFeed();
            $printer->appendText($lineSpacing);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            foreach ($percentageList as $key => $percentage) {
                if ($this->lang == 'ja') {
                    $printer->appendText(printText($percentage['tax_rate'] . '%' . __("的对象"), '', helper::amountPermillage($percentage['total_price'])  . ' (' . helper::amountPermillage($percentage['consumption_tax']) . ')', $width));
                } else {
                    $printer->appendText(printText('VAT (' . ($percentage['tax_rate'] . '%)'), '', helper::amountPermillage($percentage['total_price'])  . ' (' . helper::amountPermillage($percentage['consumption_tax']) . ')', $width));
                }
                $printer->lineFeed(1);
                $printer->setLineSpacing($isXinyeWifi ? 32 : 22);
                $printer->lineFeed();
                $printer->appendText($lineSpacing);
            }
        }
        //
        if ($order->refund_money > 0) {
            $printer->lineFeed();
            $printer->appendText(__("不包含退款金额") . " " . $this->getPriceAndUnit($order->refund_money ?: 0));
            $printer->lineFeed(1);
        }
        $printer->lineFeed();
        $printer->setLineSpacing($isXinyeWifi ? 24 : 22);
        // 支付方式
        $printer->appendText("------------------------------------------------\n");
        $printer->lineFeed();
        $printer->setCharacterSize(1, 1);
        $printer->setPrintModes(true, false, false);
        $payTypes = $order->payType()->select()->append(['pay_type'])->toArray();
        foreach ($payTypes as $key => $payType) {
            if ($payType['value'] == -1) {
                $payType['price'] = 0;
            } else if ($payType['value'] == OrderPayTypeEnum::CASH) {
                $payType['price'] = helper::bcsub($payType['price'], $order['change_due']);
            }
            $printer->appendText(printText($payType['pay_type']['text'], '', $this->getPriceAndUnit($payType['price']), $width));
            $printer->lineFeed();
            if ($key != count($payTypes) - 1) {
                $printer->lineFeed(2);
            }
        }
        //
        $printer->setPrintModes(false, false, false);
        $printer->setCharacterSize(1, 1);
        $printer->setLineSpacing($isXinyeWifi ? 48 : 20);
        $printer->lineFeed();
        //
        $printer->setLineSpacing($isXinyeWifi ? 48 : 20);
        $printer->appendText("------------------------------------------------\n");
        $printer->lineFeed();
        $printer->appendText(__("收银员") . ": " . $order->cashier?->real_name ?: '');
        $printer->lineFeed(2);
        $printer->appendText(__("订单号") . ": " . $order->order_no ?: '');
        $printer->lineFeed(2);
        $printer->appendText(__("打印次数") . ": " . $order->print_num);
        $printer->lineFeed(2);
        if ($company) {
            if ($isXinyeWifi) {
                $printer->setLineSpacing(80);
            }
            $printer->appendText(__("公司名称") . ": " . $company);
            $printer->setLineSpacing($isXinyeWifi ? 48 : 20);
            $printer->lineFeed(2);
        }
        if ($address) {
            if ($isXinyeWifi) {
                $printer->setLineSpacing(80);
            }
            $printer->appendText(__("地址") . ": " . $address);
            $printer->setLineSpacing($isXinyeWifi ? 48 : 20);
            $printer->lineFeed(2);
        }
        if ($taxNumber) {
            if ($isXinyeWifi) {
                $printer->setLineSpacing(80);
            }
            $printer->appendText(__("税号") . ": " . $taxNumber);
            $printer->setLineSpacing($isXinyeWifi ? 48 : 20);
            $printer->lineFeed(2);
        }
        if ($phone) {
            if ($isXinyeWifi) {
                $printer->setLineSpacing(80);
            }
            $printer->appendText(__("电话") . ": " . $phone);
            $printer->setLineSpacing($isXinyeWifi ? 48 : 20);
            $printer->lineFeed(2);
        }
        // 
        $printer->setLineSpacing(80);
        $printer->lineFeed(1);
        //
        $printer->appendText("*" . __("保管注意事项"));
        $printer->setLineSpacing($isXinyeWifi ? 48 : 20);
        $printer->lineFeed(2);
        $printer->setLineSpacing(80);
        $printer->appendText(__("如需保管时请将印刷页面朝内折叠"));
        // 技术支持方
        $printer->lineFeed(1);
        $printer->appendText("------------------------------------------------\n");
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        if ($this->lang == 'th') {
            $printer->appendText("ขอบคุณที่แวะมาหากัน!สนับสนุนโดย " . $brandName);
        } else {
            $printer->appendText(__("感谢您的光临！本店由") . " " . $brandName . " " . __("系统提供支持。"));
        }
        // Print and exit page mode
        $printer->lineFeed(4);
        $printer->printAndExitPageMode();
        $printer->lineFeed(4);
        $printer->cutPaper(true);
        //
        return $printer->orderData;
    }
}
