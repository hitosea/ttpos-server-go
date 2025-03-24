<?php

namespace app\common\template\business;

use help\DateHelp;
use app\common\model\product\Product;
use app\common\model\shop\BindRecord;
use app\common\template\BaseTemplate;
use app\common\model\supplier\Supplier;
use app\common\model\product\ProductSku;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\library\printer\party\SunmiCloudPrinter;
use app\common\model\product\Category as CategoryModel;

/**
 * 营业数据模版 - 芯烨打印机, Compax 收银打印机 80mm 自带
 */
class XprinterBusinessTemplate extends BaseTemplate
{
    /**
     * 生成模版
     */
    public function create($printers, $data, $printerType, $shopName, $mode, $startTime, $endTime)
    {
        $is_balance = Supplier::where('shop_supplier_id', $data['supplier']['shop_supplier_id'] ?? 0)->value('is_open_member') ?: 0;
        // 佛历
        if ($this->defaultCalendar == '3') {
            $startTime = DateHelp::changeBuddhistCalendar($startTime);
            $endTime = DateHelp::changeBuddhistCalendar($endTime);
        }
        //
        $width = 48 - ($printers == BindRecord::BRAND_A1_1510P ? 1 : 0);
        $differenceWidth = $printers == BindRecord::BRAND_A1_1510P && $this->currencyUnit == '¥' ? 1 : 0;
        if ($printers == BindRecord::BRAND_A1_1510P && $this->currencyUnit == '￥') {
            $differenceWidth = 1;
            $this->currencyUnit = "\xC2\xA5";
        }
        //
        $printer = new SunmiCloudPrinter(567);
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        $printer->appendText("{$shopName}\n");
        $printer->lineFeed();
        $printer->setLineSpacing(40);
        $printer->setPrintModes(true, true, false);
        $printer->setCharacterSize(2, 2);
        $printer->appendText(str_replace("ー", "-", __("营业数据")));
        $printer->lineFeed();
        $printer->setLineSpacing(25);
        $printer->lineFeed();
        if ($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
            $printer->lineFeed(2);
        }
        $printer->setLineSpacing(70);
        $printer->setCharacterSize(1, 1);
        $printer->restoreDefaultLineSpacing();
        //
        $printer->setPrintModes(false, false, false);
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        $printer->appendText($startTime . " " . __("至") . " " . $endTime);
        $printer->lineFeed();
        $printer->lineFeed();
        //
        $printer->restoreDefaultLineSpacing();
        $printer->setPrintModes(false, false, false);
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
        // 按支付方式
        if ($mode == 1) {
            $printer->setPrintModes(true, false, false);
            $printer->appendText(printText(__("实收金额"), '', $this->getPriceAndUnit($data['total_amount']), $width));
            $printer->setPrintModes(false, false, false);
            $printer->appendText("\n------------------------------------------------");
            $printer->lineFeed();
            foreach ($data['incomes'] as $key => $income) {
                if ($income['pay_type'] == -1) {
                    $income['pay_type_way'] = __('免单金额');
                }
                $printer->appendText(printText($income['pay_type_way'], '', $this->getPriceAndUnit($income['price']), $width, 30, 1, 24));
                $printer->lineFeed();
                $printer->lineFeed();
            }
        }
        // 按商品分类
        else if ($mode == 2) {
            $printer->setPrintModes(true, false, false);
            $printer->appendText(printText(__("分类"), __("数量"), __("小计"), $width, 29));
            $printer->setPrintModes(false, false, false);
            $printer->appendText("\n------------------------------------------------");
            $printer->lineFeed();
            foreach ($data['categorys'] as $key => $category) {
                $printer->appendText(printText((new CategoryModel)->getNameTextAttr($category['name']), $category['sales'] . '', $this->getPriceAndUnit($category['prices']), $width, 29));
                $printer->lineFeed();
                if ($key != count($data['categorys']) - 1) {
                    $printer->lineFeed();
                }
            }
            $printer->appendText("\n------------------------------------------------\n");
            //
            if ($data['sales_num'] > 0) {
                $printer->appendText(printText(__("销售笔数"), '', "{$data['sales_num']}", $width));
                $printer->lineFeed();
                $printer->lineFeed();
            }
            //
            foreach ($data['incomes'] as $key => $income) {
                if ($income['pay_type'] == -1) {
                    $income['pay_type_way'] = __('免单金额');
                }
                $printer->appendText(printText($income['pay_type_way'], '', $this->getPriceAndUnit($income['price']), $width, 30, 1, 24));
                $printer->lineFeed();
                $printer->lineFeed();
            }
            //
            if ($data['refund_amount'] > 0) {
                $printer->appendText(printText(__("退款金额"), '', $this->getPriceAndUnit($data['refund_amount']), $width));
                $printer->lineFeed();
                $printer->lineFeed();
            }
            $printer->appendText(printText(__("实收金额"), '', $this->getPriceAndUnit($data['total_amount']), $width));
        }
        // 按商品
        else if ($mode == 3) {
            $printer->setPrintModes(true, false, false);
            $printer->appendText(printText(__("商品名称"), __("销量"), __("小计"), $width, 26));
            $printer->setPrintModes(false, false, false);
            $printer->appendText("\n------------------------------------------------");
            foreach ($data['products'] as $key => $product) {
                $specNameText = ProductSku::getSpecNameTextAttr($product['spec_name'] ?: '', $product);
                $product_name = Product::getProductNameTextAttr($product['product_name'] ?: '', $product) . ($specNameText ? " ($specNameText)" : '');
                $printer->appendText(printText($product_name, ' ' . $this->getPriceAndUnit($product['product_price']) . '*' . $product['sales'], $this->getPriceAndUnit($product['prices']), $width - $differenceWidth, 26, 16, 16));
                $printer->lineFeed();
                $printer->lineFeed();
            }
        }
        // 全部
        else {
            $printer->setLineSpacing($printers == BindRecord::BRAND_A1_1510P ? 40 : 90);
            //
            $printer->appendText(printText(__("总销售额"), '', $this->getPriceAndUnit($data['all']['receivable_price']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("原商品金额"), '', $this->getPriceAndUnit($data['all']['not_tax_total_product_price']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("服务费"), '', $this->getPriceAndUnit($data['all']['service_money']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("支付手续费"), '', $this->getPriceAndUnit($data['all']['pay_fee_money']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("税费"), '', $this->getPriceAndUnit($data['all']['consumption_tax_money']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("商品数量"), '', $data['all']['product_num'], $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("优惠折扣"), '', $this->getPriceAndUnit($data['all']['discount_money']), $width));
            $printer->lineFeed(1);
            if ($is_balance == 1 || $data['all']['user_discount_money'] > 0) {
                $printer->appendText(printText(__("会员折扣"), '', $this->getPriceAndUnit($data['all']['user_discount_money']), $width));
                $printer->lineFeed(1);
            }
            $printer->appendText(printText(__("退款金额"), '', $this->getPriceAndUnit($data['all']['refund_money']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("免单金额"), '', $this->getPriceAndUnit($data['all']['free_order_price']), $width));
            $printer->lineFeed(1);
            $printer->setPrintModes(true, true, false);
            $printer->setCharacterSize(2, 1);
            $printer->appendText(printText(__("实收金额"), '', $this->getPriceAndUnit($data['all']['received_price']), $width));
            $printer->setCharacterSize(1, 1);
            $printer->setPrintModes(false, false, false);
            $printer->appendText("\n------------------------------------------------");
            $printer->lineFeed(1);
            // 税收百分比对象列表
            foreach ($data['all']['percentage_list'] as $key => $percentage) {
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
                $printer->appendText(printText(__("合计"), '', $this->getPriceAndUnit($percentage['total_price']), $width));
                $printer->lineFeed(1);
                $printer->setAlignment(SunmiCloudPrinter::ALIGN_RIGHT);
                if ($this->lang == 'ja') {
                    $printer->appendText("(" . __('其中消费税') . '' . $this->getPriceAndUnit($percentage['consumption_tax']) . ")");
                } else {
                    $printer->appendText("(" . __('其中VAT') . '' . $this->getPriceAndUnit($percentage['consumption_tax']) . ")");
                }
                $printer->lineFeed(1);
            }
            // 会员充值
            $printer->appendText("------------------------------------------------");
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('会员数据'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("充值金额"), '', $this->getPriceAndUnit($data['all']['recharge_amount']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("赠送金额"), '', $this->getPriceAndUnit($data['all']['gift_money']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("赠送积分"), '', floatval($data['all']['gift_points']) . '', $width));
            $printer->lineFeed(1);
            // 未结账相关
            $printer->appendText("------------------------------------------------");
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('未结账数据'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("订单数"), '', floatval($data['all']['not_settled_total_order_num']) . '', $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("金额"), '', $this->getPriceAndUnit($data['all']['not_settled_total_price']), $width));
            $printer->lineFeed(1);
            // 合计
            $printer->appendText("------------------------------------------------");
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('合计'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("所有订单数"), '', floatval($data['all']['total_order_num']) . '', $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("桌数"), '', floatval($data['all']['total_table_num']) . '', $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("人数"), '', floatval($data['all']['total_people_num']) . '', $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("最小/大订单金额"), '', $this->getPriceAndUnit($data['all']['min_order_price']) . '/' . $this->getPriceAndUnit($data['all']['max_order_price']), $width - $differenceWidth));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("平均订单金额"), '', $this->getPriceAndUnit($data['all']['avg_order_price']), $width));
            $printer->lineFeed(1);
            // 桌台方式
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('桌台方式'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("订单数（桌数）"), '', floatval($data['all']['table_order_num']) . '', $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("人数"), '', floatval($data['all']['table_people_num']) . '', $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("最小/大订单金额"), '', $this->getPriceAndUnit($data['all']['table_min_order_price']) . '/' . $this->getPriceAndUnit($data['all']['table_max_order_price']), $width - $differenceWidth));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("平均订单金额"), '', $this->getPriceAndUnit($data['all']['table_avg_order_price']), $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("人均"), '', $this->getPriceAndUnit($data['all']['table_people_avg']), $width));
            $printer->lineFeed(1);
            // 收银方式
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('点餐方式'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->appendText(printText(__("订单数"), '', floatval($data['all']['cashier_order_num']) . '', $width));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("最小/大订单金额"), '', $this->getPriceAndUnit($data['all']['cashier_min_order_price']) . '/' . $this->getPriceAndUnit($data['all']['cashier_max_order_price']), $width - $differenceWidth));
            $printer->lineFeed(1);
            $printer->appendText(printText(__("平均订单金额"), '', $this->getPriceAndUnit($data['all']['cashier_avg_order_price']), $width));
            $printer->lineFeed(1);
            // 支付方式
            $printer->appendText("------------------------------------------------");
            $printer->setPrintModes(true, false, false);
            $printer->appendText(printText(__('支付方式'), __('订单数'), __('金额'), $width, 24 - (($this->lang == 'en' || $this->lang == 'th' || $this->lang == 'tr') ? 4 : 0), 20, 16));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $totalPayPrice = 0;
            foreach ($data['incomes'] as $key => $income) {
                if ($income['pay_type'] !== -1) {
                    $printer->appendText(printText($income['pay_type_way'], $income['order_num'], $this->getPriceAndUnit($income['price']), $width, 26, 10, 18));
                    $printer->lineFeed();
                    $totalPayPrice += $income['price'];
                }
            }
            if ($totalPayPrice > 0) {
                $printer->appendText(printText(__("总金额"), '', $this->getPriceAndUnit($totalPayPrice), $width, 26, 10, 18));
                $printer->lineFeed();
            }
            // 高峰时间
            $printer->appendText("------------------------------------------------");
            $printer->setPrintModes(true, false, false);
            $printer->appendText(printText(__('高峰时间'), __('订单数'), $this->lang == 'en' ? 'Amount' : __('订单金额'), $width, 24 - (($this->lang == 'en' || $this->lang == 'th' || $this->lang == 'tr') ? 4 : 0), 20, 18));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            foreach ($data['all']['peak_hour_list'] as $key => $peak) {
                $printer->appendText(printText($peak['time_period'], $peak['num'], $this->getPriceAndUnit($peak['amount']), $width, 26, 10, 18));
                $printer->lineFeed();
            }
        }
        $printer->lineFeed(3);
        // Print and exit page mode
        $printer->printAndExitPageMode();
        $printer->lineFeed(4);
        $printer->cutPaper(true);
        //
        return $printer->orderData;
    }
}
