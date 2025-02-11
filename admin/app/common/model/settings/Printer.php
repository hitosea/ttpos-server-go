<?php

namespace app\common\model\settings;

use think\facade\Request;
use app\common\model\BaseModel;
use app\common\model\shop\BindRecord;
use app\common\model\supplier\Printing;
use app\common\enum\settings\SettingEnum;
use app\common\enum\settings\PrinterTypeEnum;

/**
 * 打印机模型
 */
class Printer extends BaseModel
{
    protected $name = 'printer';
    protected $pk = 'id';

    /**
     * 追加属性
     */
    protected $append = ['printer_id', 'printer_name'];

    /**
     * 兼容ID字段
     */
    public function getPrinterIdAttr()
    {
        return $this->uuid ?? 0;
    }

    /**
     * 兼容字段
     */
    public function getPrinterNameAttr()
    {
        return $this->name ?? 0;
    }

    /**
     * 获取打印机类型列表
     */
    public static function getPrinterTypeList()
    {
        static $printerTypeEnum = [];
        if (empty($printerTypeEnum)) {
            $printerTypeEnum = PrinterTypeEnum::getTypeName();
        }
        return $printerTypeEnum;
    }

    /**
     * 打印机类型名称
     */
    public function getPrinterTypeAttr($value)
    {
        $printerType = $this->getPrinterTypeList();
        return ['value' => $value, 'text' => $printerType[$value] ?? ''];
    }

    /**
     * 自动转换printer_config为array格式
     */
    public function getPrinterConfigAttr($value)
    {
        return json_decode($value, true);
    }

    /**
     * 自动转换printer_config为json格式
     */
    public function setPrinterConfigAttr($value)
    {
        return json_encode($value) ?: '';
    }

    /**
     * 获取全部
     */
    public static function getAll($shop_supplier_id = 0)
    {
        $printerList = (new static)->order(['sort' => 'asc'])->select()->toArray();
        //
        $text = __('自带');
        //
        $cashierDevices = BindRecord::alias('a')
            ->where('source', BindRecord::SOURCE_CASHIER)
            ->where('platform', 'Android')
            ->whereIn('brand', BindRecord::BRANDS_PRINTS)
            ->field("a.device_id as printer_id")
            ->field("CONCAT(if(remark='', a.device_id, remark), ' ($text)') printer_name")
            ->order('id')
            ->select()
            ->toArray();
        //
        return array_merge($cashierDevices, $printerList);
    }

    /**
     * 获取小票打印机
     */
    public static function getNoteAll($shop_supplier_id = 0)
    {
        $where = [];
        if ($shop_supplier_id) {
            $where['shop_supplier_id'] = $shop_supplier_id;
        }
        return (new static)->where('is_delete', '=', 0)
            ->where('printer_type', '<>', 'FEI_E_YUN_TAG')
            ->where($where)
            ->order(['sort' => 'asc'])->select();
    }

    /**
     * 获取标签打印机
     */
    public static function getTagAll($shop_supplier_id = 0)
    {
        return (new static)->where('printer_type', '=', 'FEI_E_YUN_TAG')->order(['sort' => 'asc'])->select();
    }

    /**
     * 获取列表
     */
    public function getList($limit = 10, $shop_supplier_id = 0)
    {
        $printer = Setting::getSupplierItem(SettingEnum::PRINTER, $shop_supplier_id);
        $printerIds = array_column($printer['cashier_printer'] ?? [], 'printer_id');
        // todo 兼容
        // $supplierPrinterIds = Printing::column('printer_id');
        // foreach ($supplierPrinterIds as $supplierPrinterId) {
        //     foreach (json_decode($supplierPrinterId) ?: [] as $id) {
        //         $printerIds[] = $id;
        //     }
        // }
        $printerIds = implode(',', $printerIds);
        //
        return $this->alias('a')
            ->field("a.*, IF(find_in_set(a.uuid, '$printerIds'), 1, 0) as is_use")
            ->order(['a.sort' => 'asc'])
            ->paginate($limit, false, [
                'query' => Request::instance()->request()
            ]);
    }

    /**
     * 详情
     */
    public static function detail($printer_id)
    {
        return self::where('uuid', $printer_id)->find();
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null)
    {
        $filter = [
            'name' => $name,
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['uuid', '<>', $id];
        }
        return static::where($filter)->value('uuid') ? true : false;
    }
}
