<?php

namespace app\common\template\returnDishes;

use help\DateHelp;
use base\imgs\ImgFont;
use app\common\template\BaseTemplate;
use app\common\enum\settings\SettingEnum;
use app\common\model\order\OrderProductReturn;
use app\common\model\settings\Setting as SettingModel;

/**
 * 图片 退菜单模版
 */
class ImgReturnDishesTemplate extends BaseTemplate
{
    /**
     * 整单模版
     */
    public function completeOrder($printerConfig, $printerItem, $order)
    {
        $name = __('人');
        // 
        $template =2;
        $buffetSignOpen = (int)($printerConfig['buffet_sign_open'] ?? 1);
        $updateTime = date('Y/m/d H:i:s', strtotime($order['update_time']));
        $settingPrinterConfig = SettingModel::getSupplierItem(SettingEnum::PRINTER, $printerItem['shop_supplier_id'], $printerItem['app_id']);
        $this->defaultCalendar = $settingPrinterConfig['default_calendar'] ?? 1;
        // 佛历
        if ($this->defaultCalendar == '3') {
            $updateTime = DateHelp::changeBuddhistCalendar($updateTime);
        }
        // 
        $mealNum = ($order['meal_num'] ? " ({$order['meal_num']}{$name})" : '');
        //
        $im = new ImgFont(568);
        $im->setAlignment(ImgFont::ALIGN_CENTER);
        $im->setImagePadding(0);
        $im->setFontWeight(5);
        $im->setFontSize(30);
        $im->appendText(__("退菜单"));
        $im->lineFeed(1, 68);
        if ($order['table_no']) {
            $im->setTextLineHeight(preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no']) ? 68 : 50);
            $im->appendText(__("桌号") . ": {$order['table_no']}" . $mealNum);
            $im->setTextLineHeight(45);
            $im->lineFeed();
        } else if ($order['call_no']) {
            $im->appendText(__("取单号") . ": {$order['call_no']}\n");
        }
        $im->setFontSize(20);
        $im->setFontWeight(1);
        $im->setTextLineHeight(36);
        $im->lineFeed(2, 32);
        $im->printInColumns(
            [__("订单号"), 280, ImgFont::ALIGN_LEFT],
            [$order['order_no'], 0, ImgFont::ALIGN_RIGHT],
        );
        $im->printInColumns(
            [__("时间"), 280, ImgFont::ALIGN_LEFT],
            [$updateTime, 0, ImgFont::ALIGN_RIGHT],
        );
        $im->lineFeed(1);
        $im->printInColumns(
            [__("商品"), 280, ImgFont::ALIGN_LEFT],
            [__("数量"), 0, ImgFont::ALIGN_RIGHT],
        );
        $im->appendSplitline();
        $im->lineFeed();
        //
        $im->setTextLineHeight($template == 2 ? 50 : 40);
        $isPrinter = false;
        foreach ($order['product'] as $product) {
            if (!$prodcutDetail = $this->verifyPrintProductTicket($product, $printerItem)) {
                continue;
            }
            //
            $buffetText = ($buffetSignOpen && $product['is_buffet_product'] == 1) ? (__('自助餐') . '-') : '';
            $productName = '(' . __('退') . ') ' . $buffetText . $prodcutDetail['product_name_text'];
            $product['product_attr'] = trim($product['product_attr'] ?: '', ';'); // 先去除前后分号
            $productAttrs = explode('|-*-|-*-|', str_replace('};{', '}|-*-|-*-|{', $product['product_attr']));
            //
            if ($template == 2) {
                $im->setTextLineHeight($this->lang == 'my' ? 90 : 64);
            } else {
                $im->setTextLineHeight(50);
            }
            $im->lineFeed(1, 12);
            $im->printInColumns(
                [$productName, 500, ImgFont::ALIGN_LEFT, 2, $template == 2 ? 30 : 20],
                ['x' . $product['total_num'], 0, ImgFont::ALIGN_RIGHT, 2, $template == 2 ? 30 : 20, 50],
            );
            if ($this->lang == 'my') {
                $im->lineFeed(1, 12);
            }
            foreach ($productAttrs as $productAttr) {
                $im->setTextLineHeight($this->lang == 'my' ? 50 : 40);
                $im->appendText(extractLanguage($productAttr));
                $im->lineFeed(1, 50);
            }
            if ($product['remark']  ?? '') {
                $im->setTextLineHeight(preg_match('/[\x{1000}-\x{109F}]/u', $product['remark']) ? 50 : 40);
                $im->appendText($product['remark']);
                $im->lineFeed(1, 50);
            }
            $im->lineFeed(1, 12);
            $im->setTextLineHeight(50);
            //
            $isPrinter = true;
        }
        if (!$isPrinter) {
            return "";
        }
        // 退菜原因
        $reason = (new OrderProductReturn([], $order['app_id']))
            ->where('order_id', $order['order_id'])
            ->where('order_product_id', $product['order_product_id'])
            ->find();
        $im->appendSplitline();
        $im->lineFeed(1, 34);
        $im->appendText(__("退菜原因") . '： '. $reason?->reason);
     
        $im->setTextLineHeight(30);
        $im->lineFeed(4);   
        // 
        return $im->save('', !$this->isSunmi);
    }


    /**
     * 判断分类
     */
    private function verifyPrintProductTicket($orderProduct, $printing)
    {
        $prodcutDetail = $this->allSourceProductList[$orderProduct['product_id']] ?? [];
        if (empty($prodcutDetail)) {
            return false;
        }
        // 不存在打印规则中
        if (!in_array($orderProduct['product_id'], array_column($printing['product_ids'], 'product_id'))) {
            return false;
        }
        return $prodcutDetail;
    }
}
