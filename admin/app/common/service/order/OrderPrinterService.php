<?php

namespace app\common\service\order;

use app\common\tasks\Task;
use app\common\tasks\DishesTask;
use app\common\model\order\Order;
use app\common\tasks\ImgBillTask;
use app\common\model\shop\BindRecord;
use app\common\enum\settings\SettingEnum;
use app\common\model\settings\PrinterLog;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\template\bill\ImgBillTemplate;
use app\common\model\store\Table as TableModel;
use app\common\template\bill\SunmiBillTemplate;
use app\common\template\bill\CompaxBillTemplate;
use app\common\template\dishes\ImgDishesTemplate;
use app\common\template\bill\CodesoftBillTemplate;
use app\common\template\bill\XprinterBillTemplate;
use app\common\model\product\Product as ProductModel;
use app\common\model\settings\Setting as SettingModel;
use app\common\template\dishes\CodesoftDishesTemplate;
use app\common\template\dishes\XprinterDishesTemplate;
use app\common\model\supplier\Printing as PrintingModel;
use app\common\template\returnDishes\ImgReturnDishesTemplate;
use app\common\template\returnDishes\XprinterReturnDishesTemplate;

/**
 * 订单打印服务类
 */
class OrderPrinterService
{
    protected $setting;
    protected $error;
    protected $allSourceProductList = [];

    /**
     * 执行订单打印
     */
    public function printTicket($order, $isQueue = true, $deviceId = '', $paramData = '', $isAsyn = true, $isPrePrint = true)
    {
        $this->setting = SettingModel::getAll($order['app_id'], $order['shop_supplier_id']);
        // 设置时区
        if ($timezone = ($this->setting[SettingEnum::STORE]['values']['time_zone'] ?? '')) {
            date_default_timezone_set($timezone);
        }
        //
        $order['_language'] = $order['_language'] ?? request()->param('print_lang') ?: ($this->setting[SettingEnum::PRINTER]['values']['default_language'] ?? '');
        // 设置打印语言
        request()->language = $order['_language'];
        // 读取订单信息
        if ($isQueue) {
            $order = Order::detail($order['order_id']);
        }
        $this->setting[SettingEnum::PRINTER]['values']['print_method'] = 2;
        // 打印
        return $this->getPrintContent($order, null, $paramData, $isPrePrint);
    }

    /**
     * 构建结账订单打印的内容
     */
    private function getPrintContent($order, $printers = null, $paramData = '', $isPrePrint = false)
    {
        $shop = SettingModel::getSupplierItem(SettingEnum::STORE, $order['shop_supplier_id'], $order['app_id']);
        $order->_shop_name = $shop['name'] ?? $order['supplier']['name'];
        $printerType = $printers['printer_type']['value'] ?? '';

        // 是否商米打印机
        $isSunmi = in_array($printers, BindRecord::BRANDS_SUNMI_ALL_PRINTS) || in_array($printerType, [PrinterTypeEnum::SUNMI_LAN, PrinterTypeEnum::SUNMI_CLOUD]);

        /* *
        * 图片打印
        */
        if (($this->setting[SettingEnum::PRINTER]['values']['print_method'] ?? 1) == 2) {
            return (new ImgBillTemplate($this->setting, null, $isSunmi))->create($order, $paramData, $isPrePrint);
        }
        /* *
        * Compax 收银打印机 80mm 自带
        */
        if ($printers == BindRecord::BRAND_A1_1510P || $printers == BindRecord::BRAND_A1_1510P) {
            return (new CompaxBillTemplate($this->setting, null, $isSunmi))->create($order, $printerType, $isPrePrint);
        }
        /* *
        * 芯烨打印机
        */
        if (in_array($printerType, [PrinterTypeEnum::XPRINTER_LAN, PrinterTypeEnum::XPRINTER_WIFI])) {
            return (new XprinterBillTemplate($this->setting, null, $isSunmi))->create($order, $printerType, $isPrePrint);
        }
        /* *
        * 商米打印机
        */
        if ($isSunmi) {
            return (new SunmiBillTemplate($this->setting, null, $isSunmi))->create($order, $printerType, $isPrePrint);
        }
        /* *
        * CODESOFT 打印机
        */
        if (in_array($printerType, [ PrinterTypeEnum::CODESOFT_LAN, PrinterTypeEnum::CODESOFT_WIFI])) {
            return (new CodesoftBillTemplate($this->setting, null, $isSunmi))->create($order, $printerType, $isPrePrint);
        }
    }

    /**
     * 菜品打印
     * @param $order 订单
     * @param string $printType 打印类型 0-为退菜打印 10-付款打印 20-下单打印 30-送厨打印
     * @param string $isAsyn 打印机名称
     * @return bool 打印是否成功
     */
    public function printProductTicket($order, $printType, $isAsyn=true)
    {
        // 暂时没有下单打印
        if ($printType == PrintingModel::PRINT_TYPE_ADD_ORDER) {
            return true;
        }
        //
        $this->setting = SettingModel::getAll($order['app_id'], $order['shop_supplier_id']);
        // 设置时区
        if ($timezone = ($this->setting[SettingEnum::STORE]['values']['time_zone'] ?? '')) {
            date_default_timezone_set($timezone);
        }
        // 打印机设置
        $printerConfig = $this->setting[SettingEnum::PRINTER]['values'] ?? [];
        // 设置打印语言
        request()->language = $printerConfig['kitchen_language'] ?? $printerConfig['default_language'] ?? '';
        // 格式化数据
        if (!is_array($order)) {
            $products = [];
            foreach ($order['product'] ?? [] as $key => $p) {
                if (!is_array($p) && (($p['is_return'] ?? 0) != 1) || $printType == PrintingModel::PRINT_TYPE_BACK_FOOD) {
                    $productAttr = $p->getData('product_attr') ?? $p['product_attr'];
                    $p = $p->toArray();
                    $p['product_attr'] = $productAttr;
                    $products[] = $p;
                }
            }
            $order = $order->hidden([])->toArray();
            $order['product'] = $products;
        }
        // 如果是图片打印模式则用异步任务执行
        if ($isAsyn && ($printerConfig['kitchen_print_method'] ?? 1) == 2) {
            Task::deliver(new DishesTask($order, $printType));
            request()->language = '';
            return true;
        }
        // 打印列表
        $printerList = PrintingModel::when($printType , function ($query) use ($printType) {
                $query->where('print_type', '=', $printType);
            })
            ->where('is_open', '=', 1)
            ->select();
        //
        if (count($printerList) > 0) {
            // 当前订单对应的区域id
            $areaId = $order['table_id'] > 0 ? TableModel::detail($order['table_id'])?->area_id : 0;
            // 原始产品列表
            $allProductIds = [];
            foreach ($order['product'] ?? [] as $p) {
                $allProductIds[] = $p['product_id'];
            }
            $allProductIds = array_values(array_unique($allProductIds));
            $allProductList = count($allProductIds) > 0 ? ProductModel::with(['category'])
                ->where('product_id', 'in', $allProductIds)
                ->select()
                ->toArray() : [];
            $this->allSourceProductList = array_column($allProductList, null, 'product_id');
            //
            foreach ($printerList as $printerItem) {
                // 小票打印才走
                if ($printerItem['type'] != 10) {
                    continue;
                }
                // 区域对的上才走
                if ($printerItem['area_id'] && !in_array($areaId, $printerItem['area_id'])) {
                    continue;
                }
                //
                foreach ($printerItem->printers() as $printer) {
                    // 获取当前的打印机
                    if ($printer->delete_time) {
                        continue;
                    }
                    $printerItem->printer = $printer;
                    // 打印数据
                    $printerRules = $printerItem->toArray();
                    if (isset($printerRules['product_ids'])) {
                        $printerRules['product_ids'] = array_column($printerRules['product_ids'], 'product_id');
                    }
                    $printerData = [
                        "printer_id" => $printer->printer_id,
                        "cashier_bind_key" => $order['settle_device_id'],
                        "app_id" => $order['app_id'],
                        "shop_supplier_id" => $order['shop_supplier_id'],
                        'order_id' => $order['order_id'],
                        "printer_rule_id" => $printerItem->id,
                        "printer_rule" => json_encode($printerRules, JSON_UNESCAPED_UNICODE),
                    ];
                    // 退菜单打印
                    if ($printType == PrintingModel::PRINT_TYPE_BACK_FOOD) {
                        $data = $this->getPrintReturnProductContent($printerConfig, $printerItem, $order);
                        if ($data) {
                            PrinterLog::addPrinterLog($printer, array_merge($printerData, [
                                "data" => $data,
                                "data_type" => PrinterLog::DATA_TYPE[9]['value'],
                            ]));
                        }
                    }
                    // 送厨单打印
                    else {
                        // 10整单打印  20按商品分类打印 30按标签打印 40一菜一單打印
                        $isCompleteOrderPrinter = $printerItem['print_method'] == 10;
                        // 一菜一单打印
                        if (!$isCompleteOrderPrinter) {
                            foreach ($order['product'] as $product) {
                                if (!$this->verifyPrintProductTicket($product, $printerItem)) {
                                    continue;
                                }
                                if ($printerItem['print_method'] == 40 || ($printerItem['is_open_one_food'] ?? 0) == 1) {
                                    $data = $this->getPrintProductOneContent($printerConfig, $printerItem, $order, $product);
                                    if ($data) {
                                        PrinterLog::addPrinterLog($printer, array_merge($printerData, [
                                            "data" => $data,
                                            "data_type" => PrinterLog::DATA_TYPE[3]['value'],
                                        ]));
                                    }
                                } else {
                                    $isCompleteOrderPrinter = true;
                                }
                            }
                        }
                        // 整单打印
                        if ($isCompleteOrderPrinter) {
                            $data = $this->getPrintProductContent($printerConfig, $printerItem, $order);
                            if ($data) {
                                PrinterLog::addPrinterLog($printer, array_merge($printerData, [
                                    "data" => $data,
                                    "data_type" => PrinterLog::DATA_TYPE[4]['value'],
                                ]));
                            }
                        }
                    }
                }
            }
        }
        request()->language = '';
        return true;
    }

    /**
     * 构建（退菜单）打印的内容
     */
    private function getPrintReturnProductContent($printerConfig, $printerItem, $order = null)
    {
        if (($printerConfig['kitchen_print_method'] ?? 1) == 2) {
            return (new ImgReturnDishesTemplate(null, $this->allSourceProductList))->completeOrder($printerConfig, $printerItem, $order);
        }
        /* *
        *商米 和 芯烨 打印机
        */
        if ($printerItem->printer) {
            return (new XprinterReturnDishesTemplate(null, $this->allSourceProductList))->completeOrder($printerConfig, $printerItem, $order);
        }
        return "";
    }

    /**
     * 构建订单菜品（整单）打印的内容
     */
    private function getPrintProductContent($printerConfig, $printerItem, $order = null, $products = null)
    {
        $printerType = $printerItem->printer['printer_type']['value'] ?? '';
        //
        if (($printerConfig['kitchen_print_method'] ?? 1) == 2) {
            return (new ImgDishesTemplate(null, $this->allSourceProductList))->completeOrder($printerConfig, $printerItem, $order, $products);
        }
        /* *
        * CODESOFT 打印机
        */
        if ($printerItem->printer && in_array($printerType, [ PrinterTypeEnum::CODESOFT_LAN, PrinterTypeEnum::CODESOFT_WIFI])) {
            return (new CodesoftDishesTemplate(null, $this->allSourceProductList))->completeOrder($printerConfig, $printerItem, $order, $products);
        }
        /* *
        *商米 和 芯烨 打印机
        */
        if ($printerItem->printer) {
            return (new XprinterDishesTemplate(null, $this->allSourceProductList))->completeOrder($printerConfig, $printerItem, $order, $products);
        }
        return "";
    }

    /**
     * 构建订单菜品（一菜一单）打印的内容
     */
    private function getPrintProductOneContent($printerConfig, $printerItem, $order, $products = null)
    {
        $printerType = $printerItem->printer['printer_type']['value'] ?? '';
        //
        if (($printerConfig['kitchen_print_method'] ?? 1) == 2) {
            return (new ImgDishesTemplate(null, $this->allSourceProductList))->oneDishOneOrder($printerConfig, $printerItem, $order, $products);
        }
        /* *
        * CODESOFT 打印机
        */
        if ($printerItem->printer && in_array($printerType, [PrinterTypeEnum::CODESOFT_LAN, PrinterTypeEnum::CODESOFT_WIFI])) {
            return (new CodesoftDishesTemplate(null, $this->allSourceProductList))->oneDishOneOrder($printerConfig, $printerItem, $order, $products);
        }
        /* *
        *商米 和 芯烨 打印机
        */
        if ($printerItem->printer) {
            return (new XprinterDishesTemplate(null, $this->allSourceProductList))->oneDishOneOrder($printerConfig, $printerItem, $order, $products);
        }
        return "";
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

    /**
     * 获取错误
     */
    public function getError()
    {
        return $this->error;
    }
}
