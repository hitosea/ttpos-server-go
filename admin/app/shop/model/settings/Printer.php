<?php

namespace app\shop\model\settings;

use app\common\enum\settings\SettingEnum;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\model\settings\Printer as PrinterModel;
use app\common\model\supplier\Printing as PrintingModel;

class Printer extends PrinterModel
{
    /**
     * 添加新记录
     */
    public function add($data)
    {
        $licenses = request()->licenses;
        $pl = ($licenses['p_l'] ?? 0);
        if ($pl != -1 && $this->where('shop_supplier_id', '=', $data['shop_supplier_id'])->where('is_delete', '=', 0)->count() >= $pl) {
            $this->error = '打印机数量已达上限，如有需要，请联系销售代表';
            return false;
        }
        if ($this->where('printer_name', '=', $data['printer_name'])->where('is_delete', '=', 0)->count()) {
            $this->error = '打印机名称已存在';
            return false;
        }
        $printerType = $data['printer_type'];
        if ($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
            $printerType = PrinterTypeEnum::XPRINTER_LAN;
        }
        if ($printerType == PrinterTypeEnum::CODESOFT_WIFI) {
            $printerType = PrinterTypeEnum::CODESOFT_LAN;
        }
        $data['printer_config'] = json_encode($data[$printerType]);
        $data['app_id'] = self::$app_id;
        return $this->save($data);
    }

    /**
     * 编辑记录
     */
    public function edit($data)
    {
        if ($this->where('printer_name', '=', $data['printer_name'])->where('is_delete', '=', 0)->where('printer_id', '<>', $data['printer_id'])->count()) {
            $this->error = '打印机名称已存在';
            return false;
        }
        $printerType = $data['printer_type'];
        if ($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
            $printerType = PrinterTypeEnum::XPRINTER_LAN;
        }
        if ($printerType == PrinterTypeEnum::CODESOFT_WIFI) {
            $printerType = PrinterTypeEnum::CODESOFT_LAN;
        }
        $data['printer_config'] = json_encode($data[$printerType]);
        return $this->save($data);
    }

    /**
     * 删除记录
     * @return bool|int
     */
    public function setDelete()
    {
        $printer = Setting::getSupplierItem(SettingEnum::PRINTER, $this->shop_supplier_id);
        $printerIds = array_column($printer['cashier_printer'] ?? [], 'printer_id');
        if (in_array($this->printer_id, $printerIds)) {
            $this->error = '该打印机已被使用，无法删除';
            return false;
        }
        if (PrintingModel::where('shop_supplier_id', '=', $this->shop_supplier_id)->like('printer_id', '"' . $this->printer_id . '"')->where('is_delete', '0')->find()) {
            $this->error = '该打印机已被使用，无法删除';
            return false;
        }
        return $this->save(['is_delete' => 1]);
    }
}
