<?php

namespace app\common\service\order;

use app\common\model\shop\BindRecord;
use app\common\enum\settings\SettingEnum;
use app\common\model\settings\PrinterLog;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\template\handover\ImgHandoverTemplate;
use app\common\model\settings\Setting as SettingModel;
use app\common\template\handover\SunmiHandoverTemplate;
use app\common\template\handover\CompaxHandoverTemplate;
use app\common\template\handover\CodesoftHandoverTemplate;
use app\common\template\handover\XprinterHandoverTemplate;

/**
 * 交班数据打印服务类
 */
class OrderHandoverPrinterService
{
    /**
     * 收银打印
     */
    public function cashierPrint($data, $deviceId, $isPrePrint = true)
    {
        $printerConfig = SettingModel::getSupplierItem('printer', $data['shop_supplier_id'], $data['app_id']);
        $printerInfo = SettingModel::getPrinterInfo($printerConfig, $deviceId);
        $printer = $printerInfo['printer'];
        $printerId = $printerInfo['printerId'];
        $isCashierPrinter = $printerInfo['isCashierPrinter'];
        $isCashierOpen = $printerInfo['isCashierOpen'];
        $cashierBindKey = $printerInfo['cashierBindKey'];
        // 如果是云打印, cashierBindKey 等于自己
        if ($deviceId && !$isCashierPrinter && env('IS_CLOUD_DEPLOY', false)) {
            $cashierBindKey = $deviceId;
        }
        //
        if ($isCashierOpen) {
            return PrinterLog::addPrinterLog($printer, [
                "printer_id" => $printerId,
                "cashier_bind_key" => $cashierBindKey,
                "app_id" => $data['app_id'],
                "shop_supplier_id" => $data['shop_supplier_id'],
                "data_type" => PrinterLog::DATA_TYPE[7]['value'],
                "data" => $printer ? $this->getPrintContent($printer, $data, $isPrePrint) : '',
                "type" => $isCashierPrinter ? 1 : 0,
                "first_execution" => 1,
            ], $deviceId);
        }
        //
        return false;
    }

    /**
     * 构建交班单打印的内容
     */
    private function getPrintContent($printers, $data, $isPrePrint)
    {
        $setting = SettingModel::getAll($data['app_id'], $data['shop_supplier_id']);
        $currency = $setting[SettingEnum::STORE]['values'];
        $shopName = $currency['name'] ?? $data['supplier']['name'];
        $printerType = $printers['printer_type']['value'] ?? '';

        // 是否商米打印机
        $isSunmi = in_array($printers, BindRecord::BRANDS_SUNMI_ALL_PRINTS) || in_array($printerType, [PrinterTypeEnum::SUNMI_LAN, PrinterTypeEnum::SUNMI_CLOUD]);

        /* *
        * 图片打印
        */
        if (($setting[SettingEnum::PRINTER]['values']['print_method'] ?? 1) == 2) {
            return (new ImgHandoverTemplate($setting, null, $isSunmi))->create($data, $shopName, $printerType, $isPrePrint);
        }
        /* *
        * 商米打印机
        */
        if ($isSunmi) {
            return (new SunmiHandoverTemplate($setting, null, true))->create($data, $shopName, $printerType, $isPrePrint);
        }
        /* *
        *  Compax 收银打印机 80mm 自带
        */
        if ($printers == BindRecord::BRAND_A1_1510P) {
            return (new CompaxHandoverTemplate($setting))->create($data, $shopName, $printerType, $isPrePrint);
        }
        /* *
        * 芯烨打印机
        */
        if (in_array($printerType, [PrinterTypeEnum::XPRINTER_LAN, PrinterTypeEnum::XPRINTER_WIFI])) {
            return (new XprinterHandoverTemplate($setting))->create($printerType, $data, $shopName, $isPrePrint);
        }
        /* *
        * CODESOFT打印机
        */
        if (in_array($printerType, [PrinterTypeEnum::CODESOFT_LAN, PrinterTypeEnum::CODESOFT_WIFI])) {
            return (new CodesoftHandoverTemplate($setting, null, $isSunmi))->create($printerType, $data, $shopName, $isPrePrint);
        }
        //
        return "";
    }
}
