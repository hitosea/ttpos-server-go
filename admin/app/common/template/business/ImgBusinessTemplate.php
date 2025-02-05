<?php

namespace app\common\template\business;

use help\DateHelp;
use base\imgs\ImgFont;
use app\common\library\helper;
use app\common\model\product\Product;
use app\common\template\BaseTemplate;
use app\common\model\supplier\Supplier;
use app\common\model\product\ProductSku;
use app\common\model\product\Category as CategoryModel;

/**
 * 图片打印 - 营业数据模版
 */
class ImgBusinessTemplate extends BaseTemplate
{
    /**
     * 生成模版
     */
    public function create($data, $printerType, $shopName, $mode, $startTime, $endTime)
    {
        $isBalance = Supplier::where('shop_supplier_id', $data['supplier']['shop_supplier_id'] ?? 0)->value('is_open_member') ?: 0;

        // 佛历
        if ($this->defaultCalendar == '3') {
            $startTime = DateHelp::changeBuddhistCalendar($startTime);
            $endTime = DateHelp::changeBuddhistCalendar($endTime);
        }
        //
        $printer = new ImgFont(568);
        $printer->setAlignment(ImgFont::ALIGN_CENTER);
        $printer->appendText("{$shopName}");
        $printer->lineFeed(1);
        $printer->lineFeed(1, 20);
        $printer->setFontSize(28);
        $printer->appendText(__("营业数据"));
        $printer->lineFeed(1, 58);
        $printer->setFontSize(20);
        $printer->appendText($startTime . " " . __("至") . " " . $endTime);
        $printer->lineFeed(1, 80);
        $printer->restoreDefault();
        $printer->setImagePadding(0);
        $printer->setAlignment(ImgFont::ALIGN_LEFT);
        // 按支付方式
        if ($mode == 1) {
            $printer->setTextLineHeight(40);
            $printer->appendText(__("实收金额"), 320);
            $printer->setAlignment(ImgFont::ALIGN_RIGHT);
            $printer->appendText($this->getPriceAndUnit($data['total_amount']), 0, 20);
            $printer->setAlignment(ImgFont::ALIGN_LEFT);
            $printer->lineFeed();
            $printer->appendSplitline();
            $printer->recoverDefaultTextLineHeight();
            foreach ($data['incomes'] as $income) {
                if ($income['pay_type'] == -1) {
                    $income['pay_type_way'] = __('免单金额');
                }
                $printer->appendText($income['pay_type_way'], 320);
                $printer->setAlignment(ImgFont::ALIGN_RIGHT);
                $printer->appendText($this->getPriceAndUnit($income['price']), 0, 20);
                $printer->setAlignment(ImgFont::ALIGN_LEFT);
                $printer->lineFeed();
            }
        }
        // 按商品分类
        else if ($mode == 2) {
            $printer->setTextLineHeight(40);
            $printer->printInColumns(
                [__("分类"), 300, ImgFont::ALIGN_LEFT, 2],
                [__("数量"), 96, ImgFont::ALIGN_RIGHT, 2],
                [__("小计"), 0, ImgFont::ALIGN_RIGHT, 2],
            );
            $printer->appendSplitline(true);
            $printer->recoverDefaultTextLineHeight();
            foreach ($data['categorys'] as $category) {
                $printer->printInColumns(
                    [(new CategoryModel)->getPathNameTextAttr($category['name'], $category) . '', 300, ImgFont::ALIGN_LEFT, 2],
                    ["{$category['sales']}", 96, ImgFont::ALIGN_RIGHT, 2],
                    [$this->getPriceAndUnit($category['prices']), 0, ImgFont::ALIGN_RIGHT, 2],
                );
            }
            // 
            $printer->lineFeed(1, 10);
            $printer->appendSplitline(true);
            $printer->printInColumns(
                [__("销售笔数"), 300, ImgFont::ALIGN_LEFT],
                [$data['sales_num'], 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->recoverDefaultTextLineHeight();
            foreach ($data['incomes'] as $income) {
                if ($income['pay_type'] == -1) {
                    $income['pay_type_way'] = __('免单金额');
                }
                $printer->printInColumns(
                    [$income['pay_type_way'], 300, ImgFont::ALIGN_LEFT],
                    [$this->getPriceAndUnit($income['price']), 0, ImgFont::ALIGN_RIGHT],
                );
            }
            if ($data['refund_amount'] > 0) {
                $printer->printInColumns(
                    [__("退款金额"), 300, ImgFont::ALIGN_LEFT],
                    [$this->getPriceAndUnit($data['refund_amount']), 0, ImgFont::ALIGN_RIGHT],
                );
            }
            $printer->printInColumns(
                [__("实收金额"), 300, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['total_amount']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->setTextLineHeight(40);
        }
        // 按商品
        else if ($mode == 3) {
            $printer->setTextLineHeight(40);
            $printer->printInColumns(
                [__("商品名称"), 300, ImgFont::ALIGN_LEFT, 2],
                [__("销量"), 120, ImgFont::ALIGN_RIGHT, 2],
                [__("小计"), 0, ImgFont::ALIGN_RIGHT, 2],
            );
            $printer->appendSplitline(true);
            $printer->recoverDefaultTextLineHeight();
            foreach ($data['products'] as $key => $product) {
                $specNameText = ProductSku::getSpecNameTextAttr($product['spec_name'] ?: '', $product);
                $product_name = Product::getProductNameTextAttr($product['product_name'] ?: '', $product) . ($specNameText ? " ($specNameText)" : '');
                $printer->printInColumns(
                    [$product_name, 300, ImgFont::ALIGN_LEFT],
                    [' ' . helper::amountPermillage($product['product_price']) . '*' . "{$product['sales']}", 120, ImgFont::ALIGN_RIGHT],
                    [$this->getPriceAndUnit($product['prices']), 0, ImgFont::ALIGN_RIGHT],
                );
            }
        }
        // 全部
        else {
            $printer->recoverDefaultTextLineHeight();
            // 
            $printer->printInColumns(
                [__("总销售额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['receivable_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("原商品金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['not_tax_total_product_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("服务费"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['service_money']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("支付手续费"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['pay_fee_money']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("税费"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['consumption_tax_money']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("商品数量"), 350, ImgFont::ALIGN_LEFT],
                [$data['all']['product_num'] . '', 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("优惠折扣"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['discount_money']), 0, ImgFont::ALIGN_RIGHT],
            );
            if ($isBalance == 1 || $data['all']['user_discount_money'] > 0) {
                $printer->printInColumns(
                    [__("会员折扣"), 350, ImgFont::ALIGN_LEFT],
                    [$this->getPriceAndUnit($data['all']['user_discount_money']), 0, ImgFont::ALIGN_RIGHT],
                );
            }
            $printer->printInColumns(
                [__("退款金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['refund_money']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("免单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['free_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("实收金额"), 400, ImgFont::ALIGN_LEFT, 2, 22],
                [$this->getPriceAndUnit($data['all']['received_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            // 税收百分比对象列表
            if ($data['all']['percentage_list']) {
                $printer->appendSplitline();
            }
            foreach ($data['all']['percentage_list'] as $key => $percentage) {
                $printer->setAlignment(ImgFont::ALIGN_LEFT);
                $printer->setFontWeight(2);
                if ($this->lang == 'ja') {
                    $printer->appendText($percentage['tax_rate'] . '%' . __('的对象'));
                } else {
                    $printer->appendText('VAT (' . $percentage['tax_rate'] . '%)');
                }
                $printer->setFontWeight(1);
                $printer->lineFeed(1);
                $printer->printInColumns(
                    [__("合计"), 400, ImgFont::ALIGN_LEFT],
                    [$this->getPriceAndUnit($percentage['total_price']), 0, ImgFont::ALIGN_RIGHT],
                );
                $printer->setAlignment(ImgFont::ALIGN_RIGHT);
                if ($this->lang == 'ja') {
                    $printer->appendText("(" . __('其中消费税') . '' . $this->getPriceAndUnit($percentage['consumption_tax']) . ")");
                } else {
                    $printer->appendText("(" . __('其中VAT') . '' . $this->getPriceAndUnit($percentage['consumption_tax']) . ")");
                }
                $printer->lineFeed(1);
            }
            // 充值相关
            $printer->appendSplitline();
            $printer->lineFeed(1);
            $printer->setAlignment(ImgFont::ALIGN_CENTER);
            $printer->setFontWeight(2);
            $printer->appendText(__('会员数据'));
            $printer->setFontWeight(1);
            $printer->lineFeed(1);
            $printer->setAlignment(ImgFont::ALIGN_LEFT);
            $printer->printInColumns(
                [__("充值金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['recharge_amount']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("赠送金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['gift_money']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("赠送积分"), 350, ImgFont::ALIGN_LEFT],
                [$data['all']['gift_points'], 0, ImgFont::ALIGN_RIGHT],
            );
            // 未结账相关
            $printer->appendSplitline();
            $printer->lineFeed(1);
            $printer->setAlignment(ImgFont::ALIGN_CENTER);
            $printer->setFontWeight(2);
            $printer->appendText(__('未结账数据'));
            $printer->setFontWeight(1);
            $printer->lineFeed(1);
            $printer->setAlignment(ImgFont::ALIGN_LEFT);
            $printer->printInColumns(
                [__("订单数"), 350, ImgFont::ALIGN_LEFT],
                [$data['all']['not_settled_total_order_num'], 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['not_settled_total_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            // 合计
            $printer->appendSplitline();
            $printer->lineFeed(1);
            $printer->setAlignment(ImgFont::ALIGN_CENTER);
            $printer->setFontWeight(2);
            $printer->appendText(__('合计'));
            $printer->setFontWeight(1);
            $printer->lineFeed(1);
            $printer->setAlignment(ImgFont::ALIGN_LEFT);
            $printer->printInColumns(
                [__("所有订单数"), 350, ImgFont::ALIGN_LEFT],
                [$data['all']['total_order_num'], 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("桌数"), 350, ImgFont::ALIGN_LEFT],
                [$data['all']['total_table_num'], 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("人数"), 350, ImgFont::ALIGN_LEFT],
                [$data['all']['total_people_num'], 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("最小/大订单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['min_order_price']) . '/' . $this->getPriceAndUnit($data['all']['max_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("平均订单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['avg_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            // 桌台方式
            $printer->lineFeed(1, 12);
            $printer->setAlignment(ImgFont::ALIGN_CENTER);
            $printer->setFontWeight(2);
            $printer->appendText(__('桌台方式'));
            $printer->setFontWeight(1);
            $printer->lineFeed(1);
            $printer->printInColumns(
                [__("订单数（桌数）"), 400, ImgFont::ALIGN_LEFT],
                [$data['all']['table_order_num'], 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("人数"), 350, ImgFont::ALIGN_LEFT],
                [$data['all']['table_people_num'], 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("最小/大订单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['table_min_order_price']) . '/' . $this->getPriceAndUnit($data['all']['table_max_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("平均订单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['table_avg_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("人均"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['table_people_avg']), 0, ImgFont::ALIGN_RIGHT],
            );
            // 收银方式
            $printer->lineFeed(1, 12);
            $printer->setAlignment(ImgFont::ALIGN_CENTER);
            $printer->setFontWeight(2);
            $printer->appendText(__('点餐方式'));
            $printer->setFontWeight(1);
            $printer->lineFeed(1);
            $printer->printInColumns(
                [__("订单数"), 350, ImgFont::ALIGN_LEFT],
                [$data['all']['cashier_order_num'], 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("最小/大订单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['cashier_min_order_price']) . '/' . $this->getPriceAndUnit($data['all']['cashier_max_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            $printer->printInColumns(
                [__("平均订单金额"), 350, ImgFont::ALIGN_LEFT],
                [$this->getPriceAndUnit($data['all']['cashier_avg_order_price']), 0, ImgFont::ALIGN_RIGHT],
            );
            // 支付方式
            $printer->appendSplitline();
            $printer->lineFeed(1);
            $printer->printInColumns(
                [__("支付方式"), $this->lang == 'en' || $this->lang == 'tr'  || $this->lang == 'th' ? 230 : 270, ImgFont::ALIGN_LEFT, 2],
                [__("订单数"), $this->lang == 'en' ? 220 : 180, ImgFont::ALIGN_LEFT, 2],
                [__("金额"), 0, ImgFont::ALIGN_RIGHT, 2],
            );
            $totalPayPrice = 0;
            foreach ($data['incomes'] as $income) {
                if ($income['pay_type'] !== -1) {
                    $printer->printInColumns(
                        [$income['pay_type_way'], 300, ImgFont::ALIGN_LEFT, 2],
                        [$income['order_num'], 96, ImgFont::ALIGN_LEFT, 2],
                        [$this->getPriceAndUnit($income['price']), 0, ImgFont::ALIGN_RIGHT, 2],
                    );
                    $totalPayPrice += $income['price'];
                }
            }
            if ($totalPayPrice > 0) {
                $printer->printInColumns(
                    [__("总金额"), 300, ImgFont::ALIGN_LEFT, 2],
                    ['', 96, ImgFont::ALIGN_LEFT, 2],
                    [$this->getPriceAndUnit($totalPayPrice), 0, ImgFont::ALIGN_RIGHT, 2],
                );
            }
            // 高峰时间
            $printer->appendSplitline();
            $printer->lineFeed(1);
            $printer->printInColumns(
                [__("高峰时间"), $this->lang == 'en' || $this->lang == 'tr'  || $this->lang == 'th' ? 230 : 270, ImgFont::ALIGN_LEFT, 2],
                [__("订单数"), $this->lang == 'en' ? 220 : 180, ImgFont::ALIGN_LEFT, 2],
                [__("金额"), 0, ImgFont::ALIGN_RIGHT, 2],
            );
            foreach ($data['all']['peak_hour_list'] as $key => $peak) {
                $printer->printInColumns(
                    [$peak['time_period'], 300, ImgFont::ALIGN_LEFT, 2],
                    [$peak['num'], 96, ImgFont::ALIGN_LEFT, 2],
                    [$this->getPriceAndUnit($peak['amount']), 0, ImgFont::ALIGN_RIGHT, 2],
                );
            }
            $printer->lineFeed(1);
        }
        $printer->lineFeed(2);
        // 
        return $printer->save('', !$this->isSunmi);
    }
}
