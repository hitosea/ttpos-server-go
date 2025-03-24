<?php

namespace app\common\template\business;

use help\DateHelp;
use app\common\library\helper;
use app\common\model\product\Product;
use app\common\template\BaseTemplate;
use app\common\model\supplier\Supplier;
use app\common\model\product\ProductSku;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\library\printer\party\SunmiCloudPrinter;
use app\common\model\product\Category as CategoryModel;

/**
 * 商米 - 营业数据模版
 */
class SunmiBusinessTemplate extends BaseTemplate
{
    /**
     * 生成模版
     */
    public function create($data, $printerType, $shopName, $mode, $startTime, $endTime)
    {
        $isBalance = Supplier::where('shop_supplier_id', $data['supplier']['shop_supplier_id'] ?? 0)->value('is_open_member') ?: 0;
        $isOneself = $printerType != PrinterTypeEnum::SUNMI_LAN && $printerType != PrinterTypeEnum::SUNMI_CLOUD;
        // 佛历
        if ($this->defaultCalendar == '3') {
            $startTime = DateHelp::changeBuddhistCalendar($startTime);
            $endTime = DateHelp::changeBuddhistCalendar($endTime);
        }
        //
        $printer = new SunmiCloudPrinter(567);
        $printer->lineFeed();
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        $printer->appendText("{$shopName}\n");
        $printer->lineFeed();
        $printer->setLineSpacing(45);
        $printer->setPrintModes(true, true, false);
        $printer->setCharacterSize(2, 1);
        $printer->appendText(__("营业数据"));
        $printer->setCharacterSize(1, 1);
        $printer->lineFeed();
        if ($isOneself) {
            $printer->setLineSpacing(20);
        }
        $printer->lineFeed();
        $printer->setLineSpacing(45);
        //
        $printer->setPrintModes(false, false, false);
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        $printer->appendText($startTime . " " . __("至") . " " . $endTime);
        $printer->lineFeed();
        $printer->setLineSpacing($isOneself ? 25 : 40);
        $printer->lineFeed();
        $printer->setLineSpacing(45);
        //
        $printer->restoreDefaultLineSpacing();
        $printer->setPrintModes(false, false, false);
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
        // 按支付方式
        if ($mode == 1) {
            $printer->setupColumns(
                [360, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $printer->setPrintModes(true, false, false);
            $printer->printInColumns(__("实收金额"), $this->getPriceAndUnit($data['total_amount']));
            $printer->setPrintModes(false, false, false);
            $printer->appendText("------------------------------------------------\n");
            $printer->lineFeed();
            foreach ($data['incomes'] as $key => $income) {
                if ($income['pay_type'] == -1) {
                    $income['pay_type_way'] = __('免单金额');
                }
                $printer->printInColumns($income['pay_type_way'], $this->getPriceAndUnit($income['price']));
                $printer->lineFeed();
            }
        }
        // 按商品分类
        else if ($mode == 2) {
            $printer->setupColumns(
                [300, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [96, SunmiCloudPrinter::ALIGN_CENTER, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->setPrintModes(true, false, false);
            $printer->printInColumns(__("分类"), __("数量"), __("小计"));
            $printer->setPrintModes(false, false, false);
            $printer->appendText("------------------------------------------------\n");
            foreach ($data['categorys'] as $key => $category) {
                $printer->printInColumns((new CategoryModel)->getPathNameTextAttr($category['name'], $category) . '', "{$category['sales']}",  $this->getPriceAndUnit($category['prices']));
                $printer->lineFeed();
            }
            $printer->appendText("------------------------------------------------\n");
            //
            $printer->setupColumns(
                [360, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            if ($data['sales_num'] > 0) {
                $printer->printInColumns(__("销售笔数"), "{$data['sales_num']}");
                $printer->lineFeed();
            }
            //
            foreach ($data['incomes'] as $key => $income) {
                if ($income['pay_type'] == -1) {
                    $income['pay_type_way'] = __('免单金额');
                }
                $printer->printInColumns($income['pay_type_way'], $this->getPriceAndUnit($income['price']));
                $printer->lineFeed();
            }
            //
            if ($data['refund_amount'] > 0) {
                $printer->printInColumns(__("退款金额"), $this->getPriceAndUnit($data['refund_amount']));
                $printer->lineFeed();
            }
            $printer->printInColumns(__("实收金额"), $this->getPriceAndUnit($data['total_amount']));
            $printer->lineFeed();
        }
        // 按商品
        else if ($mode == 3) {
            $printer->setupColumns(
                [300, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [120, SunmiCloudPrinter::ALIGN_CENTER, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->setPrintModes(true, false, false);
            $printer->printInColumns(__("商品名称"), __("销量"), __("小计"));
            $printer->setPrintModes(false, false, false);
            $printer->appendText("------------------------------------------------\n");
            foreach ($data['products'] as $key => $product) {
                $specNameText = ProductSku::getSpecNameTextAttr($product['spec_name'] ?: '', $product);
                $product_name = Product::getProductNameTextAttr($product['product_name'] ?: '', $product) . ($specNameText ? " ($specNameText)" : '');
                $printer->printInColumns($product_name, ' ' . helper::amountPermillage($product['product_price']) . '*' . "{$product['sales']}", ' ' . $this->getPriceAndUnit($product['prices']));
                $printer->lineFeed();
            }
        }
        // 全部
        else {
            $printer->setLineSpacing($isOneself ? 25 : 20);
            $printer->setupColumns(
                [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->printInColumns(__("总销售额"), $this->getPriceAndUnit($data['all']['receivable_price']));
            $printer->lineFeed();
            $printer->printInColumns(__("原商品金额"), $this->getPriceAndUnit($data['all']['not_tax_total_product_price']));
            $printer->lineFeed();
            $printer->printInColumns(__("服务费"), $this->getPriceAndUnit($data['all']['service_money']));
            $printer->lineFeed();
            $printer->setupColumns(
                [400, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->printInColumns(__("支付手续费"), $this->getPriceAndUnit($data['all']['pay_fee_money']));
            $printer->lineFeed();
            $printer->setupColumns(
                [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->printInColumns(__("税费"), $this->getPriceAndUnit($data['all']['consumption_tax_money']));
            $printer->lineFeed();
            $printer->printInColumns(__("商品数量"), $data['all']['product_num'] . '');
            $printer->lineFeed();
            $printer->printInColumns(__("优惠折扣"), $this->getPriceAndUnit($data['all']['discount_money']));
            $printer->lineFeed();
            if ($isBalance == 1 || $data['all']['user_discount_money'] > 0) {
                $printer->printInColumns(__("会员折扣"), $this->getPriceAndUnit($data['all']['user_discount_money']));
                $printer->lineFeed();
            }
            $printer->printInColumns(__("退款金额"), $this->getPriceAndUnit($data['all']['refund_money']));
            $printer->lineFeed();
            $printer->printInColumns(__("免单金额"), $this->getPriceAndUnit($data['all']['free_order_price']));
            $printer->lineFeed();
            $printer->setPrintModes(true, false, false);
            $printer->setCharacterSize(2, 1);
            $printer->printInColumns(__("实收金额"), $this->getPriceAndUnit($data['all']['received_price']));
            $printer->setCharacterSize(1, 1);
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed();
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2);
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
            // 会员充值
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2); 
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('会员数据'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->printInColumns(__("充值金额"), $this->getPriceAndUnit($data['all']['recharge_amount']));
            $printer->lineFeed(1);
            $printer->printInColumns(__("赠送金额"), $this->getPriceAndUnit($data['all']['gift_money']));
            $printer->lineFeed(1);
            $printer->printInColumns(__("赠送积分"), $data['all']['gift_points'] . '');
            $printer->lineFeed(1);
            // 未结账相关
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2); 
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('未结账数据'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(2);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->printInColumns(__("订单数"), $data['all']['not_settled_total_order_num'] . '');
            $printer->lineFeed(1);
            $printer->printInColumns(__("金额"), $this->getPriceAndUnit($data['all']['not_settled_total_price']));
            $printer->lineFeed(1);
            // 合计
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2); 
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('合计'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(2);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->printInColumns(__("所有订单数"), $data['all']['total_order_num'] . '');
            $printer->lineFeed(1);
            $printer->printInColumns(__("桌数"), $data['all']['total_table_num'] . '');
            $printer->lineFeed(1);
            $printer->printInColumns(__("人数"), $data['all']['total_people_num'] . '');
            $printer->lineFeed(1);
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("最小/大订单金额"), $this->getPriceAndUnit($data['all']['min_order_price']) . '/' . $this->getPriceAndUnit($data['all']['max_order_price']));
            if ($this->lang == 'my') {
                $printer->setLineSpacing($isOneself ? 25 : 20);
            }
            $printer->lineFeed(1);
            $printer->printInColumns(__("平均订单金额"), $this->getPriceAndUnit($data['all']['avg_order_price']));
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
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->printInColumns(__("订单数（桌数）"), $data['all']['table_order_num'] . '');
            $printer->lineFeed(1);
            $printer->setupColumns(
                [320, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->printInColumns(__("人数"), $data['all']['table_people_num'] . '');
            $printer->lineFeed(1);
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("最小/大订单金额"), $this->getPriceAndUnit($data['all']['table_min_order_price']) . '/' . $this->getPriceAndUnit($data['all']['table_max_order_price']));
            if ($this->lang == 'my') {
                $printer->setLineSpacing($isOneself ? 25 : 20);
            }
            $printer->lineFeed(1);
            $printer->printInColumns(__("平均订单金额"), $this->getPriceAndUnit($data['all']['table_avg_order_price']));
            $printer->lineFeed(1);
            $printer->printInColumns(__("人均"), $this->getPriceAndUnit($data['all']['table_people_avg']));
            $printer->lineFeed(2);
            // 收银方式
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, false, false);
            $printer->appendText(__('点餐方式'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(2);
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->printInColumns(__("订单数"), $data['all']['cashier_order_num'] . '');
            $printer->lineFeed(1);
            if ($this->lang == 'my') {
                $printer->setLineSpacing(50);
            }
            $printer->printInColumns(__("最小/大订单金额"), $this->getPriceAndUnit($data['all']['cashier_min_order_price']) . '/' . $this->getPriceAndUnit($data['all']['cashier_max_order_price']));
            if ($this->lang == 'my') {
                $printer->setLineSpacing($isOneself ? 25 : 20);
            }
            $printer->lineFeed(1);
            $printer->printInColumns(__("平均订单金额"), $this->getPriceAndUnit($data['all']['cashier_avg_order_price']));
            $printer->lineFeed(1);
            // 支付方式
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2);
            $printer->setPrintModes(true, false, false);
            $printer->setupColumns(
                [$this->lang == 'en' || $this->lang == 'tr' || $this->lang == 'th' ? 220 : 270, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [$this->lang == 'en' || $this->lang == 'tr' || $this->lang == 'my' ? 200 : 180, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->printInColumns(__("支付方式"), __('订单数'), $this->lang == 'my' ? 'ငွေပမာဏ ' : __('金额'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setupColumns(
                [300, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [20, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $totalPayPrice = 0;
            foreach ($data['incomes'] as $key => $income) {
                if ($income['pay_type'] !== -1) {
                    $printer->printInColumns($income['pay_type_way'] . '',  $income['order_num'] . '',  $this->getPriceAndUnit($income['price']));
                    $printer->lineFeed();
                    $totalPayPrice += $income['price'];
                }
            }
            if ($totalPayPrice > 0) {
                $printer->printInColumns(__("总金额"), '', $this->getPriceAndUnit($totalPayPrice));
                $printer->lineFeed(1);
            }
            // 高峰时间
            $printer->appendText("------------------------------------------------");
            $printer->lineFeed(2);
            $printer->setPrintModes(true, false, false);
            $printer->setupColumns(
                [$this->lang == 'en' || $this->lang == 'tr' || $this->lang == 'th'  ? 220 : 270, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [$this->lang == 'en' || $this->lang == 'tr' || $this->lang == 'my' ? 200 : 180, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            $printer->printInColumns(__('高峰时间'), __('订单数'), $this->lang == 'my' ? 'ငွေပမာဏ ' : __('金额'));
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed(1);
            $printer->setupColumns(
                [300, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [20, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0]
            );
            foreach ($data['all']['peak_hour_list'] as $key => $peak) {
                $printer->printInColumns($peak['time_period'], $peak['num'] . '', $this->getPriceAndUnit($peak['amount']));
                $printer->lineFeed();
            }
            $printer->lineFeed(4);
        }
        // Print and exit page mode
        $printer->printAndExitPageMode();
        $printer->lineFeed(4);
        $printer->cutPaper(false);
        //
        return $printer->orderData;
    }
}
