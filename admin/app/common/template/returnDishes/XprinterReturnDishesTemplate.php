<?php

namespace app\common\template\returnDishes;

use help\DateHelp;
use app\common\template\BaseTemplate;
use app\common\enum\settings\SettingEnum;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\model\order\OrderProductReturn;
use app\common\model\settings\Setting as SettingModel;
use app\common\library\printer\party\SunmiCloudPrinter;

/**
 * 心烨 退菜单模版
 */
class XprinterReturnDishesTemplate extends BaseTemplate
{
    /**
     * 整单模版
     */
    public function completeOrder($printerConfig, $printerItem, $order)
    {
        $name = __('人');
        // 
        $printerType = $printerItem->printer['printer_type']['value'] ?? '';
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
        /**
         * 模版
         */
        $printer = new SunmiCloudPrinter(567);
        if ($printerType != PrinterTypeEnum::XPRINTER_LAN && $printerType != PrinterTypeEnum::XPRINTER_WIFI) {
            $printer->lineFeed();
        }
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
        $printer->setLineSpacing(30);
        $printer->setPrintModes(true, true, false);
        $printer->setCharacterSize(2, 2);
        $printer->appendText(__("退菜单"));
        $printer->lineFeed(2);
        if ($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
            $printer->lineFeed();
        } 
        if ($order['table_no']) {
            $printer->setLineSpacing($printerType == PrinterTypeEnum::XPRINTER_WIFI ? 120 : (preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no']) ? 80 : 60));
            $printer->appendText(__("桌号") . ": {$order['table_no']}" . $mealNum);
            $printer->setLineSpacing(30);
            $printer->lineFeed();
        } else if ($order['call_no']) {
            $printer->appendText(__("取单号") . ": {$order['call_no']}");
        }
        $printer->lineFeed();
        $printer->setLineSpacing(50);
        $printer->lineFeed();
        //
        $printer->restoreDefaultLineSpacing();
        $printer->setCharacterSize(1, 1);
        $printer->setPrintModes(false, false, false);
        $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
        $printer->setupColumns(
            [260, SunmiCloudPrinter::ALIGN_LEFT, 0],
            [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
        );
        $printer->printInColumns(__("订单号"), $order['order_no']);
        $printer->setLineSpacing(20);
        $printer->lineFeed();
        $printer->restoreDefaultLineSpacing();
        $printer->printInColumns(__("时间"), $updateTime);
        $printer->lineFeed();
        //
        if ($this->lang == 'my') {
            $printer->appendText('ကုန်စည်                                  ပမာဏ');
        } else {
            $printer->appendText(printText(__("商品"), '', __("数量"), 47));
        }
        $printer->appendText("\n------------------------------------------------\n");
        $printer->setupColumns(
            [500, SunmiCloudPrinter::ALIGN_LEFT, 0],
            [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
        );
        //
        $isPrinter = false;

        foreach ($order['product'] as $product) {
            if (!$prodcutDetail = $this->verifyPrintProductTicket($product, $printerItem)) {
                continue;
            }
            //
            $buffetText = ($buffetSignOpen && $product['is_buffet_product'] == 1) ? (__('自助餐') . ' - ') : '';
            $productName = '(' . __('退') . ') ' . $buffetText . $prodcutDetail['product_name_text'];
            $product['product_attr'] = trim($product['product_attr'] ?: '', ';'); // 先去除前后分号
            $productAttrs = explode('|-*-|-*-|', str_replace('};{', '}|-*-|-*-|{', $product['product_attr']));
            // 
            if($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
                $printer->setLineSpacing(120);
            } else {
                $printer->setLineSpacing(60);
            }
            $printer->setCharacterSize(2, 2);
            $printer->printInColumns($productName, 'x' . $product['total_num'] . '');
            $printer->setCharacterSize(1, 1);
            // 
            $printer->restoreDefaultLineSpacing();
            // 
            if ($printerType != PrinterTypeEnum::XPRINTER_LAN) {
                $printer->lineFeed();
            } 
            // 
            foreach ($productAttrs as $productAttr) {
                $printer->appendText(extractLanguage($productAttr));
                if($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
                    $printer->setLineSpacing(40);
                } else {
                    $printer->setLineSpacing(22);
                }
                $printer->lineFeed(2);
                $printer->restoreDefaultLineSpacing();
            }
            if ($product['remark']  ?? '') {
                $printer->appendText($product['remark']);
                $printer->setCharacterSize(1, 1);
                if($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
                    $printer->setLineSpacing(40);
                } else {
                    $printer->setLineSpacing(22);
                }
                $printer->lineFeed(2);
                $printer->restoreDefaultLineSpacing();
            }
            $printer->lineFeed();
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
        $printer->appendText("------------------------------------------------");
        $printer->lineFeed(1, 34);
        $printer->appendText(__("退菜原因") . '： '. $reason?->reason);
        //
        $printer->lineFeed();
        $printer->lineFeed();
        if ($printerType == PrinterTypeEnum::XPRINTER_LAN || $printerType == PrinterTypeEnum::XPRINTER_WIFI) {
            $printer->lineFeed();
        }
        // Print and exit page mode
        $printer->printAndExitPageMode();
        $printer->lineFeed(6);
        $printer->cutPaper(true);
        //
        return $printer->orderData;
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
