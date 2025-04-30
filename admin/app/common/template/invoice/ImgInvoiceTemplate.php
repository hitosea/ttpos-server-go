<?php

namespace app\common\template\invoice;

use help\DateHelp;
use base\imgs\ImgFont;
use app\common\library\helper;
use app\common\template\BaseTemplate;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\template\bill\ImgBillTemplate;
use app\common\model\settings\PrinterTemplate;
use app\common\model\settings\Setting as SettingModel;

/**
 * 图片打印 - 发票模版
 */
class ImgInvoiceTemplate  extends BaseTemplate
{
    /**
     * 生成模版
     */
    public function create($order, $printType)
    {
        $template = (new PrinterTemplate([], request()->newAppId))->where('id', '=', 7)->value('template');

        $settingStore = $this->setting[SettingEnum::STORE]['values'];
        $shopName = $settingStore['name'] ?? '';
        
        /* *
        * 模版2
        */
        if ($template == 2) {
            $order->_shop_name = $shopName;
            $order->_template = 3;
            $order->_title = __("发票");
            return (new ImgBillTemplate($this->setting, $this->allSourceProductList, $this->isSunmi))->create($order, $printType);
        }

        /* *
        * 模版1
        */
        $settingStore = $this->setting[SettingEnum::STORE]['values'];
        $settingPrinterConfig = $this->setting[SettingEnum::PRINTER]['values'] ?? [];
        $consumptionTax = $settingPrinterConfig['consumption_tax'] ?? 1;
        $settingCloud = SettingModel::getCloudBasic();
        $brandName = $settingCloud['base']['brand_name'] ?? $settingStore['brand_name'] ?? '';
        //
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

        /* *
        * 打印模版
        */
        $printer = new ImgFont(568);
        $printer->setTextLineHeight(50);
        $printer->setImagePadding(0);
        $printer->setAlignment(ImgFont::ALIGN_CENTER);
        $printer->setFontWeight(2);
        $printer->setFontSize(24);
        $printer->appendText("{$shopName}");
        $printer->setFontSize(20);
        $printer->setFontWeight(1);
        $printer->lineFeed();
        $printer->lineFeed(1, 10);
        $printer->appendText(__("非常感谢您今天的到来，我们期待您的再次光临"));
        $printer->lineFeed(1, 60);
        $printer->setFontSize(28);
        $printer->appendText(__("发票"));
        $printer->setFontSize(20);
        $printer->lineFeed();
        $printer->appendText($payTime);
        $printer->lineFeed();
        if ($this->lang == 'ja') {
            $printer->setAlignment(ImgFont::ALIGN_RIGHT);
            $printer->appendText(__("先生/小姐"));
            $printer->lineFeed(1, 20);
            $printer->setAlignment(ImgFont::ALIGN_LEFT);
        }
        $printer->appendSplitline(true);
        // 
        $printer->printInColumns(
            [__("合计"), 260, ImgFont::ALIGN_LEFT, 1, 28],
            [$this->getPriceAndUnit($order->actual_receive_price), 0, ImgFont::ALIGN_RIGHT, 1, 28],
        );
        $printer->printInColumns(
            ['(' .__("其中服务费"), 350, ImgFont::ALIGN_LEFT],
            [$this->getPriceAndUnit($order->service_money). ')', 0, ImgFont::ALIGN_RIGHT],
        );
        if ($consumptionTax != 4) {
            $printer->printInColumns(
                ['(' . __("其中VAT"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($order->consumption_tax_money ?: 0) . ')', 0, ImgFont::ALIGN_RIGHT],
            );
        }
        $printer->appendSplitline(true);
        // 
        if ($consumptionTax != 4) {
            $printer->appendText(__("仅作为餐饮费收取以上金额"));
            $printer->lineFeed(1);
            $printer->setAlignment(ImgFont::ALIGN_RIGHT);
            $printer->appendText(__("合计 (其中VAT)"));
            $printer->lineFeed();
            foreach ($percentageList as $percentage) {
                if ($this->lang == 'ja') {
                    $printer->printInColumns(
                        [floatval($percentage['tax_rate']) . '%' . __("的对象"), 350, ImgFont::ALIGN_LEFT],
                        [helper::amountPermillage($percentage['total_price'])  . ' (' . helper::amountPermillage($percentage['consumption_tax']) . ')', 0, ImgFont::ALIGN_RIGHT],
                    );
                } else {
                    $printer->printInColumns(
                        ['VAT (' . ($percentage['tax_rate'] . '%)'), 350, ImgFont::ALIGN_LEFT],
                        [helper::amountPermillage($percentage['total_price'])  . ' (' . helper::amountPermillage($percentage['consumption_tax']) . ')', 0, ImgFont::ALIGN_RIGHT],
                    );
                }
            }
        }
        if ($order->refund_money > 0) {
            $printer->appendText(__("不包含退款金额") . " " . $this->getPriceAndUnit($order->refund_money ?: 0));
            $printer->lineFeed(1, 40);
        }
        // 支付方式
        $printer->appendSplitline(true, 40);
        $printer->lineFeed(1, 5);
        $payTypes = $order->payType()->select()->append(['pay_type'])->toArray();
        foreach ($payTypes as $key => $payType) {
            if ($payType['value'] == -1) {
                $payType['price'] = 0;
            } else if ($payType['value'] == OrderPayTypeEnum::CASH) {
                $payType['price'] = helper::bcsub($payType['price'], $order['change_due']);
            }
            $printer->setTextLineHeight(40);
            $printer->printInColumns(
                [$payType['pay_type']['text'], 320, ImgFont::ALIGN_LEFT, 2, 20],
                [$this->getPriceAndUnit($payType['price']), 0, ImgFont::ALIGN_RIGHT, 2, 20],
            );
            if ($key < count($payTypes) - 1) {
                $printer->lineFeed(1, 10);
            }
            $printer->setTextLineHeight(50);
        }
        // 发票信息
        if ($order->invoiceInfo && ($order->invoiceInfo->company_name || $order->invoiceInfo->company_addr || $order->invoiceInfo->company_tax_number || $order->invoiceInfo->company_phone)) {
            $printer->appendSplitline(true, 40);
            $printer->appendText(__("发票信息"));
            $printer->lineFeed(1);
            if ($order->invoiceInfo->company_name) {
                $printer->appendText(__("公司名称") . ": " . $order->invoiceInfo->company_name ?: '');
                $printer->lineFeed(1, ($order->invoiceInfo->company_addr || $order->invoiceInfo->company_tax_number || $order->invoiceInfo->company_phone) ? 50 : 40);
            }
            if ($order->invoiceInfo->company_addr) {
                $printer->appendText(__("公司地址") . ": " . $order->invoiceInfo->company_addr ?: '');
                $printer->lineFeed(1, ($order->invoiceInfo->company_tax_number || $order->invoiceInfo->company_phone) ? 50 : 40);
            }
            if ($order->invoiceInfo->company_tax_number) {
                $printer->appendText(__("税号") . ": " . $order->invoiceInfo->company_tax_number ?: '');
                $printer->lineFeed(1, $order->invoiceInfo->company_phone ? 50 : 40);
            }
            if ($order->invoiceInfo->company_phone) {
                $printer->appendText(__("联系电话") . ": " . $order->invoiceInfo->company_phone ?: '');
                $printer->lineFeed(1, 40);
            }
        }
        // 信息
        $printer->setTextLineHeight(50);
        $printer->appendSplitline(true, 40);
        $printer->appendText(__("收银员") . ": " . $order->cashier?->real_name ?: '');
        $printer->lineFeed(1);
        $printer->appendText(__("订单号") . ": " . $order->order_no ?: '');
        $printer->lineFeed(1);
        $printer->appendText(__("打印次数") . ": " . $order->print_num);
        $printer->lineFeed(1);
        if ($company) {
            $printer->setTextLineHeight(40);
            $printer->appendText(__("公司名称") . ": " . $company);
            $printer->recoverDefaultTextLineHeight();
            $printer->lineFeed(1);
        }
        if ($address) {
            $printer->setTextLineHeight(40);
            $printer->appendText(__("地址") . ($this->lang == 'th' ? " : " : ": ") . $address);
            $printer->recoverDefaultTextLineHeight();
            $printer->lineFeed(1);
        }
        if ($taxNumber) {
            $printer->setTextLineHeight(40);
            $printer->appendText(__("税号") . ": " . $taxNumber);
            $printer->recoverDefaultTextLineHeight();
            $printer->lineFeed(1);
        }
        if ($phone) {
            $printer->setTextLineHeight(40);
            $printer->appendText(__("电话") . ": " . $phone);
            $printer->recoverDefaultTextLineHeight();
            $printer->lineFeed(1);
        }
        $printer->lineFeed(1, 40);
        //
        $printer->appendText("*" . __("保管注意事项"));
        $printer->lineFeed(1);
        $printer->appendText(__("如需保管时请将印刷页面朝内折叠"));
        // 技术支持方
        $printer->lineFeed(1);
        //
        $printer->appendSplitline(true);
        // 
        $printer->setAlignment(ImgFont::ALIGN_CENTER);
        if ($this->lang == 'tr') {
            $printer->appendText("Ziyaretiniz için teşekkür ederiz! Bu mağaza");
            $printer->lineFeed(1);
            $printer->appendText("tarafından : " . $brandName . " Sistem destek sağlar.");
        } else if ($this->lang == 'th') {
            $printer->appendText("ขอบคุณที่แวะมาหากัน!สนับสนุนโดย " . $brandName);
        } else {
            $printer->appendText(__("感谢您的光临！本店由") . " " . $brandName . " " . __("系统提供支持。"));
        }
        $printer->lineFeed(3);
        // 
        return $printer->save('', !$this->isSunmi);
    }
}
