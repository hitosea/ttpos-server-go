<?php

namespace app\common\template\invoice;

use help\DateHelp;
use app\common\library\helper;
use app\common\template\BaseTemplate;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\model\settings\PrinterTemplate;
use app\common\model\order\Order as OrderModel;
use app\common\template\bill\CompaxBillTemplate;
use app\common\model\settings\Setting as SettingModel;
use app\common\library\printer\party\SunmiCloudPrinter;

/**
 * Compax 发票模版
 */
class CompaxInvoiceTemplate extends BaseTemplate
{
    /**
     * 生成模版
     */
    public function create(OrderModel $order, $printType, $isCashierPrinter)
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
            return (new CompaxBillTemplate($this->setting, $this->allSourceProductList, $this->isSunmi))->create($order, $printType);
        }

        /* *
        * 模版 1
        */
        if ($this->currencyUnit == "￥") {
            $this->currencyUnit = "\xC2\xA5";
        }
        //
        $width = 48;
        $leftWidth = 28;
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
        $currencyWidth = $width;
        if ($this->lang == 'th') {
            $currencyWidth = $currencyWidth - 1;
        } else if ($this->currencyUnit == '฿' || $this->currencyUnit == '¥') {
            $currencyWidth = $currencyWidth - 1;
        }
        //
        $leftWidth = 28;
        $lineSpacing = "\x1B\x33\x32";
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        $printer->setCharacterSize(2, 1);
        $printer->appendText("{$shopName}");
        $printer->setCharacterSize(1, 1);
        $printer->setLineSpacing($isCashierPrinter ? 60 : 80);
        $printer->lineFeed();
        $printer->appendText("\x1B\x33\x28");
        $printer->appendText(__("非常感谢您今天的到来，我们期待您的再次光临"));
        $printer->appendText($lineSpacing);
        $printer->lineFeed(1);
        $printer->setCharacterSize(2, 2);
        $printer->appendText(__("发票"));
        $printer->setCharacterSize(1, 1);
        $printer->lineFeed(1);
        $printer->appendText($payTime);
        $printer->setLineSpacing(60);
        $printer->lineFeed(1);
        //
        if ($this->lang == 'ja') {
            $printer->setLineSpacing(34);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
            $printer->appendText(__("先生/小姐"));
            $printer->lineFeed();
        }
        $printer->setLineSpacing(30);
        $printer->appendText("------------------------------------------------\n");
        $printer->setCharacterSize(2, 2);
        $printer->appendText(printText(__("合计"), '', $this->getPriceAndUnit($order->actual_receive_price), $currencyWidth - 24));
        $printer->setCharacterSize(1, 1);
        $printer->lineFeed(2);
        $printer->appendText(printText('(' . __("其中服务费"), '', $this->getPriceAndUnit($order->service_money ?: 0) . ')', $currencyWidth, $leftWidth));
        $printer->lineFeed(2);
        if ($consumptionTax != 4) {
            $printer->appendText(printText('(' . __("其中VAT"), '', $this->getPriceAndUnit($order->consumption_tax_money ?: 0) . ')', $currencyWidth, $leftWidth));
            $printer->lineFeed(2);
        }
        $printer->appendText("------------------------------------------------\n");
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
        //
        if ($consumptionTax != 4) {
            $printer->setLineSpacing(12);
            $printer->lineFeed();
            $printer->appendText($lineSpacing);
            $printer->appendText(__("仅作为餐饮费收取以上金额"));
            $printer->lineFeed();
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
            $printer->appendText(__("合计 (其中VAT)"));
            $printer->lineFeed();
            $printer->appendText($lineSpacing);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            foreach ($percentageList as $percentage) {
                if ($this->lang == 'ja') {
                    $printer->appendText(printText($percentage['tax_rate'] . '%' . __("的对象"), '', helper::amountPermillage($percentage['total_price'])  . ' (' . helper::amountPermillage($percentage['consumption_tax']) . ')', $width));
                } else {
                    $printer->appendText(printText('VAT (' . ($percentage['tax_rate'] . '%)'), '', helper::amountPermillage($percentage['total_price'])  . ' (' . helper::amountPermillage($percentage['consumption_tax']) . ')', $width));
                }
                $printer->lineFeed(1);
                $printer->appendText($lineSpacing);
            }
        } else if ($order->refund_money > 0) {
            $printer->setLineSpacing(12);
            $printer->lineFeed();
            $printer->appendText($lineSpacing);
        }
        //
        if ($order->refund_money > 0) {
            $printer->appendText(__("不包含退款金额") . " " . $this->getPriceAndUnit($order->refund_money ?: 0));
            $printer->lineFeed();
        }
        // 支付方式
        $printer->setLineSpacing(40);
        $printer->setPrintModes(true, false, false);
        $printer->appendText("------------------------------------------------");
        $printer->setLineSpacing(40);
        $printer->lineFeed();
        $printer->setPrintModes(true, false, false);
        $payTypes = $order->payType()->select()->append(['pay_type'])->toArray();
        foreach ($payTypes as $key => $payType) {
            if ($payType['value'] == -1) {
                $payType['price'] = 0;
            } else if ($payType['value'] == OrderPayTypeEnum::CASH) {
                $payType['price'] = helper::bcsub($payType['price'], $order['change_due']);
            }
            $printer->setLineSpacing(50);
            $printer->appendText(printText($payType['pay_type']['text'], '', $this->getPriceAndUnit($payType['price']), $currencyWidth));
            if ($key != count($payTypes) - 1) {
                $printer->lineFeed();
            }
        }
        //
        $printer->setPrintModes(false, false, false);
        $printer->setLineSpacing(40);
        $printer->lineFeed();
        $printer->setLineSpacing(10);
        // 发票信息
        if ($order->invoiceInfo && ($order->invoiceInfo->company_name || $order->invoiceInfo->company_addr || $order->invoiceInfo->company_tax_number || $order->invoiceInfo->company_phone)) {
            $printer->appendText("------------------------------------------------\n");
            $printer->lineFeed(1);
            $printer->setLineSpacing(40);
            $printer->lineFeed();
            $printer->appendText($lineSpacing);
            $printer->appendText(__("发票信息"));
            $printer->lineFeed(1);
            if ($order->invoiceInfo->company_name) {
                $printer->appendText(__("公司名称") . ": " . $order->invoiceInfo->company_name);
                $printer->lineFeed(1);
            }
            if ($order->invoiceInfo->company_addr) {
                $printer->appendText(__("公司地址") . ": " . $order->invoiceInfo->company_addr);
                $printer->lineFeed(1);
            }
            if ($order->invoiceInfo->company_tax_number) {
                $printer->appendText(__("税号") . ": " . $order->invoiceInfo->company_tax_number);
                $printer->lineFeed(1);
            }
            if ($order->invoiceInfo->company_phone) {
                $printer->appendText(__("联系电话") . ": " . $order->invoiceInfo->company_phone);
                $printer->lineFeed(1);
            }
        }
        // 
        $printer->appendText("------------------------------------------------");
        $printer->setLineSpacing(40);
        $printer->lineFeed();
        $printer->appendText($lineSpacing);
        $printer->appendText(__("收银员") . ": " . $order->cashier?->real_name ?: '');
        $printer->lineFeed();
        $printer->appendText(__("订单号") . ": " . $order->order_no ?: '');
        $printer->lineFeed();
        $printer->appendText(__("打印次数") . ": " . $order->print_num);
        $printer->lineFeed();
        if ($company) {
            $printer->appendText("\x1B\x33\x28");
            $printer->appendText(__("公司名称") . ": " . $company);
            $printer->appendText($lineSpacing);
            $printer->lineFeed();
        }
        if ($address) {
            $printer->appendText("\x1B\x33\x28");
            $printer->appendText(__("地址") . ": " . $address);
            $printer->appendText($lineSpacing);
            $printer->lineFeed();
        }
        if ($taxNumber) {
            $printer->appendText(__("税号") . ": " . $taxNumber);
            $printer->lineFeed();
        }
        if ($phone) {
            $printer->appendText(__("电话") . ": " . $phone);
            $printer->lineFeed();
        }
        $printer->appendText("\x1B\x33\x28");
        $printer->lineFeed();
        $printer->appendText($lineSpacing);
        $printer->appendText("*" . __("保管注意事项"));
        $printer->lineFeed();
        $printer->appendText(__("如需保管时请将印刷页面朝内折叠"));
        // 技术支持方
        $printer->lineFeed();
        $printer->appendText("------------------------------------------------\n");
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        if ($this->lang == 'th') {
            $printer->appendText("ขอบคุณที่แวะมาหากัน!สนับสนุนโดย " . $brandName);
        } else {
            $printer->appendText(__("感谢您的光临！本店由") . " " . $brandName . " " . __("系统提供支持。"));
        }
        // Print and exit page mode
        $printer->printAndExitPageMode();
        $printer->lineFeed(4);
        $printer->cutPaper(true);
        //
        return $printer->orderData;
    }
}
