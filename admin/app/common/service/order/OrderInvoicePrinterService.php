<?php

namespace app\common\service\order;

use app\common\model\shop\BindRecord;
use app\common\enum\settings\SettingEnum;
use app\common\enum\settings\PrinterTypeEnum;
use app\shop\model_old\order\Order as OrderModel;
use app\common\template\invoice\ImgInvoiceTemplate;
use app\common\template\invoice\SunmiInvoiceTemplate;
use app\common\model\settings\Setting as SettingModel;
use app\common\template\invoice\CompaxInvoiceTemplate;
use app\common\template\invoice\CodesoftInvoiceTemplate;
use app\common\template\invoice\XprinterInvoiceTemplate;

/**
 * 发票打印服务类
 */
class OrderInvoicePrinterService
{
    protected $param = [];
    //
    protected $order = null;
    //
    protected $error;

    /**
     * 收银打印
     */
    public function cashierPrint($param = [])
    {
        $this->param = $param;
        // 
        $printerConfig = SettingModel::getSupplierItem(SettingEnum::PRINTER, $this->param['shop_supplier_id'] ?? 0, $this->param['app_id'] ?? 0);
        if ($printLang = ($param['print_lang'] ?? '')) {
            request()->language = $printLang;
        } else {
            request()->language = $printerConfig['default_language'] ?? '';
        }
        //
        $this->order = OrderModel::detail($param['order_id'] ?? 0);
        if (!$this->order) {
            $this->error = '订单不存在';
            request()->language = '';
            return false;
        }
        // 保存开票信息
        if (isset($param['invoice_info'])) {
            $this->order->invoiceInfo()->save([
                'company_name' => $param['invoice_info']['company_name'] ?? '',
                'company_addr' => $param['invoice_info']['company_addr'] ?? '',
                'company_tax_number' => $param['invoice_info']['company_tax_number'] ?? '',
                'company_phone' => $param['invoice_info']['company_phone'] ?? '',
                'remark' => $param['invoice_info']['remark'] ?? '',
            ]);
        } else {
            $this->order->invoiceInfo = null;
        }
        //
        $this->order->print_num = $this->order->print_num + 1;
        $this->order->save();
        //
        $content = $this->getPrintContent("", false);
        request()->language = '';
        // 
        return $content;
    }

    /**
     * 构建打印的内容
     */
    private function getPrintContent($printType, $isCashierPrinter)
    {
        $setting = SettingModel::getAll($this->param['app_id'], $this->param['shop_supplier_id']);

        $setting[SettingEnum::PRINTER]['values']['print_method'] = 2;

        // 是否商米打印机
        $isSunmi = in_array($printType, BindRecord::BRANDS_SUNMI_ALL_PRINTS) || in_array($printType, [PrinterTypeEnum::SUNMI_LAN, PrinterTypeEnum::SUNMI_CLOUD]);

        /* *
        * 图片打印
        */
        if (($setting[SettingEnum::PRINTER]['values']['print_method'] ?? 1) == 2) {
            return (new ImgInvoiceTemplate($setting, null, $isSunmi))->create($this->order, $printType);
        }
        /* *
        * 商米收银自带打印机
        */
        if ($isSunmi) {
            return (new SunmiInvoiceTemplate($setting, null, $isSunmi))->create($this->order, $printType);
        }
        /* *
        * Compax 收银打印机 80mm 自带
        */
        if ($printType == BindRecord::BRAND_A1_1510P) {
            return (new CompaxInvoiceTemplate($setting))->create($this->order, $printType, $isCashierPrinter);
        }
        /* *
        * 芯烨打印机
        */
        if (in_array($printType, [PrinterTypeEnum::XPRINTER_LAN, PrinterTypeEnum::XPRINTER_WIFI])) {
            return (new XprinterInvoiceTemplate($setting))->create($this->order, $printType, $isCashierPrinter);
        }
        /* *
        * Codesoft打印机
        */
        if (in_array($printType, [PrinterTypeEnum::CODESOFT_LAN, PrinterTypeEnum::CODESOFT_WIFI])) {
            return (new CodesoftInvoiceTemplate($setting))->create($this->order, $printType, $isCashierPrinter);
        }
        //
        return "";
    }

    /**
     * 获取错误
     */
    public function getError()
    {
        return $this->error;
    }
}
