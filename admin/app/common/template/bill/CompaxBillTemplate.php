<?php

namespace app\common\template\bill;

use help\DateHelp;
use app\common\library\helper;
use app\common\model\user\BalanceLog;
use app\common\template\BaseTemplate;
use app\common\model\order\OrderProduct;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\model\settings\PrinterTemplate;
use app\common\model\order\Order as OrderModel;
use app\common\model\settings\Setting as SettingModel;
use app\common\library\printer\party\SunmiCloudPrinter;

/**
 * Compax 结账单模版
 */
class CompaxBillTemplate extends BaseTemplate
{
    /**
     * 生成模版
     * @param OrderModel $order
     */
    public function create(OrderModel $order, $printerType = '', $isPrePrint = true)
    {
        $name = __('人');
        $width = 48;
        $leftWidth = 28;
        $template = $order->_template ?: PrinterTemplate::getTemplate($order['pay_time'] ? 2 : 3);
        $products = $order->getPrinterProductsList();
        $percentageList = $order->getPercentageList();
        //
        if ($this->currencyUnit == "￥") {
            $this->currencyUnit = "\xC2\xA5";
        }
        // 店铺设置
        $settingStore = $this->setting[SettingEnum::STORE]['values'];
        $settingCloud = SettingModel::getCloudBasic();
        $company = $settingStore['company'] ?? '';
        $address = $settingStore['address'] ?? '';
        $phone = $settingStore['phone'] ?? '';
        $taxNumber = $settingStore['tax_number'] ?? '';
        $chainNumber = $settingCloud['shop']['chain_number'] ?? $settingStore['chain_number'] ?? '';
        $brandName = $settingCloud['base']['brand_name'] ?? $settingStore['brand_name'] ?? '';
        // 日历
        $payTime = '';
        if ($order->pay_time) {
            $payTime = date('Y/m/d H:i:s', $order->pay_time);
            if ($this->defaultCalendar == '3') {
                $payTime = DateHelp::changeBuddhistCalendar($payTime);
            }
        }
        
        // 
        $orderName = $order['order_name'] ? ("-" . $order['order_name']) : '';

        //
        $printer = new SunmiCloudPrinter(567);
        if ($template != 3) {
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText($order->_title ?: ($order->pay_time ? __("结账单") : __("预结账单")));
            $printer->lineFeed();
        }
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        $printer->setPrintModes(false, false, false);
        $printer->setCharacterSize(2, 1);
        $printer->setLineSpacing(30);
        $printer->appendText("{$order->_shop_name}\n");
        $printer->lineFeed();
        // 欢迎语
        $printer->setLineSpacing(45);
        $printer->setPrintModes(false, false, false);
        $printer->setCharacterSize(1, 1);
        /* *
        * 模版1
        */
        if ($template == 1) {
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setCharacterSize(1, 1);
            $printer->appendText("\x1D\x21\x01");
            $printer->setPrintModes(true, true, false);
            if ($order['table_no']) {
                $printer->appendText(__("桌号") . ": {$order['table_no']}{$orderName} ({$order['meal_num']}{$name})\n");
            } else if ($order['call_no']) {
                $printer->appendText(__("取单号") . ": {$order['call_no']}{$orderName}\n");
            }
            //
            $printer->setCharacterSize(1, 1);
            $printer->setPrintModes(false, false, false);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("订单号"), '', $order->order_no, $width));
            $printer->lineFeed();
            $printer->appendText(printText(__("收银员"), '', $order->cashier?->real_name ?: '', $width));
            $printer->lineFeed();
            if ($payTime) {
                $printer->appendText(printText(__("时间"), '', $payTime . "\n", $width));
                $printer->lineFeed();
            }
            $printer->setLineSpacing(20);
            $printer->lineFeed();
            $printer->setLineSpacing(45);
        }
        /* *
        * 模版 2
        */ else if ($template == 2) {
            $printer->appendText(__("非常感谢您今天的到来，我们期待您的再次光临") . "\n");
            // 付款时间
            if ($order->pay_time) {
                $printer->appendText(date('Y/m/d H:i:s', $order->pay_time) . "\n");
            }
            // 桌号
            $printer->setLineSpacing(5);
            $printer->lineFeed();
            $printer->setLineSpacing(35);
            $printer->setPrintModes(true, true, false);
            if ($order['table_no']) {
                $printer->appendText(__("桌号") . ": {$order['table_no']}{$orderName} ({$order['meal_num']}{$name})\n");
            } else if ($order['call_no']) {
                $printer->appendText(__("取单号") . ": {$order['call_no']}{$orderName}\n");
            }
            $printer->lineFeed();
            //
            $printer->setPrintModes(false, false, false);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("订单号"), '', $order->order_no, $width));
            $printer->lineFeed();
            $printer->appendText(printText(__("收银员"), '', $order->cashier?->real_name ?: '', $width));
            $printer->lineFeed();
            $printer->lineFeed();
        }
        /* *
        * 模版 3
        */ 
        else if ($template == 3) {
            //
            $printer->setCharacterSize(2, 2);
            $printer->setPrintModes(true, true, false);
            $printer->appendText($order->_title ?: ($order->pay_time ? __("结账单") : __("预结账单")));
            $printer->setPrintModes(false, false, false);
            $printer->setCharacterSize(1, 1);
            $printer->lineFeed();
            //
            if ($company) {
                $printer->appendText(__("公司名称：") . $company);
                $printer->lineFeed();
            }
            if ($chainNumber) {
                $printer->appendText(__("连锁店编号：") . $chainNumber);
                $printer->lineFeed();
            }
            if ($address) {
                $printer->appendText(__("商家地址：") . $address);
                $printer->lineFeed();
            }
            if ($phone) {
                $printer->appendText(__("电话：") . $phone);
                $printer->lineFeed();
            }
            if ($taxNumber) {
                $printer->appendText(__("税号：") . $taxNumber);
                $printer->lineFeed();
            }
            //
            $printer->appendText("------------------------------------------------\n");
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setCharacterSize(2, 2);
            $printer->setPrintModes(true, true, false);
            if ($order['table_no']) {
                $printer->appendText(__("桌号") . ": {$order['table_no']}{$orderName} ({$order['meal_num']}{$name})");
                $printer->lineFeed();
            } else if ($order['call_no']) {
                $printer->appendText(__("取单号") . ": {$order['call_no']}{$orderName}");
                $printer->lineFeed();
            }
            $printer->setCharacterSize(1, 1);
            $printer->setPrintModes(false, false, false);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(__("收银员") . ': ' . $order->cashier?->real_name ?: '');
            $printer->lineFeed();
            if ($payTime) {
                $printer->appendText(__("时间") . ': ' . $payTime);
                $printer->lineFeed();
            }
            $printer->appendText(__("订单号") . ': ' . $order->order_no);
            $printer->lineFeed();
        }
        //
        $leftWidth = 25;
        $centerWidth = 16;
        $rightWidth = 16;
        if ($template == 3) {
            $printer->appendText("------------------------------------------------");
        } else {
            $hLeftWidth = $leftWidth - ($this->lang == 'en' ? 10 : ($this->lang == 'th' ? 8 : 0));
            $printer->setPrintModes(true, false, false);
            $printer->appendText(printText(__("商品"), __("单价") . '|' . __("数量"), __("小计"), $width, $hLeftWidth));
            $printer->setPrintModes(false, false, false);
            $printer->appendText("\n------------------------------------------------");
        }
        // 自助餐
        $productNum  = 0;
        $currencyWidth = $this->currencyUnit == '¥' ? $width - 1 : $width;
        $printer->setLineSpacing(35);
        // 赠品金额
        $freeMoney = 0;
        foreach ($products as $item) {
            if ($order->is_merge == 1) {
                $printer->setPrintModes(true, false, false);
                $printer->appendText(__('桌号：') . $item['no'] . ($item['meal_num'] ? " ({$item['meal_num']}{$name})" : ''));
                $printer->setPrintModes(false, false, false);
                $printer->lineFeed();
                $printer->setLineSpacing(35);
            }
            foreach ($item['buffetCustomerType'] ?? [] as $buffet) {
                $productNum += $buffet['num'];
                $buffetNameText = $buffet['buffet_name_text'];
                if ($buffet['customer_type_name_text'] ?? '') {
                    $buffetNameText .=  "\n" . '(' . $buffet['customer_type_name_text'] . ')';
                }
                $printer->appendText(printText($buffetNameText, helper::amountPermillage($buffet['price']) . '*' . $buffet['num'], $this->getPriceAndUnit($buffet['total_price']), $currencyWidth, $leftWidth, $centerWidth, $rightWidth));
                $printer->lineFeed();
                $printer->setLineSpacing(10);
                $printer->lineFeed();
                $printer->setLineSpacing(35);
            }
            foreach ($item['buffetDiscount'] ?? [] as $discount) {
                $productNum += $discount['num'];
                $printer->appendText(printText($discount['name_text'], helper::amountPermillage($discount['price']) . '*' . $discount['num'], '-' . $this->getPriceAndUnit($discount['total_price']), $currencyWidth, $leftWidth, $centerWidth, $rightWidth));
                $printer->lineFeed();
                $printer->setLineSpacing(10);
                $printer->lineFeed();
                $printer->setLineSpacing(35);
            }
            foreach ($item['delay'] ?? [] as $delay) {
                $productNum += $delay['num'];
                $printer->appendText(printText($delay['name_text'], helper::amountPermillage($delay['price']) . '*' . $delay['num'], $this->getPriceAndUnit($delay['total_price']), $currencyWidth, $leftWidth, $centerWidth, $rightWidth));
                $printer->lineFeed();
                $printer->setLineSpacing(10);
                $printer->lineFeed();
                $printer->setLineSpacing(35);
            }
            foreach ($item['products'] as $key => $product) {
                $productNum += $product['total_num'];
                if ($key == count($products) - 1) {
                    $product['product_name'] = str_replace("ท์)", 'ท์ )', $product['product_name']);
                    $product['product_attr'] = str_replace("ท์)", 'ท์ )', $product['product_attr']);
                }
                $gift = ($product['is_free'] ?? 0) > 0 ? ('(' . __("赠") . ') ') : '';
                $productAttr = (new OrderProduct)->getProductAttrAttr($product['product_attr'], true);
                $productName = $gift . $product['product_name'] . "\n" . $productAttr;
                $product_total_product_price = $product['total_product_price'];
                if ($gift) {
                    $freeMoney += $product['total_product_price'];
                    $product_total_product_price = 0;
                }
                $printer->appendText(printText($productName, helper::amountPermillage($product['product_price']) . '*' . $product['total_num'], $this->getPriceAndUnit($product_total_product_price), $currencyWidth, $leftWidth, $centerWidth, $rightWidth));
                $printer->lineFeed();
                $printer->setLineSpacing(10);
                $printer->lineFeed();
                $printer->setLineSpacing(35);
            }
        }
        //
        $printer->appendText("------------------------------------------------");
        $printer->setLineSpacing(45);
        $printer->lineFeed();
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
        // 商品金额 = 订单总价 - 赠品金额
        $total_product_price = $order['total_product_price'] - $freeMoney;
        if ($template == 3) {
            $printer->appendText(printText(__("商品数量") . ': ' . $productNum, '', __("商品金额") . ': ' . $this->getPriceAndUnit($total_product_price), $currencyWidth));
            $printer->lineFeed();
        } else {
            $printer->appendText(__("商品数量") . ': ' . $productNum);
            $printer->lineFeed();
            $printer->appendText(__("商品金额") . ': ' . $this->getPriceAndUnit($total_product_price));
            $printer->lineFeed();
        }
        if ($order['service_money'] > 0) {
            $printer->appendText(__("服务费") . ': ' . $this->getPriceAndUnit($order['service_money']));
            $printer->lineFeed();
        }
        // 商品未含税
        if ($order['consumption_tax_money'] > 0 && $order['consumption_tax_type'] == 2 && ($this->consumptionTax == 1 || $this->consumptionTax == 3)) {
            foreach ($percentageList as $percentage) {
                if ($this->lang == 'ja') {
                    $printer->appendText($percentage['tax_rate'] . '%' . __("的对象消费税") . ': ' . $this->getPriceAndUnit($percentage['consumption_tax']));
                } else {
                    $printer->appendText('VAT (' . ($percentage['tax_rate'] . '%)') . ': ' . $this->getPriceAndUnit($percentage['consumption_tax']));
                }
                $printer->lineFeed();
            }
        }
        if ($order['is_free'] == 0 && $order['discount_money'] != 0) {
            $product_discount_money = 0;
            foreach ($products as $p) {
                foreach ($p['products'] as $p) {
                    if ($p['is_free'] > 0) {
                        $product_discount_money += $p['product_discount_money'];
                    }
                }
            }
            $order['discount_money'] = $order['discount_money'] - $product_discount_money;
            if ($order['discount_money'] != 0) {
                $ratio = $template == 3 ? ' (' . ($order['total_product_price'] <= 0 ? 0 : round($order['discount_money'] / $order['total_product_price'], 4)) * 100 . '% OFF)' : '';
                $printer->appendText(__("优惠折扣") . ': ' . $this->getPriceAndUnit($order['discount_money']) . $ratio);
                $printer->lineFeed();
            }
        }
        if ($order['user_discount_money'] != 0) {
            $printer->appendText(__("会员优惠") . ': ' . $this->getPriceAndUnit($order['user_discount_money']));
            $printer->lineFeed();
            // 
            $gradeEquity = $template == 3 ? ($order['user']['grade']['equity'] ?? 100) : 100;
            $cardDiscount = $template == 3 ? ($order['user']['card']['discount'] ?? 100) : 100;
            if ($gradeEquity && $gradeEquity != 100 && $gradeEquity > 0) {
                $printer->appendText(__("会员折扣") . ': ' . (($this->lang == 'zh' || $this->lang == 'zhtw') ? (floatval($gradeEquity / 10) . '折') : (floatval($gradeEquity) . '%')));
                $printer->lineFeed();
            }
            if ($cardDiscount && $cardDiscount != 100 && $cardDiscount > 0) {
                $printer->appendText(__("会员卡折扣") . ': ' . (($this->lang == 'zh' || $this->lang == 'zhtw') ? (floatval($cardDiscount / 10) . '折') : (floatval($cardDiscount) . '%')));
                $printer->lineFeed();
            }
        }
        if ($order['checkout_diff_money'] > 0) {
            $printer->appendText(__("手动抹零") . ': ' . $this->getPriceAndUnit($order['checkout_diff_money']));
            $printer->lineFeed();
        }
        if ($order['refund_money'] > 0) {
            $printer->appendText(__("退款金额") . ': ' . $this->getPriceAndUnit($order['refund_money']));
            $printer->lineFeed();
        }
        if ($order['pay_fee_money'] > 0) {
            $printer->appendText(__("支付手续费") . ': ' . $this->getPriceAndUnit($order['pay_fee_money']));
            $printer->lineFeed();
        }
        if ($order['free_pay_price'] > 0) {
            $printer->appendText(__("免单金额") . ': ' . $this->getPriceAndUnit($order['free_pay_price']));
            $printer->lineFeed();
        }
        //
        if ($template == 3) {
            $printer->appendText("------------------------------------------------\n");
        }
        $printer->appendText("\x1D\x21\x01\x01");
        $printer->setPrintModes(true, true, false);
        $printer->appendText(printText(__("合计应收"), '', $this->getPriceAndUnit($order['actual_receive_price']), $currencyWidth, 34));
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
        $printer->setPrintModes(false, false, false);
        $printer->appendText("\x1D\x21\x00\x00");
        // 商品已含税
        if ($order['consumption_tax_money'] > 0  && $order['consumption_tax_type'] == 1 && ($this->consumptionTax == 1 || $this->consumptionTax == 2)) {
            $printer->appendText("------------------------------------------------\n");
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
            $printer->appendText(__("合计 (其中VAT)"));
            $printer->lineFeed(1);
            foreach ($percentageList as $percentage) {
                if ($this->lang == 'ja') {
                    $printer->appendText(printText($percentage['tax_rate'] . '%' . __("的对象"), '', helper::amountPermillage($percentage['total_price'])  . ' (' . helper::amountPermillage($percentage['consumption_tax']) . ')', $currencyWidth, 24));
                } else {
                    $printer->appendText(printText('VAT (' . ($percentage['tax_rate'] . '%)'), '', helper::amountPermillage($percentage['total_price'])  . ' (' . helper::amountPermillage($percentage['consumption_tax']) . ')', $currencyWidth, 24));
                }
                $printer->lineFeed(1);
            }
        }
        //
        if ($order->pay_status['value'] == OrderPayStatusEnum::SUCCESS) {
            $payTypes = $order->payType()->select()->append(['pay_type'])->toArray();
            if (count($payTypes) > 0) {
                $printer->appendText("------------------------------------------------\n");
                foreach ($payTypes as $payType) {
                    $printer->appendText(printText(__("支付方式"), '', $payType['pay_type']['text'], $width, 20, 0, 28) . "\n");
                    $printer->appendText(printText(__("实收金额"), '', $payType['value'] == -1 ? '0' : $this->getPriceAndUnit($payType['price']), $currencyWidth, 34));
                    if (($order['change_due'] ?? 0) > 0 && $payType['pay_type']['value'] == OrderPayTypeEnum::CASH) {
                        $printer->appendText(printText(__("找零"), '', helper::amountPermillage($order['change_due'] ?? 0), $width, 34));
                    }
                }
            }
        }
        //
        if ($order->user) {
            $printer->appendText("\n------------------------------------------------\n");
            $pointnum = $order['is_free'] != 0 ? 0 : ($order->points_bonus ?: 0);
            $balance = $payTime ? ($order['surplus_balance'] ?? 0) : $order->surplusBalance();
            $printer->appendText(printText(__("会员剩余余额"), '',  $this->getPriceAndUnit($balance), $currencyWidth, 34) . "\n");
            $printer->appendText(printText(__("本次积分"), '', $pointnum > 0 ? strval(helper::number2($pointnum)) : '0', $width, 34));
        }
        // 发票信息
        if ($order->_template == 3 && $order->invoiceInfo && ($order->invoiceInfo->company_name || $order->invoiceInfo->company_addr || $order->invoiceInfo->company_tax_number || $order->invoiceInfo->company_phone)) {
            $printer->appendText("------------------------------------------------\n");
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
        // 技术支持方
        $printer->appendText("\n------------------------------------------------\n");
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        if ($this->lang == 'th') {
            $printer->appendText("ขอบคุณที่แวะมาหากัน!สนับสนุนโดย " . $brandName);
        } else {
            $printer->appendText(__("感谢您的光临！本店由") . " " . $brandName . " " . __("系统提供支持。"));
        }
        //
        // Print and exit page mode
        $printer->restoreDefaultLineSpacing();
        $printer->lineFeed();
        $printer->printAndExitPageMode();
        $printer->lineFeed(6);
        $printer->cutPaper(true);
        // 打开钱箱
        // $openMoneybox = false;
        // if (!$isPrePrint && $order->pay_time) {
        //     foreach ($order->payType as $pay) {
        //         if ($pay['value'] == OrderPayTypeEnum::CASH) {
        //             $openMoneybox = true;
        //             break;
        //         }
        //     }
        // }
        // if ($openMoneybox) {
        //     $printer->appendText("\x10\x14\x01\x00\x01");
        // }
        //
        return $printer->orderData;
    }
}
