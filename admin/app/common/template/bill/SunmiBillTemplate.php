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
use app\common\enum\settings\PrinterTypeEnum;
use app\common\model\settings\PrinterTemplate;
use app\common\model\order\Order as OrderModel;
use app\common\model\settings\Setting as SettingModel;
use app\common\library\printer\party\SunmiCloudPrinter;

/**
 * 商米 - 结账单模版
 */
class SunmiBillTemplate extends BaseTemplate
{
    /**
     * 生成模版
     * @param OrderModel $order
     * @param string $printerType
     */
    public function create(OrderModel $order, $printerType, $isPrePrint = true)
    {
        $name = __('人');
        $products = $order->getPrinterProductsList();
        $percentageList = $order->getPercentageList();
        $template = $order->_template ?: PrinterTemplate::getTemplate($order['pay_time'] ? 2 : 3);
        // 店铺设置
        $settingStore = $this->setting[SettingEnum::STORE]['values'];
        $settingCloud = SettingModel::getCloudBasic();
        $company = $settingStore['company'] ?? '';
        $address = $settingStore['address'] ?? '';
        $phone = $settingStore['phone'] ?? '';
        $taxNumber = $settingStore['tax_number'] ?? '';
        $chainNumber = $settingCloud['shop']['chain_number'] ?? $settingStore['chain_number'] ?? '';
        // 品牌
        $brandName = $settingCloud['base']['brand_name'] ?? $settingStore['brand_name'] ?? '';
        // 日历
        $payTime = '';
        if ($order->pay_time) {
            $payTime = date('Y/m/d H:i:s', $order->pay_time);
            if ($this->defaultCalendar == '3') {
                $payTime = DateHelp::changeBuddhistCalendar($payTime);
            }
        }
        // 是否自己打印
        $isOneself = $printerType != PrinterTypeEnum::SUNMI_LAN && $printerType != PrinterTypeEnum::SUNMI_CLOUD;

        // 
        $orderName = $order['order_name'] ? ("-" . $order['order_name']) : '';

        /* *
        * 打印模版
        */
        $printer = new SunmiCloudPrinter(567);
        if ($template != 3) {
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText($order->_title ?: ($order->pay_time ? __("结账单") : __("预结账单")));
            $printer->lineFeed();
            $printer->setLineSpacing($isOneself ? 20 : 40);
        } else {
            $printer->setLineSpacing($isOneself ? 20 : 25);
        }
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        $printer->setPrintModes(true, true, false);
        $printer->setCharacterSize(2, 1);
        $printer->appendText("{$order->_shop_name}\n");
        $printer->lineFeed();
        $printer->setCharacterSize(1, 1);
        /* *
        * 模版1
        */
        if ($template == 1) {
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setLineSpacing($isOneself ? 30 : 70);
            $printer->setPrintModes(true, true, false);
            if ($order['table_no']) {
                if ($this->lang == 'my' || preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no'])) {
                    $printer->setLineSpacing(80);
                }
                $printer->appendText(__("桌号") . ": {$order['table_no']}{$orderName} ({$order['meal_num']}{$name})");
                $printer->lineFeed();
            } else if ($order['call_no']) {
                $printer->appendText(__("取单号") . ": {$order['call_no']}{$orderName}");
                $printer->lineFeed();
            }
            //
            $printer->restoreDefaultLineSpacing();
            $printer->setLineSpacing(40);
            $printer->setPrintModes(false, false, false);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setupColumns(
                [260, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->printInColumns(__("订单号"), $order->order_no);
            $printer->printInColumns(__("收银员"), $order->cashier?->real_name);
            if ($payTime) {
                $printer->printInColumns(__("时间"), $payTime);
                $printer->setLineSpacing(20);
            }
            $printer->lineFeed();
            $printer->setLineSpacing(40);
        }
        /* *
        * 模版 2
        */ else if ($template == 2) {
            $printer->setLineSpacing(50);
            $printer->setPrintModes(false, false, false);
            $printer->setCharacterSize(1, 1);
            $printer->appendText(__("非常感谢您今天的到来，我们期待您的再次光临"));
            $printer->lineFeed();
            //
            if ($payTime) {
                $printer->appendText($payTime . "\n");
                if ($isOneself) {
                    $printer->setLineSpacing(20);
                    $printer->lineFeed();
                }
            }
            //
            $printer->setLineSpacing($isOneself ? 20 : 40);
            $printer->setPrintModes(true, true, false);
            if ($order['table_no']) {
                $printer->setLineSpacing($isOneself ? 30 : 70);
                if ($this->lang == 'my' || preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no'])) {
                    if ($isOneself) {
                        $printer->lineFeed();
                    }
                    $printer->setLineSpacing(80);
                }
                $printer->appendText(__("桌号") . ": {$order['table_no']}{$orderName} ({$order['meal_num']}{$name})\n");
                $printer->setLineSpacing($isOneself ? 20 : 40);
                $printer->lineFeed();
            } else if ($order['call_no']) {
                $printer->appendText(__("取单号") . ": {$order['call_no']}{$orderName}\n");
                $printer->lineFeed();
            }
            //
            $printer->restoreDefaultLineSpacing();
            $printer->setLineSpacing(40);
            $printer->setPrintModes(false, false, false);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setupColumns(
                [260, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->printInColumns(__("订单号"), $order->order_no);
            if ($order->cashier?->real_name) {
                $printer->printInColumns(__("收银员"), $order->cashier?->real_name);
            }
            $printer->setLineSpacing(20);
            $printer->lineFeed();
            $printer->setLineSpacing(40);
        }
        /* *
        * 模版 3
        */ 
        else if ($template == 3) {
            //
            $printer->setCharacterSize(2, 1);
            $printer->setPrintModes(true, true, false);
            $printer->appendText($order->_title ?: ($payTime ? __("结账单") : __("预结账单")));
            $printer->setCharacterSize(1, 1);
            $printer->lineFeed();
            $printer->setLineSpacing(25);
            $printer->lineFeed();
            $printer->setLineSpacing(25);
            //
            $printer->setPrintModes(false, false, false);
            $printer->setLineSpacing(45);
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
            $printer->appendText("------------------------------------------------\n");
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setCharacterSize(2, 1);
            $printer->setPrintModes(true, true, false);
            $printer->setLineSpacing($isOneself ? 30 : 70);
            if ($order['table_no']) {
                if ($this->lang == 'my' || preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no'])) {
                    $printer->setLineSpacing(80);
                }
                $printer->appendText(__("桌号") . ": {$order['table_no']}{$orderName} ({$order['meal_num']}{$name})");
                $printer->lineFeed();
            } else if ($order['call_no']) {
                $printer->appendText(__("取单号") . ": {$order['call_no']}{$orderName}");
                $printer->lineFeed();
            }
            //
            $printer->setLineSpacing(45);
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
        $printer->restoreDefaultLineSpacing();
        $printer->appendText("\x1B\x33\x28");
        $printer->setPrintModes(false, false, false);
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
        $printer->setupColumns(
            [($this->lang == 'en' || $this->lang == 'th' || $this->lang == 'tr' || $this->lang == 'my') ? ($this->lang == 'my' ? 200 : 220) : 320, SunmiCloudPrinter::ALIGN_LEFT, 0],
            [($this->lang == 'en' || $this->lang == 'th' || $this->lang == 'tr' || $this->lang == 'my') ? ($this->lang == 'my' ? 240 : 230) : 120, SunmiCloudPrinter::ALIGN_CENTER, 0],
            [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
        );
        if ($template != 3) {
            $printer->printInColumns(__("商品"), __("单价") . '|' . __("数量"), __("小计"));
        }
        $printer->appendText("------------------------------------------------\n");
        // 自助餐
        $productNum = 0;
        $printer->setupColumns(
            [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
            [120, SunmiCloudPrinter::ALIGN_CENTER, 0],
            [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
        );
        // 赠品金额
        $freeMoney = 0;
        foreach ($products as $item) {
            if ($order->is_merge == 1) {
                $printer->appendText(__('桌号：') . $item['no'] . ($item['meal_num'] ? " ({$item['meal_num']}{$name})" : ''));
                $printer->lineFeed();
            }
            foreach ($item['buffetCustomerType'] ?? [] as $buffet) {
                $productNum += $buffet['num'];
                $buffetNameText = $buffet['buffet_name_text'];
                if ($buffet['customer_type_name_text'] ?? '') {
                    $buffetNameText .=  "\n" . '(' . $buffet['customer_type_name_text'] . ')';
                }
                $printer->printInColumns($buffetNameText, helper::amountPermillage($buffet['price']) . '*' . $buffet['num'], $this->getPriceAndUnit($buffet['total_price']));
                $printer->setLineSpacing(20);
                $printer->lineFeed();
                $printer->setLineSpacing(40);
            }
            foreach ($item['buffetDiscount'] ?? [] as $discount) {
                $productNum += $discount['num'];
                $printer->printInColumns($discount['name_text'], helper::amountPermillage($discount['price']) . '*' . $discount['num'], '-' . $this->getPriceAndUnit($discount['total_price']));
                $printer->setLineSpacing(20);
                $printer->lineFeed();
                $printer->setLineSpacing(40);
            }
            foreach ($item['delay'] ?? [] as $delay) {
                $productNum += $delay['num'];
                $printer->printInColumns($delay['name_text'], helper::amountPermillage($delay['price']) . '*' . $delay['num'], $this->getPriceAndUnit($delay['total_price']));
                $printer->setLineSpacing(20);
                $printer->lineFeed();
                $printer->setLineSpacing(40);
            }
            foreach ($item['products'] as $product) {
                $productNum += $product['total_num'];
                $gift = ($product['is_free'] ?? 0) > 0 ? ('(' . __("赠") . ') ') : '';
                $productAttr = (new OrderProduct)->getProductAttrAttr($product['product_attr'], true);
                $productName = $gift . $product['product_name'] . "\n" . $productAttr;
                $product_total_product_price = $product['total_product_price'];
                if ($gift) {
                    $freeMoney += $product['total_product_price'];
                    $product_total_product_price = 0;
                }
                $printer->printInColumns($productName, helper::amountPermillage($product['product_price']) . '*' . $product['total_num'], $this->getPriceAndUnit($product_total_product_price));
                $printer->setLineSpacing(20);
                $printer->lineFeed();
                $printer->setLineSpacing(40);
            }
        }
        //
        $printer->appendText("------------------------------------------------\n");
        $printer->setLineSpacing(45);
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
        // 商品金额 = 订单总价 - 赠品金额
        $total_product_price = $order['total_product_price'] - $freeMoney;
        if ($template == 3) {
            $printer->setupColumns(
                [200, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->printInColumns(__("商品数量") . ': ' . $productNum, __("商品金额") . ': ' . $this->getPriceAndUnit($total_product_price));
        } else {
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
            $printer->appendText(__("商品数量") . ': ' . $productNum);
            $printer->lineFeed();
            $printer->appendText(__("商品金额") . ': ' . $this->getPriceAndUnit($total_product_price));
            $printer->lineFeed();
        }
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
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
        if ($isOneself) {
            $printer->setLineSpacing(12);
            $printer->lineFeed();
            $printer->setLineSpacing(40);
        }
        $printer->setupColumns(
            [280, SunmiCloudPrinter::ALIGN_LEFT, 0],
            [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
        );
        $printer->appendText("\x1D\x21\x01\x01");
        $printer->setPrintModes(true, true, false);
        $printer->printInColumns(__("合计应收"), $this->getPriceAndUnit($order['actual_receive_price']));
        $printer->setPrintModes(false, false, false);
        $printer->setLineSpacing(20);
        // 商品已含税
        if ($order['consumption_tax_money'] > 0 && $order['consumption_tax_type'] == 1 && ($this->consumptionTax == 1 || $this->consumptionTax == 2)) {
            $printer->lineFeed();
            $printer->appendText("------------------------------------------------\n");
            $printer->lineFeed();
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
            $printer->appendText(__("合计 (其中VAT)"));
            $printer->lineFeed(2);
            foreach ($percentageList as $key => $percentage) {
                if ($this->lang == 'ja') {
                    $printer->printInColumns($percentage['tax_rate'] . '%' . __("的对象"),  helper::amountPermillage($percentage['total_price'])  . ' (' . helper::amountPermillage($percentage['consumption_tax']) . ')');
                } else {
                    $printer->printInColumns('VAT (' . ($percentage['tax_rate'] . '%)'),  helper::amountPermillage($percentage['total_price'])  . ' (' . helper::amountPermillage($percentage['consumption_tax']) . ')');
                }
                if ($key != count($percentageList) - 1) {
                    $printer->lineFeed();
                }
            }
        }
        if (!$isOneself) {
            $printer->lineFeed();
        }
        // 支付方式
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
        $printer->setLineSpacing(45);
        if ($order->pay_status['value'] == OrderPayStatusEnum::SUCCESS) {
            $payTypes = $order->payType()->select()->append(['pay_type'])->toArray();
            if (count($payTypes) > 0) {
                $printer->appendText("------------------------------------------------\n");
                $printer->setupColumns(
                    [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                    [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
                );
                foreach ($payTypes as $payType) {
                    $printer->printInColumns(__("支付方式"), $payType['pay_type']['text']);
                    $printer->printInColumns(__("实收金额"), $payType['value'] == -1 ? '0' : $this->getPriceAndUnit($payType['price']));
                    if (($order['change_due'] ?? 0) > 0 && $payType['pay_type']['value'] == OrderPayTypeEnum::CASH) {
                        $printer->printInColumns(__("找零"), helper::amountPermillage($order['change_due'] ?? 0) . '');
                    }
                }
            }
        }
        // 会员积分
        if ($order->user) {
            $printer->setupColumns(
                [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->appendText("------------------------------------------------\n");
            $pointnum = $order['is_free'] != 0 ? 0 : ($order->points_bonus ?: 0);
            $balance = $payTime ? ($order['surplus_balance'] ?? 0) : $order->surplusBalance();
            $printer->printInColumns(__("会员剩余余额"), $this->getPriceAndUnit($balance));
            $printer->printInColumns(__("本次积分"), $pointnum > 0 ? strval(helper::number2($pointnum)) : '0');
        }
        // 发票信息
        if ($order->_template == 3) {
            if ($order->invoiceInfo && ($order->invoiceInfo->company_name || $order->invoiceInfo->company_addr || $order->invoiceInfo->company_tax_number || $order->invoiceInfo->company_phone)) {
                $printer->setLineSpacing(10);
                $printer->lineFeed();
                $printer->appendText("------------------------------------------------\n");
                $printer->lineFeed($isOneself ? 1 : 2);
                $printer->appendText(__("发票信息"));
                $printer->setLineSpacing(24);
                $printer->lineFeed($isOneself ? 1 : 2);
                if ($order->invoiceInfo->company_name) {
                    $printer->setLineSpacing($this->isMy($order->invoiceInfo->company_name) && !$isOneself ? 50 : 40);
                    $printer->appendText(__("公司名称") . ": " . $order->invoiceInfo->company_name);
                    if (!$isOneself) {
                        $printer->setLineSpacing(20);
                        $printer->lineFeed(($order->invoiceInfo->company_addr || $order->invoiceInfo->company_tax_number || $order->invoiceInfo->company_phone) ? 2 : 1);
                    } else {
                        $printer->lineFeed(1);
                        $printer->setLineSpacing(40);
                    }
                }
                if ($order->invoiceInfo->company_addr) {
                    $printer->setLineSpacing($this->isMy($order->invoiceInfo->company_addr) && !$isOneself ? 50 : 40);
                    $printer->appendText(__("公司地址") . ": " . $order->invoiceInfo->company_addr);
                    if (!$isOneself) {
                        $printer->setLineSpacing(20);
                        $printer->lineFeed(($order->invoiceInfo->company_tax_number || $order->invoiceInfo->company_phone) ? 2 : 1);
                    } else {
                        $printer->lineFeed(1);
                        $printer->setLineSpacing(60);
                    }
                }
                if ($order->invoiceInfo->company_tax_number) {
                    $printer->setLineSpacing($this->isMy($order->invoiceInfo->company_tax_number) && !$isOneself  ? 40 : 40);
                    $printer->appendText(__("税号") . ": " . $order->invoiceInfo->company_tax_number);
                    if (!$isOneself) {
                        $printer->setLineSpacing(20);
                        $printer->lineFeed($order->invoiceInfo->company_phone ? 2 : 1);
                    } else {
                        $printer->lineFeed(1);
                        $printer->setLineSpacing(40);
                    }
                }
                if ($order->invoiceInfo->company_phone) {
                    $printer->setLineSpacing($this->isMy($order->invoiceInfo->company_phone) && !$isOneself  ? 40 : 40);
                    $printer->appendText(__("联系电话") . ": " . $order->invoiceInfo->company_phone);
                    if (!$isOneself) {
                        $printer->setLineSpacing(20);
                        $printer->lineFeed($order->invoiceInfo->company_phone ? 2 : 1);
                    } else {
                        $printer->lineFeed(1);
                        $printer->setLineSpacing(40);
                    }
                }
            } else {
                $printer->setLineSpacing(10);
                $printer->lineFeed();
            }
        }
        // 技术支持方
        $printer->setLineSpacing(45);
        $printer->appendText("------------------------------------------------\n");
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        if ($this->lang == 'th') {
            $printer->appendText("ขอบคุณที่แวะมาหากัน!สนับสนุนโดย " . $brandName);
        } else {
            $printer->appendText(__("感谢您的光临！本店由") . " " . $brandName . " " . __("系统提供支持。"));
        }
        // Print and exit page mode
        $printer->lineFeed();
        $printer->printAndExitPageMode();
        $printer->lineFeed(3);
        $printer->cutPaper(false);
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
        //     $printer->appendText(chr(27) . chr(112) . chr(0) . chr(25) . chr(250));
        // }
        //
        return $printer->orderData;
    }
}
