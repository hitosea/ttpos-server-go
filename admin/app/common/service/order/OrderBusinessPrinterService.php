<?php

namespace app\common\service\order;

use app\common\model\shop\BindRecord;
use app\common\enum\settings\SettingEnum;
use app\common\model\settings\PrinterLog;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\template\business\ImgBusinessTemplate;
use app\common\model\settings\Setting as SettingModel;
use app\common\template\business\SunmiBusinessTemplate;
use app\common\template\business\XprinterBusinessTemplate;

/**
 * 营业数据打印服务类
 */
class OrderBusinessPrinterService
{
    protected $error;

    /**
     * 收银打印
     */
    public function cashierPrint($deviceId, $printerConfig, $data, $param = [])
    {
        $printerInfo = SettingModel::getPrinterInfo($printerConfig, $deviceId);
        $printer = $printerInfo['printer'];
        $printerId = $printerInfo['printerId'];
        $isCashierPrinter = $printerInfo['isCashierPrinter'];
        $isCashierOpen = $printerInfo['isCashierOpen'];
        $cashierBindKey = $printerInfo['cashierBindKey'];
        // 主动点击打印-但未开启打印
        if (!$isCashierOpen) {
            $this->error = '未开启打印, 请联系管理员';
            return false;
        }
        // 主动点击打印-但未配置打印机
        if (!$printer) {
            $this->error = '未配置打印机, 请联系管理员';
            return false;
        }
        // 如果是云打印, cashierBindKey 等于自己
        if ($deviceId && !$isCashierPrinter && env('IS_CLOUD_DEPLOY', false)) {
            $cashierBindKey = $deviceId;
        }
        //
        if ($isCashierOpen) {
            //
            $printerData = PrinterLog::addPrinterLog($printer, [
                "printer_id" => $printerId,
                "cashier_bind_key" => $cashierBindKey,
                "app_id" => $data['supplier']['app_id'],
                "shop_supplier_id" => $data['supplier']['shop_supplier_id'],
                "data_type" => PrinterLog::DATA_TYPE[6]['value'],
                "data" => $printer ? $this->getPrintContent($printer, $data, $param) : '',
                "type" => $isCashierPrinter ? 1 : 0,
                "first_execution" => 1,
            ], $deviceId);
            //
            if (!$printerData) {
                $this->error = '打印失败，未连接打印机';
                return false;
            }
            return $printerData;
        }
        //
        return true;
    }

    /**
     * 构建订单打印的内容
     */
    private function getPrintContent($printers, $data, $param)
    {
        $mode = $param['mode'] ?? 2;
        $startTime = date('Y/m/d H:i:s', $data['times'][0]);
        $endTime = $data['times'][1] ? date('Y/m/d H:i:s', $data['times'][1]) : date('Y/m/d H:i:s');
        $printerType = $printers['printer_type']['value'] ?? '';
        $setting = SettingModel::getAll($data['supplier']['app_id'], $data['supplier']['shop_supplier_id']);
        $currency = $setting[SettingEnum::STORE]['values'];
        $shopName = $currency['name'] ?? $data['supplier']['name'];
        
        // 是否商米打印机
        $isSunmi = in_array($printers, BindRecord::BRANDS_SUNMI_ALL_PRINTS) || in_array($printerType, [PrinterTypeEnum::SUNMI_LAN, PrinterTypeEnum::SUNMI_CLOUD]);

        /* *
        * 图片打印
        */
        if (($setting[SettingEnum::PRINTER]['values']['print_method'] ?? 1) == 2) {
            return (new ImgBusinessTemplate($setting, null, $isSunmi))->create($data, $printerType, $shopName, $mode, $startTime, $endTime);
        }
        /* *
        *商米打印机
        */
        if ($isSunmi) {
            return (new SunmiBusinessTemplate($setting, null, $isSunmi))->create($data, $printerType, $shopName, $mode, $startTime, $endTime);
        }
        /* *
        * 芯烨打印机, Compax 收银打印机 80mm 自带
        */
        if ($printers == BindRecord::BRAND_A1_1510P || in_array($printerType, [PrinterTypeEnum::XPRINTER_LAN, PrinterTypeEnum::XPRINTER_WIFI, PrinterTypeEnum::CODESOFT_LAN, PrinterTypeEnum::CODESOFT_WIFI])) {
            return (new XprinterBusinessTemplate($setting))->create($printers, $data, $printerType, $shopName, $mode, $startTime, $endTime);
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
