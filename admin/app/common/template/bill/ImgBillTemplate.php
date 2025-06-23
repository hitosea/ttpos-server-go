<?php

namespace app\common\template\bill;

use help\ImgHelp;
use help\DateHelp;
use base\imgs\ImgFont;
use app\common\library\helper;
use app\common\model_old\store\PayType;
use app\common\template\BaseTemplate;
use app\admin\model\supplier\Supplier;
use app\common\model_old\order\OrderProduct;
use app\common\enum\settings\SettingEnum;
use app\common\model_old\payment\PaymentOrder;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\model_old\settings\PrinterTemplate;
use app\common\model_old\order\Order as OrderModel;
use app\common\model_old\settings\Setting as SettingModel;

/**
 * 图片打印 - 结账单模版
 */
class ImgBillTemplate extends BaseTemplate
{
    /**
     * 生成模版
     * @param OrderModel $order
     * @param string $printerType
     */
    public function create($order, $paramData = '', $isPrePrint = true)
    {
        $name = __('人');
        $products = $order->getPrinterProductsList();
        $percentageList = $order->getPercentageList();
        $template = $order->_template ?: $template = (new PrinterTemplate([], request()->newAppId))->where('id', '=', $order['pay_time'] ? 2 : 3)->value('template');
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

        // 
        $orderName = $order['order_name'] ? ("-" . $order['order_name']) : '';

        /* *
        * 打印模版
        */
        $printer = new ImgFont(568, 45);
        $printer->setTextLineHeight(50);
        $printer->setImagePadding($this->lang == 'my' ? 3 : 0);
        if ($template != 3) {
            $printer->setAlignment(ImgFont::ALIGN_LEFT);
            $printer->appendText($order->_title ?: ($order->pay_time ? __("结账单") : __("预结账单")));
            $printer->lineFeed(1, 60);
        }
        $printer->setAlignment(ImgFont::ALIGN_CENTER);
        $printer->setFontWeight(2);
        $printer->setFontSize(26);
        $printer->appendText("{$order->_shop_name}");
        $printer->setFontSize(20);
        $printer->setFontWeight(1);
        $printer->lineFeed(1, 60);
        
        /* *
        * 模版1
        */
        if ($template == 1) {
            $printer->lineFeed(1, 10);
            $printer->recoverDefaultTextLineHeight();
            $printer->setAlignment(ImgFont::ALIGN_LEFT);
            if ($order['table_no']) {
                $printer->setFontWeight(2);
                $printer->setFontSize(28);
                $printer->setTextLineHeight(preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no']) ? 68 : 50);
                $printer->appendText(__("桌号") . "： {$order['table_no']}{$orderName} ({$order['meal_num']}{$name})");
                $printer->recoverDefaultTextLineHeight();
                $printer->setFontSize(20);
                $printer->lineFeed();
            } else if ($order['call_no']) {
                $printer->setFontWeight(2);
                $printer->setFontSize(28);
                $printer->appendText(__("取单号") . "： {$order['call_no']}{$orderName}");
                $printer->setFontSize(20);
                $printer->lineFeed();
            }
            //
            $printer->setFontWeight(1);
            $printer->printInColumns(
                [__("订单号"), 300, ImgFont::ALIGN_LEFT],
                [$order->order_no, 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("收银员"), 300, ImgFont::ALIGN_LEFT],
                [$order->cashier?->real_name ?: '', 0, ImgFont::ALIGN_RIGHT],
            );
            if ($payTime) {
                $printer->printInColumns(
                    [__("时间"), 300, ImgFont::ALIGN_LEFT],
                    [$payTime, 0, ImgFont::ALIGN_RIGHT],
                );
            }
            $printer->lineFeed();
        }
        /* *
        * 模版 2
        */ 
        else if ($template == 2) {
            $printer->appendText(__("非常感谢您今天的到来，我们期待您的再次光临"));
            $printer->lineFeed(1, 60);
            //
            if ($payTime) {
                $printer->appendText($payTime);
                $printer->lineFeed();
            }
            //
            $printer->setFontWeight(2);
            $printer->setFontSize(28);
            if ($order['table_no']) {
                $printer->setTextLineHeight(preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no']) ? 68 : 50);
                $printer->appendText(__("桌号") . "： {$order['table_no']}{$orderName} ({$order['meal_num']}{$name})");
                $printer->recoverDefaultTextLineHeight();
            } else if ($order['call_no']) {
                $printer->appendText(__("取单号") . "： {$order['call_no']}{$orderName}");
            }
            $printer->setFontSize(20);
            $printer->lineFeed(1, 60);
            $printer->recoverDefaultTextLineHeight();
            $printer->setFontWeight(1);
            $printer->printInColumns(
                [__("订单号"), 300, ImgFont::ALIGN_LEFT, 2],
                [$order->order_no, 0, ImgFont::ALIGN_RIGHT, 2],
            );
            if ($order->cashier?->real_name) {
                $printer->printInColumns(
                    [__("收银员"), 300, ImgFont::ALIGN_LEFT, 2],
                    [$order->cashier?->real_name, 0, ImgFont::ALIGN_RIGHT, 2],
                );
            }
            $printer->lineFeed();
        }
        /* *
        * 模版 3
        */ 
        else if ($template == 3) {
            // 打印logo
            if ($this->setting) {
                try {
                    $storeSetting = $this->setting[SettingEnum::STORE]['values'] ?? [];
                    $printer->setTextLineHeight(25);
                    $printer->setAlignment(ImgFont::ALIGN_CENTER);
                    $whiteBackgroundWithBlackTextLogoPath = Supplier::getWhiteBackgroundWithBlackTextLogoPath($order['app_id'], 'http://nginx' . ImgHelp::removeImageDomain($storeSetting['logoUrl'] ?? ''));
                    $printer->appendImg($whiteBackgroundWithBlackTextLogoPath, 150, false, -25);
                    $printer->lineFeed(1);
                } catch (\Throwable $th) {
                    //throw $th;
                }
            }
            //
            $printer->lineFeed(1, 10);
            $printer->setAlignment(ImgFont::ALIGN_CENTER);
            $printer->setFontWeight(2);
            $printer->setFontSize(26);
            $printer->setTextLineHeight(50);
            $printer->appendText($order->_title ?: ($order->pay_time ? __("结账单") : __("预结账单")));
            $printer->setFontSize(20);
            $printer->setFontWeight(1);
            $printer->lineFeed();
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
            // 发票信息
            if ($order->_template == 3 && $order->invoiceInfo && ($order->invoiceInfo->company_name || $order->invoiceInfo->company_addr || $order->invoiceInfo->company_tax_number || $order->invoiceInfo->company_phone)) {
                $printer->appendSplitline(true, 40);
                if ($order->invoiceInfo->company_name) {
                    $printer->appendText($order->invoiceInfo->company_name ?: '');
                    $printer->lineFeed(1, ($order->invoiceInfo->company_addr || $order->invoiceInfo->company_tax_number || $order->invoiceInfo->company_phone) ? 50 : 40);
                }
                if ($order->invoiceInfo->company_addr) {
                    $printer->appendText($order->invoiceInfo->company_addr ?: '');
                    $printer->lineFeed(1, ($order->invoiceInfo->company_tax_number || $order->invoiceInfo->company_phone) ? 50 : 40);
                }
                if ($order->invoiceInfo->company_tax_number) {
                    $printer->appendText($order->invoiceInfo->company_tax_number ?: '');
                    $printer->lineFeed(1, $order->invoiceInfo->company_phone ? 50 : 40);
                }
                if ($order->invoiceInfo->company_phone) {
                    $printer->appendText($order->invoiceInfo->company_phone ?: '');
                    $printer->lineFeed(1, 40);
                }
            }
            // 
            $printer->appendSplitline();
            $printer->recoverDefaultTextLineHeight();
            $printer->setAlignment(ImgFont::ALIGN_LEFT);
            if ($order['table_no']) {
                $printer->setFontWeight(2);
                $printer->setFontSize(28);
                $printer->setTextLineHeight(preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no']) ? 68 : 50);
                $printer->appendText(__("桌号") . " : {$order['table_no']}{$orderName} ({$order['meal_num']}{$name})");
                $printer->recoverDefaultTextLineHeight();
                $printer->lineFeed();
            } else if ($order['call_no']) {
                $printer->setFontWeight(2);
                $printer->setFontSize(28);
                $printer->appendText(__("取单号") . " : {$order['call_no']}{$orderName}");
                $printer->lineFeed();
            }
            $printer->setFontWeight(1);
            $printer->setFontSize(20);
            $printer->setAlignment(ImgFont::ALIGN_LEFT);
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
        if ($template != 3) {
            $printer->setTextLineHeight(30);
            $printer->printInColumns(
                [__("商品"), ($this->lang == 'en' || $this->lang == 'th' || $this->lang == 'tr' || $this->lang == 'my') ? 220 : 320, ImgFont::ALIGN_LEFT],
                [ __("单价") . '|' . __("数量"), ($this->lang == 'en' || $this->lang == 'th' || $this->lang == 'tr' || $this->lang == 'my') ? 230 : 120 , ImgFont::ALIGN_CENTER],
                [__("小计"), 0, ImgFont::ALIGN_RIGHT],
            );
        }
        $printer->appendSplitline();
        $printer->lineFeed(1, 40);
        $printer->setTextLineHeight(50);
        // 赠品金额
        $freeMoney = 0;
        // 商品列表
        $productNum = 0;
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
                $printer->printInColumns(
                    [$buffetNameText, 320, ImgFont::ALIGN_LEFT],
                    [helper::amountPermillage($buffet['price']) . '*' . $buffet['num'] , 120 , ImgFont::ALIGN_CENTER],
                    [$this->getPriceAndUnit($buffet['total_price']), 0, ImgFont::ALIGN_RIGHT],
                );
            }
            foreach ($item['buffetDiscount'] ?? [] as $discount) {
                $productNum += $discount['num'];
                $printer->printInColumns(
                    [$discount['name_text'], 320, ImgFont::ALIGN_LEFT],
                    [helper::amountPermillage($discount['price']) . '*' . $discount['num'] , 120 , ImgFont::ALIGN_CENTER],
                    ['-' . $this->getPriceAndUnit($discount['total_price']), 0, ImgFont::ALIGN_RIGHT],
                );
            }
            foreach ($item['delay'] ?? [] as $delay) {
                $productNum += $delay['num'];
                $printer->printInColumns(
                    [$delay['name_text'], 320, ImgFont::ALIGN_LEFT],
                    [helper::amountPermillage($delay['price']) . '*' . $delay['num'] , 120 , ImgFont::ALIGN_CENTER],
                    [$this->getPriceAndUnit($delay['total_price']), 0, ImgFont::ALIGN_RIGHT],
                );
            }
            foreach (array_values($item['products']) as $key => $product) {
                $productNum += $product['total_num'];
                $gift = ($product['is_free'] ?? 0) > 0 ? ('(' . __("赠") . ') ') : '';
                $productAttr = (new OrderProduct)->getProductAttrAttr($product['product_attr'], true);
                $productName = $gift . $product['product_name'] . "\n" . $productAttr;
                $product_total_product_price = $product['total_product_price'];
                if ($gift) {
                    $freeMoney += $product['total_product_price'];
                    $product_total_product_price = 0;
                }
                $printer->setTextLineHeight(45);
                $printer->printInColumns(
                    [$productName, 320, ImgFont::ALIGN_LEFT],
                    [helper::amountPermillage($product['product_price']) . '*' . $product['total_num'] , 120 , ImgFont::ALIGN_CENTER],
                    [$this->getPriceAndUnit($product_total_product_price), 0, ImgFont::ALIGN_RIGHT],
                );
                $printer->setTextLineHeight(50);
                if ($key != count($item['products']) - 1) {
                    $printer->lineFeed(1, 10);
                }
            }
        }
        // 商品数量
        $printer->appendSplitline();
        $printer->lineFeed();
        $printer->setAlignment(ImgFont::ALIGN_RIGHT);
        // 商品金额 = 订单总价 - 赠品金额
        $total_product_price = $order['total_product_price'] - $freeMoney;
        if ($template == 3) {
            $printer->printInColumns(
                [__("商品数量") . ': ' . $productNum, 250, ImgFont::ALIGN_LEFT],
                [__("商品金额") . ': ' . $this->getPriceAndUnit($total_product_price), 0, ImgFont::ALIGN_RIGHT],
            );
        } else {
            $printer->setAlignment(ImgFont::ALIGN_RIGHT);
            $printer->appendText(__("商品数量") . ': ' . $productNum);
            $printer->lineFeed();
            $printer->appendText(__("商品金额") . ': ' . $this->getPriceAndUnit($total_product_price));
            $printer->lineFeed();
        }
        $printer->setAlignment(ImgFont::ALIGN_RIGHT);
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
            $printer->appendSplitline();
            $printer->lineFeed(1);
        } else {
            $printer->lineFeed(1, 10);
        }
        $printer->setFontWeight(2);
        $printer->setFontSize(26);
        $printer->printInColumns(
            [__("合计应收"), 280, ImgFont::ALIGN_LEFT],
            [$this->getPriceAndUnit($order['actual_receive_price']), 0, ImgFont::ALIGN_RIGHT],
        );
        $printer->setFontWeight(1);
        $printer->setFontSize(20);
        // 商品已含税
        if ($order['consumption_tax_money'] > 0 && $order['consumption_tax_type'] == 1 && ($this->consumptionTax == 1 || $this->consumptionTax == 2)) {
            $printer->appendSplitline();
            $printer->lineFeed(1);
            $printer->setAlignment(ImgFont::ALIGN_RIGHT);
            $printer->appendText(__("合计 (其中VAT)"));
            $printer->lineFeed(1);
            foreach ($percentageList as $key => $percentage) {
                if ($this->lang == 'ja') {
                    $printer->printInColumns(
                        [$percentage['tax_rate'] . '%' . __("的对象"), 280, ImgFont::ALIGN_LEFT],
                        [helper::amountPermillage($percentage['total_price'])  . ' (' . helper::amountPermillage($percentage['consumption_tax']) . ')', 0, ImgFont::ALIGN_RIGHT],
                    );
                } else {
                    $printer->printInColumns(
                        ['VAT (' . ($percentage['tax_rate'] . '%)'), 280, ImgFont::ALIGN_LEFT],
                        [helper::amountPermillage($percentage['total_price'])  . ' (' . helper::amountPermillage($percentage['consumption_tax']) . ')', 0, ImgFont::ALIGN_RIGHT],
                    );
                }
            }
        }
        // 支付方式
        if ($order->pay_status['value'] == OrderPayStatusEnum::SUCCESS) {
            $payTypes = $order->payType()->select()->append(['pay_type'])->toArray();
            if (count($payTypes) > 0) {
                $printer->appendSplitline();
                $printer->lineFeed(1);
                foreach ($payTypes as $payType) {
                    $printer->printInColumns(
                        [__("支付方式"), 280, ImgFont::ALIGN_LEFT],
                        [$payType['pay_type']['text'] .  ( $this->lang == 'my' ? ' ' : ''), 0, ImgFont::ALIGN_RIGHT],
                    );
                    $printer->printInColumns(
                        [__("实收金额"), 280, ImgFont::ALIGN_LEFT],
                        [$payType['value'] == -1 ? '0' : $this->getPriceAndUnit($payType['price']), 0, ImgFont::ALIGN_RIGHT],
                    );
                    if (($order['change_due'] ?? 0) > 0 && $payType['pay_type']['value'] == OrderPayTypeEnum::CASH) {
                        $printer->printInColumns(
                            [__("找零"), 280, ImgFont::ALIGN_LEFT],
                            [helper::amountPermillage($order['change_due'] ?? 0), 0, ImgFont::ALIGN_RIGHT],
                        );
                    }
                }
            }
        }
        // 会员积分
        if ($order->user) {
            $printer->appendSplitline();
            $printer->lineFeed(1);
            $pointnum = $order['is_free'] != 0 ? 0 : ($order->points_bonus ?: 0);
            $balance = $payTime ? ($order['surplus_balance'] ?? 0) : $order->surplusBalance();
            $printer->printInColumns(
                [__("会员剩余余额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($balance), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("本次积分"), 280, ImgFont::ALIGN_LEFT],
                [$pointnum > 0 ? strval(helper::number2($pointnum)) : '0', 0, ImgFont::ALIGN_RIGHT],
            );
        }
        // 打印支付二维码
        if ($paramData && ($paramData['pay_value'] ?? 0)) {
            $payValue = $paramData['pay_value'] ?? 0;
            $paymentOrderId = $paramData['payment_order_id'] ?? 0;
            $payType = (new PayType([], $order['app_id']))->where('source', '<>', 0)->where('value', $payValue)->find();
            $paymentOrder = $payType?->source == 2 ? (new PaymentOrder([], $order['app_id']))->where('id', $paymentOrderId)->find() : null;
            if ($payType && ($payType->source != 2 || ($payType->source == 2 && $paymentOrder))) {
                $printer->appendSplitline();
                $printer->lineFeed(1);
                $printer->setAlignment(ImgFont::ALIGN_CENTER);
                $printer->setTextLineHeight(35);
                $printer->appendText(__('请用') .' '. $payType->name .' '. __('扫一扫支付'));
                $printer->lineFeed(1);
                // 1-自行添加 2-LianLianPay
                if ($payType->source == 1) {
                    $printer->appendImg('http://nginx' . $payType->qrcode);
                } else {
                    if ($paymentOrder->link_url) {
                        $printer->appendQrcode($paymentOrder->link_url);
                    } else {
                        $printer->appendQrcode($paymentOrder->qr_code, 280, 10, true);
                    }
                }
                $printer->setTextLineHeight(50);
            }
        }
        // 技术支持方
        $printer->appendSplitline();
        $printer->lineFeed(1);
        $printer->setAlignment(ImgFont::ALIGN_CENTER);
        if ($this->lang == 'tr') {
            $printer->appendText("Ziyaretiniz için teşekkür ederiz! Bu mağaza");
            $printer->lineFeed(1);
            $printer->appendText("tarafından: " . $brandName . " Sistem destek sağlar.");
        } else if ($this->lang == 'th') {
            $printer->appendText("ขอบคุณที่แวะมาหากัน!สนับสนุนโดย " . $brandName);
        } else {
            $printer->appendText(__("感谢您的光临！本店由") . " " . $brandName . " " . __("系统提供支持。"));
        }
        // 
        $printer->lineFeed(3);
        // 
        $openMoneybox = false;
        // if (!$isPrePrint && $order->pay_time) {
        //     foreach ($order->payType as $pay) {
        //         if ($pay['value'] == OrderPayTypeEnum::CASH) {
        //             $openMoneybox = $this->isSunmi ? 2 : true;
        //             break;
        //         }
        //     }
        // }
        //
        return $printer->save('', !$this->isSunmi, $openMoneybox);
    }
}
