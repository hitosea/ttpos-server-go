<?php

namespace app\common\template\dishes;

use help\DateHelp;
use base\imgs\ImgFont;
use app\common\template\BaseTemplate;
use app\common\enum\settings\SettingEnum;
use app\common\model\settings\PrinterTemplate;
use app\common\model\settings\Setting as SettingModel;

/**
 * 图片 菜品单模版
 */
class ImgDishesTemplate extends BaseTemplate
{
    /**
     * 整单模版
     */
    public function completeOrder($printerConfig, $printerItem, $order, $products)
    {
        $name = __('人');
        // 
        $template = PrinterTemplate::getTemplate(6);
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
        if ($order['table_no']) {
            $im->setTextLineHeight(preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no']) ? 68 : 50);
            $im->appendText(__("桌号") . ": {$order['table_no']}" . $mealNum);
            $im->setTextLineHeight(45);
            $im->lineFeed(1, 20);
        } else if ($order['call_no']) {
            $im->appendText(__("取单号") . ": {$order['call_no']}\n");
        }
        $im->setFontSize(20);
        $im->setFontWeight(1);
        $im->setTextLineHeight(36);
        $im->lineFeed(2);
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
            if ($products && md5(json_encode($products)) != md5(json_encode($product))) {
                continue;
            }
            //
            $buffetText = ($buffetSignOpen && $product['is_buffet_product'] == 1) ? (__('自助餐') . '-') : '';
            $productName = $buffetText . $prodcutDetail['product_name_text'];
            $product['product_attr'] = trim($product['product_attr'] ?: '', ';'); // 先去除前后分号
            $productAttrs = explode('|-*-|-*-|', str_replace('};{', '}|-*-|-*-|{', $product['product_attr']));
            //
            if ($template == 2) {
                $im->setTextLineHeight($this->lang == 'my' ? 75 : 60);
            } else {
                $im->setTextLineHeight(50);
            }
            $im->lineFeed(1, 12);
            $im->printInColumns(
                [$productName, 500, ImgFont::ALIGN_LEFT, 2, $template == 2 ? 30 : 20],
                ['x' . $product['total_num'], 0, ImgFont::ALIGN_RIGHT, 2, $template == 2 ? 30 : 20, 42],
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
                // 判断是否缅甸语
                $im->setTextLineHeight(preg_match('/[\x{1000}-\x{109F}]/u', $product['remark']) ? 68 : 50);
                $im->lineFeed(1, 12);
                $im->setFontSize(28);
                $im->appendText($product['remark']);
                $im->lineFeed(1, 50);
                $im->setFontSize(20);
            }
            $im->lineFeed(1, 12);
            $im->setTextLineHeight(50);
            //
            $isPrinter = true;
        }
        if (!$isPrinter) {
            return "";
        }
        $im->setTextLineHeight(30);
        $im->lineFeed(4);
        // 
        return $im->save('', !$this->isSunmi);
    }

    /**
     * 一菜一单
     */
    public function oneDishOneOrder($printerConfig, $printerItem, $order, $products)
    {
        $name = __('人');
        $template = PrinterTemplate::getTemplate(4);
        $buffetSignOpen = (int)($printerConfig['buffet_sign_open'] ?? 1);
        $updateTime = date('Y/m/d H:i:s', strtotime($order['update_time']));
        // 佛历
        $settingPrinterConfig = SettingModel::getSupplierItem(SettingEnum::PRINTER, $printerItem['shop_supplier_id'], $printerItem['app_id']);
        $this->defaultCalendar = $settingPrinterConfig['default_calendar'] ?? 1;
        // 佛历
        if ($this->defaultCalendar == '3') {
            $updateTime = DateHelp::changeBuddhistCalendar($updateTime);
        }

        $mealNum = ($order['meal_num'] ? " ({$order['meal_num']}{$name})" : '');

        $isPrinter = false;

        /**
         * 模版二
         */
        if ($template == 2) {
            $im = new ImgFont(568);
            $im->setAlignment(ImgFont::ALIGN_CENTER);
            $im->setImagePadding(0);
            $im->setFontWeight(5);
            $im->setFontSize(32);
            if ($order['table_no']) {
                $im->setTextLineHeight(preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no']) ? 68 : 50);
                $im->appendText(__("桌号") . ": {$order['table_no']}" . $mealNum);
                $im->setTextLineHeight(45);
            } else if ($order['call_no']) {
                $im->appendText(__("取单号") . "： {$order['call_no']}" . $mealNum);
            }
            $im->setFontSize(20);
            $im->setFontWeight(1);
            $im->setTextLineHeight(36);
            $im->lineFeed(1, 40);
            $im->lineFeed(2);
            //
            $im->setTextLineHeight(50);
            foreach ($order['product'] as $product) {
                if (!$prodcutDetail = $this->verifyPrintProductTicket($product, $printerItem)) {
                    continue;
                }
                if ($products && md5(json_encode($products)) != md5(json_encode($product))) {
                    continue;
                }
                //
                $buffetText = ($buffetSignOpen && $product['is_buffet_product'] == 1) ? (__('自助餐') . '-') : '';
                $productName = $buffetText . $prodcutDetail['product_name_text'];
                $product['product_attr'] = trim($product['product_attr'] ?: '', ';'); // 先去除前后分号
                $productAttrs = explode('|-*-|-*-|', str_replace('};{', '}|-*-|-*-|{', $product['product_attr']));
                //
                $exportation = function ($num) use ($im, $productName, $product, $productAttrs) {
                    $im->setTextLineHeight($this->lang == 'my' ? 90 : 64);
                    $im->lineFeed(1, 12);
                    $im->printInColumns(
                        [$productName, 500, ImgFont::ALIGN_LEFT, 2, 30],
                        ['x' . $num, 0, ImgFont::ALIGN_RIGHT, 2, 30, 50],
                    );
                    if ($this->lang == 'my') {
                        $im->lineFeed(1, 12);
                    }
                    foreach ($productAttrs as $productAttr) {
                        $im->setFontSize(24);
                        $im->setTextLineHeight($this->lang == 'my' ? 50 : 40);
                        $im->appendText(extractLanguage($productAttr));
                        $im->setFontSize(20);
                        $im->lineFeed(1, 50);
                    }
                    if ($product['remark']  ?? '') {
                        // 判断是否缅甸语
                        $im->setTextLineHeight(preg_match('/[\x{1000}-\x{109F}]/u', $product['remark']) ? 68 : 50);
                        $im->lineFeed(1, 12);
                        $im->setFontSize(28);
                        $im->appendText($product['remark']);
                        $im->lineFeed(1, 50);
                        $im->setFontSize(20);
                    }
                    $im->lineFeed(1, 12);
                    $im->setTextLineHeight(50);
                };
                //
                if (($printerItem['print_select'] ?? 1) == 2) {
                    for ($i = 0; $i < $product['total_num']; $i++) {
                        $exportation(1);
                    }
                } else {
                    $exportation($product['total_num']);
                }
                //
                $isPrinter = true;
            }
            //
            $im->lineFeed(1, 40);
            $im->setTextLineHeight(27);
            $im->appendSplitline(true);
            $im->setFontSize(20);
            $im->setAlignment(ImgFont::ALIGN_CENTER);
            $im->appendText($updateTime);
            $im->lineFeed(1, 70);
        }
        /**
         * 模版一
         */
        else {
            $im = new ImgFont(568);
            $im->setAlignment(ImgFont::ALIGN_LEFT);
            $im->setImagePadding(0);
            $im->setFontWeight(5);
            $im->setFontSize(32);
            if ($order['table_no']) {
                $im->setTextLineHeight(preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no']) ? 68 : 50);
                $im->appendText(__("桌号") . ": {$order['table_no']}" . $mealNum);
                $im->setTextLineHeight(45);
            } else if ($order['call_no']) {
                $im->appendText(__("取单号") . "： {$order['call_no']}" . $mealNum);
            }
            $im->lineFeed(1);
            $im->setFontSize(20);
            $im->appendText($updateTime);
            $im->lineFeed(1);
            $im->lineFeed(1, 24);
            //
            $im->setTextLineHeight(50);
            foreach ($order['product'] as $product) {
                if (!$prodcutDetail = $this->verifyPrintProductTicket($product, $printerItem)) {
                    continue;
                }
                if ($products && md5(json_encode($products)) != md5(json_encode($product))) {
                    continue;
                }
                //
                $buffetText = ($buffetSignOpen && $product['is_buffet_product'] == 1) ? (__('自助餐') . '-') : '';
                $productName = $buffetText . $prodcutDetail['product_name_text'];
                $product['product_attr'] = trim($product['product_attr'] ?: '', ';'); // 先去除前后分号
                $productAttrs = explode('|-*-|-*-|', str_replace('};{', '}|-*-|-*-|{', $product['product_attr']));
                //
                $exportation = function ($num) use ($im, $productName, $product, $productAttrs) {
                    $im->setTextLineHeight($this->lang == 'my' ? 90 : 64);
                    $im->lineFeed(1, 12);
                    $im->printInColumns(
                        [$productName, 500, ImgFont::ALIGN_LEFT, 2, 30],
                        ['x' . $num, 0, ImgFont::ALIGN_RIGHT, 2, 30, 50],
                    );
                    if ($this->lang == 'my') {
                        $im->lineFeed(1, 12);
                    }
                    foreach ($productAttrs as $productAttr) {
                        $im->setFontSize(24);
                        $im->setTextLineHeight($this->lang == 'my' ? 50 : 40);
                        $im->appendText(extractLanguage($productAttr));
                        $im->lineFeed(1, 50);
                        $im->setFontSize(20);
                    }
                    if ($product['remark']  ?? '') {
                        // 判断是否缅甸语
                        $im->setTextLineHeight(preg_match('/[\x{1000}-\x{109F}]/u', $product['remark']) ? 68 : 50);
                        $im->lineFeed(1, 12);
                        $im->setFontSize(28);
                        $im->appendText($product['remark']);
                        $im->lineFeed(1, 50);
                        $im->setFontSize(20);
                    }
                    $im->lineFeed(1, 12);
                    $im->setTextLineHeight(50);
                };
                //
                if (($printerItem['print_select'] ?? 1) == 2) {
                    for ($i = 0; $i < $product['total_num']; $i++) {
                        $exportation(1);
                    }
                } else {
                    $exportation($product['total_num']);
                }
                //
                $isPrinter = true;
            }
            $im->lineFeed(2);
        }
        //
        if (!$isPrinter) {
            return "";
        }
        // Print and exit page mode
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
