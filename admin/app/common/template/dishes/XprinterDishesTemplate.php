<?php

namespace app\common\template\dishes;

use help\DateHelp;
use app\common\template\BaseTemplate;
use app\common\enum\settings\SettingEnum;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\model\settings\PrinterTemplate;
use app\common\model\settings\Setting as SettingModel;
use app\common\library\printer\party\SunmiCloudPrinter;

/**
 * 心烨 菜品单模版
 */
class XprinterDishesTemplate extends BaseTemplate
{
    /**
     * 整单模版
     */
    public function completeOrder($printerConfig, $printerItem, $order, $products)
    {
        $name = __('人');
        // 
        $template = PrinterTemplate::getTemplate(6);
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
        
        // 
        $isPrinter = false;

        /**
         * 模版
         */
        $printer = new SunmiCloudPrinter(567);
        if ($printerType != PrinterTypeEnum::XPRINTER_LAN && $printerType != PrinterTypeEnum::XPRINTER_WIFI) {
            $printer->lineFeed();
        }
        
        /**
         * 模版 一
         */
        if ($template == 1) {
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setCharacterSize(2, 2);
            $printer->setPrintModes(true, true, false);
            if ($order['table_no']) {
                $printer->setLineSpacing($printerType == PrinterTypeEnum::XPRINTER_WIFI ? 120 : (preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no']) ? 80 : 60));
                $printer->appendText(__("桌号") . ": {$order['table_no']}" . $mealNum);
                $printer->restoreDefaultLineSpacing();
                $printer->lineFeed();
            } else if ($order['call_no']) {
                $printer->appendText(__("取单号") . ": {$order['call_no']}\n");
            }
            $printer->lineFeed();
            $printer->setLineSpacing(50);
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
            if ($printerType != PrinterTypeEnum::XPRINTER_WIFI) {
                $printer->setLineSpacing(40);
            }
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
            foreach ($order['product'] as $product) {
                if (!$prodcutDetail = $this->verifyPrintProductTicket($product, $printerItem)) {
                    continue;
                }
                if ($products && md5(json_encode($products)) != md5(json_encode($product))) {
                    continue;
                }
                //
                if ($printerType == PrinterTypeEnum::XPRINTER_LAN || $printerType == PrinterTypeEnum::XPRINTER_WIFI) {
                    $printer->setLineSpacing(90);
                } else {
                    $printer->setLineSpacing(45);
                }
                // 
                $buffetText = ($buffetSignOpen && $product['is_buffet_product'] == 1) ? (__('自助餐') . '-') : '';
                $productName = $buffetText . $prodcutDetail['product_name_text'];
                $product['product_attr'] = trim($product['product_attr'] ?: '', ';'); // 先去除前后分号
                $productAttrs = explode('|-*-|-*-|', str_replace('};{', '}|-*-|-*-|{', $product['product_attr']));
                $printer->printInColumns($productName, 'x' . $product['total_num'] . '');
                // 
                $printer->setCharacterSize(1, 1);
                if ($printerType == PrinterTypeEnum::XPRINTER_LAN) {
                    $printer->setLineSpacing(90);
                } else if ($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
                    $printer->setLineSpacing(10);
                    $printer->lineFeed();
                    $printer->setLineSpacing(90);
                } else {
                    $printer->setLineSpacing(45);
                }
                foreach ($productAttrs as $productAttr) {
                    $printer->appendText(extractLanguage($productAttr));
                    $printer->lineFeed();
                }
                if ($product['remark']  ?? '') {
                    $printer->setLineSpacing($this->isMyText($product['remark']) ? 85 : 55);
                    $printer->setCharacterSize(2, 2);
                    $printer->printInColumns($product['remark']);
                    $printer->setCharacterSize(1, 1);
                    $printer->setLineSpacing(20);
                    $printer->lineFeed();
                }
                //
                $printer->setLineSpacing(12);
                $printer->lineFeed();
                // 
                $isPrinter = true;
            }
        } 
        /**
         * 模版 二
         */
        else {
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setCharacterSize(2, 2);
            $printer->setPrintModes(true, true, false);
            if ($order['table_no']) {
                $printer->setLineSpacing($printerType == PrinterTypeEnum::XPRINTER_WIFI ? 120 : (preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no']) ? 80 : 60));
                $printer->appendText(__("桌号") . ": {$order['table_no']}" . $mealNum);
                $printer->restoreDefaultLineSpacing();
                $printer->lineFeed();
            } else if ($order['call_no']) {
                $printer->appendText(__("取单号") . ": {$order['call_no']}\n");
            }
            $printer->lineFeed();
            $printer->setLineSpacing(50);
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
            if ($printerType != PrinterTypeEnum::XPRINTER_WIFI) {
                $printer->setLineSpacing(40);
            }
            if ($this->lang == 'my') {
                $printer->appendText('ကုန်စည်                                  ပမာဏ');
            } else {
                $printer->appendText(printText(__("商品"), '', __("数量"), 47));
            }
            $printer->appendText("\n------------------------------------------------\n");
            //
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
                $printer->setupColumns(
                    [$this->isThText($productName) ? 450 : 480, SunmiCloudPrinter::ALIGN_LEFT, 0],
                    [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
                );
                //
                if ($printerType == PrinterTypeEnum::XPRINTER_LAN) {
                    $printer->setLineSpacing(20);
                } else if ($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
                    $printer->setLineSpacing(125);
                } else {
                    $printer->setLineSpacing($this->isMyText($productName) ? 80 : 68);
                }
                $printer->setCharacterSize(2, 2);
                $printer->printInColumns($productName, 'x' . $product['total_num'] . '');
                $printer->setCharacterSize(1, 1);
                // 
                if ($printerType == PrinterTypeEnum::XPRINTER_LAN) {
                    $printer->setLineSpacing(90);
                } else if ($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
                    $printer->setLineSpacing(20);
                    $printer->lineFeed();
                    $printer->setLineSpacing(90);
                } else {
                    $printer->setLineSpacing(50);
                }
                foreach ($productAttrs as $productAttr) {
                    $printer->appendText(extractLanguage($productAttr));
                    $printer->setLineSpacing(45);
                    $printer->lineFeed();
                }
                if ($product['remark']  ?? '') {
                    $printer->setLineSpacing($this->isMyText($product['remark']) ? 85 : 55);
                    $printer->setCharacterSize(2, 2);
                    $printer->printInColumns($product['remark']);
                    $printer->setCharacterSize(1, 1);
                    $printer->setLineSpacing(20);
                    $printer->lineFeed();
                }
                //
                $printer->setLineSpacing(12);
                $printer->lineFeed();
                //
                $isPrinter = true;
            }
        }
        // 
        $printer->restoreDefaultLineSpacing();
        // 
        if (!$isPrinter) {
            return "";
        }
        $printer->lineFeed();
        $printer->lineFeed();
        if ($printerType == PrinterTypeEnum::XPRINTER_LAN || $printerType == PrinterTypeEnum::XPRINTER_WIFI) {
            $printer->lineFeed();
        }
        // Print and exit page mode
        $printer->printAndExitPageMode();
        $printer->lineFeed(4);
        $printer->cutPaper(true);
        //
        return $printer->orderData;
    }

    /**
     * 一菜一单
     */
    public function oneDishOneOrder($printerConfig, $printerItem, $order, $products)
    {
        $name = __('人');
        $template = PrinterTemplate::getTemplate(4);
        $printerType = $printerItem->printer['printer_type']['value'] ?? '';
        $buffetSignOpen = (int)($printerConfig['buffet_sign_open'] ?? 1);
        $updateTime = date('Y/m/d H:i:s', strtotime($order['update_time']));
        $isThai =  preg_match('/[\p{Thai}]/u', __("金额"));
        // 佛历
        $settingPrinterConfig = SettingModel::getSupplierItem(SettingEnum::PRINTER, $printerItem['shop_supplier_id'], $printerItem['app_id']);
        $this->defaultCalendar = $settingPrinterConfig['default_calendar'] ?? 1;
        // 佛历
        if ($this->defaultCalendar == '3') {
            $updateTime = DateHelp::changeBuddhistCalendar($updateTime);
        }

        // 模版
        $printer = new SunmiCloudPrinter(567);
        if ($printerType != PrinterTypeEnum::XPRINTER_LAN && $printerType != PrinterTypeEnum::XPRINTER_WIFI) {
            $printer->lineFeed();
        }
        $printer->restoreDefaultLineSpacing();

        /**
         * 模版二
         */
        if ($template == 2) {
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->setPrintModes(true, true, false);
            $printer->setCharacterSize(2, 2);
            if ($order['table_no']) {
                $printer->setLineSpacing($printerType == PrinterTypeEnum::XPRINTER_WIFI ? 120 : (preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no']) ? 80 : 60));
                $printer->appendText(__("桌号") . ": {$order['table_no']}" . ($order['meal_num'] ? " ({$order['meal_num']}{$name})" : ''));
                $printer->restoreDefaultLineSpacing();
            } else if ($order['call_no']) {
                $printer->appendText(__("取单号") . ": {$order['call_no']}" . ($order['meal_num'] ? " ({$order['meal_num']}{$name})" : ''));
            }
            $printer->lineFeed();
            $printer->lineFeed();
            $printer->lineFeed();
            //
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setupColumns(
                [490, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $isPrinter = false;
            $printer->setPrintModes(false, false, false);
            foreach ($order['product'] as $product) {
                if (!$prodcutDetail = $this->verifyPrintProductTicket($product, $printerItem)) {
                    continue;
                }
                if ($products && md5(json_encode($products)) != md5(json_encode($product))) {
                    continue;
                }
                //
                $buffetText = ($buffetSignOpen && $product['is_buffet_product'] == 1) ? (__('自助餐') . ' - ') : '';
                $productName = $buffetText . $prodcutDetail['product_name_text'];
                $product['product_attr'] = trim($product['product_attr'] ?: '', ';'); // 先去除前后分号
                $productAttrs = explode('|-*-|-*-|', str_replace('};{', '}|-*-|-*-|{', $product['product_attr']));
                $exportation = function ($num) use ($printer, $productName, $product, $printerType, $productAttrs) {
                    if($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
                        $printer->setLineSpacing(120);
                    } else {
                        $printer->setLineSpacing(60);
                    }
                    $printer->setCharacterSize(2, 2);
                    $printer->printInColumns($productName, 'x' . $num . '');
                    $printer->setCharacterSize(1, 2);
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
                        if($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
                            $printer->setLineSpacing(120);
                        } else {
                            $printer->setLineSpacing(60);
                        }
                        $printer->setCharacterSize(2, 2);
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
            $printer->setCharacterSize(1, 1);
            $printer->setPrintModes(false, false, false);
            $printer->lineFeed();
            $printer->appendText("------------------------------------------------\n");
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_CENTER);
            $printer->appendText($updateTime);
        }
        /**
         * 模版一
         */
        else {
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setPrintModes(true, true, false);
            $printer->setCharacterSize(2, 2);
            if ($order['table_no']) {
                $printer->setLineSpacing($printerType == PrinterTypeEnum::XPRINTER_WIFI ? 120 : (preg_match('/[\x{1000}-\x{109F}]/u', $order['table_no']) ? 80 : 60));
                $printer->appendText(__("桌号") . ": {$order['table_no']}" . ($order['meal_num'] ? " ({$order['meal_num']}{$name})" : ''));
                $printer->restoreDefaultLineSpacing();
            } else if ($order['call_no']) {
                $printer->appendText(__("取单号") . ": {$order['call_no']}" . ($order['meal_num'] ? " ({$order['meal_num']}{$name})" : ''));
            }
            if ($printerType != PrinterTypeEnum::XPRINTER_LAN) {
                $printer->lineFeed();
            } 
            $printer->lineFeed();
            $printer->setCharacterSize(1, 1);
            $printer->setPrintModes(false, false, false);
            $printer->appendText($updateTime);
            $printer->lineFeed();
            $printer->lineFeed();
            $printer->lineFeed();
            //
            $printer->setAlignment(SunmiCloudPrinter::ALIGN_LEFT);
            $printer->setupColumns(
                [490, SunmiCloudPrinter::ALIGN_LEFT, 0],
                [0, SunmiCloudPrinter::ALIGN_RIGHT, 0],
            );
            $isPrinter = false;
            foreach ($order['product'] as $product) {
                if (!$prodcutDetail = $this->verifyPrintProductTicket($product, $printerItem)) {
                    continue;
                }
                if ($products && md5(json_encode($products)) != md5(json_encode($product))) {
                    continue;
                }
                //
                $buffetText = ($buffetSignOpen && $product['is_buffet_product'] == 1) ? (__('自助餐') . ' - ') : '';
                $productName = $buffetText . $prodcutDetail['product_name_text'];
                $product['product_attr'] = trim($product['product_attr'] ?: '', ';'); // 先去除前后分号
                $productAttrs = explode('|-*-|-*-|', str_replace('};{', '}|-*-|-*-|{', $product['product_attr']));
                $exportation = function ($num) use ($printer, $productName, $product, $printerType, $productAttrs) {
                    if($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
                        $printer->setLineSpacing(120);
                    } else {
                        $printer->setLineSpacing(60);
                    }
                    $printer->setCharacterSize(2, 2);
                    $printer->printInColumns($productName, 'x' . $num . '');
                    $printer->setCharacterSize(1, 2);
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
                        $printer->setCharacterSize(2, 2);
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
        }
        if (!$isPrinter) {
            return "";
        }
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
