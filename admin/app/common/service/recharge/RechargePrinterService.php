<?php

namespace app\common\service\recharge;

use app\common\model\shop\BindRecord;
use app\common\enum\settings\SettingEnum;
use app\common\model\settings\PrinterLog;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\template\recharge\ImgRechargeTemplate;
use app\common\model\settings\Setting as SettingModel;
use app\common\template\recharge\SunmiRechargeTemplate;
use app\common\template\recharge\CompaxRechargeTemplate;
use app\common\template\recharge\CodesoftRechargeTemplate;
use app\common\template\recharge\XprinterRechargeTemplate;

/**
 * 订单打印服务类
 */
class RechargePrinterService
{
    protected $setting;
    protected $error;
    protected $allSourceProductList = [];

    /**
     * 执行订单打印
     */
    public function printTicket($rechargeOrder, $isQueue = true, $deviceId = '', $printLang = '')
    {
        $printerConfig = SettingModel::getSupplierItem(SettingEnum::PRINTER, $rechargeOrder['shop_supplier_id'], $rechargeOrder['app_id']);
        if ($printLang) {
            request()->language = $printLang;
        } else {
            request()->language = $printerConfig['default_language'] ?? '';
        }
        //
        $printerInfo = SettingModel::getPrinterInfo($printerConfig, $deviceId ?: $rechargeOrder->device_id);
        $printer = $printerInfo['printer'];
        $printerId = $printerInfo['printerId'];
        $cashierBindKey = $printerInfo['cashierBindKey'];
        $isCashierPrinter = $printerInfo['isCashierPrinter'];
        $isCashierOpen = $printerInfo['isCashierOpen'];
        // 主动点击打印-但未开启打印
        if (!$isQueue && !$isCashierOpen) {
            $this->error = '未开启打印, 请联系管理员';
            request()->language = '';
            return false;
        }
        // 主动点击打印-但未配置打印机
        if (!$isQueue && !$printer) {
            $this->error = '未配置打印机, 请联系管理员';
            request()->language = '';
            return false;
        }
        // 如果是云打印, cashierBindKey 等于自己
        if ($deviceId && !$isCashierPrinter && env('IS_CLOUD_DEPLOY', false)) {
            $cashierBindKey = $deviceId;
        }
        //
        if ($isCashierOpen) {
            //
            $content = $printer ? $this->getPrintContent($rechargeOrder, is_string($printer) ? $printer : $printer['printer_type']['value']) : '';
            request()->language = '';
            //
            $printerData = PrinterLog::addPrinterLog($printer, [
                "printer_id" => $printerId,
                "cashier_bind_key" => $cashierBindKey,
                "app_id" => $rechargeOrder->app_id,
                "shop_supplier_id" => $rechargeOrder->shop_supplier_id,
                'order_id' => $rechargeOrder->order_id,
                "data_type" => PrinterLog::DATA_TYPE[8]['value'],
                "data" => $content,
                "type" => $isCashierPrinter ? 1 : 0,
                "first_execution" => !$isQueue ? 1 : 0,
            ], $deviceId);
            // 
            if (!$printerData) {
                $this->error = '打印失败，未连接打印机';
                return false;
            }
            return $printerData;
        }
        // 
        request()->language = '';
        //
        return true;
    }

    /**
     * 构建结账订单打印的内容
     */
    private function getPrintContent($order, $printerType)
    {
        $setting = SettingModel::getAll($order['app_id'], $order['shop_supplier_id']);

        // 是否商米打印机
        $isSunmi = in_array($printerType, BindRecord::BRANDS_SUNMI_ALL_PRINTS) || $printerType == PrinterTypeEnum::SUNMI_LAN || $printerType == PrinterTypeEnum::SUNMI_CLOUD;

        /* *
        * 图片打印
        */
        if (($setting[SettingEnum::PRINTER]['values']['print_method'] ?? 1) == 2) {
            return (new ImgRechargeTemplate($setting, null, $isSunmi))->create($order);
        }
        /* *
        * Compax 收银打印机 80mm 自带
        */
        if ($printerType == BindRecord::BRAND_A1_1510P) {
            return (new CompaxRechargeTemplate($setting, null, $isSunmi))->create($order, $printerType);
        }
        /* *
        *芯烨打印机
        */
        if (in_array($printerType, [PrinterTypeEnum::XPRINTER_LAN, PrinterTypeEnum::XPRINTER_WIFI])) {
            return (new XprinterRechargeTemplate($setting, null, $isSunmi))->create($order, $printerType);
        }
        /* *
        *商米打印机
        */
        if (in_array($printerType, BindRecord::BRANDS_SUNMI_ALL_PRINTS) || $printerType == PrinterTypeEnum::SUNMI_LAN || $printerType == PrinterTypeEnum::SUNMI_CLOUD) {
            return (new SunmiRechargeTemplate($setting, null, $isSunmi))->create($order, $printerType);
        }
        /* *
        *CODESOFT打印机
        */
        if (in_array($printerType, [PrinterTypeEnum::CODESOFT_LAN, PrinterTypeEnum::CODESOFT_WIFI])) {
            return (new CodesoftRechargeTemplate($setting, null, $isSunmi))->create($order, $printerType);
        }
    }

    /**
     * 获取错误
     */
    public function getError()
    {
        return $this->error;
    }
}
