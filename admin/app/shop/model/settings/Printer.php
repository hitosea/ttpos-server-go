<?php

namespace app\shop\model\settings;

use app\common\enum\settings\SettingEnum;
use app\common\enum\settings\PrinterTypeEnum;
use app\common\model\settings\Printer as PrinterModel;
use app\common\model\settings\PrinterType;
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
        if ($pl != -1 && $this->count() >= $pl) {
            $this->error = '打印机数量已达上限，如有需要，请联系销售代表';
            return false;
        }
        if ($this->where('name', '=', $data['printer_name'])->count()) {
            $this->error = '打印机名称已存在';
            return false;
        }
        $printerType = $data['printer_type'];
        $type = PrinterType::getPrinterTypeByKey($printerType);
        if (!$type) {
            $this->error = '打印机类型不存在';
            return false;
        }
        if ($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
            $printerType = PrinterTypeEnum::XPRINTER_LAN;
        }
        if ($printerType == PrinterTypeEnum::CODESOFT_WIFI) {
            $printerType = PrinterTypeEnum::CODESOFT_LAN;
        }
        $data['uuid'] = createUuid();
        $data['name'] = $data['printer_name'] ?? '';
        $data['copies'] = $data['print_times'] ?? 0;
        $data['width'] = $data['width'] ?? 80; // 纸张宽度，默认80mm
        $data['enable_status_check'] = $data['enable_status_check'] ?? 1; // 是否启用状态检查，默认开启
        $data['enable_sound'] = $data['enable_sound'] ?? 1; // 是否启用打印提示音，默认开启
        $data['print_speed'] = $data['print_speed'] ?? 2; // 打印速度，默认稳定模式
        $data['printer_type_uuid'] = $type['uuid'];
        $data['config_json'] = json_encode($data[$printerType], JSON_UNESCAPED_UNICODE);
        if (!$this->save($data)) {
            return false;
        }

        return true;
    }

    /**
     * 编辑记录
     */
    public function edit($data)
    {
        if ($this->where('name', '=', $data['printer_name'])->where('uuid', '<>', $data['printer_id'])->count()) {
            $this->error = '打印机名称已存在';
            return false;
        }
        $printerType = $data['printer_type'];
        $type = PrinterType::getPrinterTypeByKey($printerType);
        if (!$type) {
            $this->error = '打印机类型不存在';
            return false;
        }
        if ($printerType == PrinterTypeEnum::XPRINTER_WIFI) {
            $printerType = PrinterTypeEnum::XPRINTER_LAN;
        }
        if ($printerType == PrinterTypeEnum::CODESOFT_WIFI) {
            $printerType = PrinterTypeEnum::CODESOFT_LAN;
        }
        $data['name'] = $data['printer_name'] ?? '';
        $data['copies'] = $data['print_times'] ?? 0;
        $data['width'] = $data['width'] ?? 80; // 纸张宽度，默认80mm
        $data['enable_status_check'] = $data['enable_status_check'] ?? 1; // 是否启用状态检查，默认开启
        $data['enable_sound'] = $data['enable_sound'] ?? 1; // 是否启用打印提示音，默认开启
        $data['print_speed'] = $data['print_speed'] ?? 2; // 打印速度，默认稳定模式
        $data['printer_type_uuid'] = $type['uuid'];
        if (!empty($data[$printerType]) && $this->is_usb != 1) {
            $data['config_json'] = json_encode($data[$printerType], JSON_UNESCAPED_UNICODE);
        }

        if (!$this->save($data)) {
            return false;
        }

        return true;
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
        if (PrintingModel::like('uuid', '"' . $this->printer_id . '"')->find()) {
            $this->error = '该打印机已被使用，无法删除';
            return false;
        }
        if (!$this->delete()) {
            return false;
        }

        return true;
    }

}
