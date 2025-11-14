<?php

namespace app\common\model\settings;

use app\common\model\BaseModel;
use app\common\enum\settings\SettingEnum;

/**
 * 打印模板
 */
class PrinterTemplate extends BaseModel
{
    protected $name = 'printer_template';
    protected $pk = 'id';

    /**
     * 详情
     */
    public static function detail($id, $with = [])
    {
        return static::with($with)->where('id', $id)->find();
    }

    /**
     * 列表
     */
    public function getList($shop_supplier_id)
    {
        // 未开启外送，不显示外送单模板
        $list = $this->when(request()->licenses['is_open_delivery'] == 0, function ($q) {
            return $q->where('uuid', '<>', 12);
        })->field('id, name, template, is_show_sku, tmp_uuid, create_time,update_time,delete_time')->select()->toArray() ?: [];
        $printerSettings = Setting::getSupplierItem(SettingEnum::PRINTER, $shop_supplier_id);
        // 授权无排队叫号权限
        foreach ($list as $key => &$item) {
            $item['print_method'] = $printerSettings['print_method'] ?? 1;
            $item['kitchen_print_method'] = $printerSettings['kitchen_print_method'] ?? 1;
        }
        return array_values($list);
    }

    /**
     * 设置模板
     */
    public function setTemplate($data)
    {
        $customizeData = '';
        PrinterCustomize::where('template_id', $this->id)->update(['is_use' => 0]);
        if ($data['tmp_uuid'] ?? 0) {
            $customizeData = PrinterCustomize::where('uuid', $data['tmp_uuid'])->value('data') ?: '';
            PrinterCustomize::where('uuid', $data['tmp_uuid'])->update(['is_use' => 1]);
        }
        return $this->save([
            'template' => $data['template'] ?? 1,
            'is_show_sku' => $data['is_show_sku'] ?? 1,
            'tmp_uuid' => $data['tmp_uuid'] ?? 0,
            'tmp_data' => $customizeData,
        ]);
    }

    /**
     * 获取模板
     */
    public static function getTemplate($id)
    {
        return self::where('id', '=', $id)->value('template') ?: 1;
    }
}
