<?php

namespace app\common\model\settings;

use think\facade\Cache;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\shop\BindRecord;
use app\common\model\supplier\Printing;
use app\common\enum\settings\SettingEnum;
use app\common\model\supplier\PrintingItem;

/**
 * 打印机模型
 */
class Printer extends BaseModel
{
    use SoftDelete;

    protected $name = 'printer';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $autoWriteTimestamp = true;
    protected $defaultSoftDelete = 0;

    /**
     * 追加属性
     */
    protected $append = ['printer_id', 'printer_name', 'printer_config', 'print_times'];

    /**
     * 分类更新后推送通知
     */
    public static function onAfterWrite()
    {
        Printing::clearProductPrinterListCache(self::$app_id);
    }

    /**
     * 分类删除后推送通知
     */
    public static function onAfterDelete()
    {
        Printing::clearProductPrinterListCache(self::$app_id);
    }
    
    /**
     * 兼容字段
     */
    public function getPrinterIdAttr($value, $data)
    {
        return $data['uuid'] ?: 0;
    }

    /**
     * 兼容字段
     */
    public function getPrinterNameAttr($value, $data)
    {
        return $data['name'] ?: '';
    }

    /**
     * 获取打印机类型列表
     */
    public static function getPrinterTypeList()
    {
        return PrinterType::getPrinterType();
    }


    /**
     * 自动转换printer_config为array格式
     */
    public function getPrinterConfigAttr($value, $data)
    {
        return json_decode($data['config_json'], true);
    }

    /**
     * 自动转换printer_config为json格式
     */
    public function setPrinterConfigAttr($value)
    {
        return json_encode($value) ?: '';
    }

    public function getPrintTimesAttr($value, $data)
    {
        return $data['copies'];
    }

    /**
     * 关联打印机类型
     */
    public function printerType()
    {
        return $this->belongsTo(PrinterType::class, 'printer_type_uuid', 'uuid');
    }

    /**
     * 获取全部
     */
    public static function getAll($isComesWithBind = true)
    {
        $printerList = (new static)->order(['sort' => 'asc'])
            ->select()
            ->toArray();
       
        // 添加USB标识
        foreach ($printerList as &$printer) {
            $printer['printer_name'] = $printer['printer_name'] . ($printer['is_usb'] == 1 ? ' (USB)' : '');
        }

        //
        $cashierDevices = [];
        if ($isComesWithBind) {
            $text = __('自带');
            $cashierDevices = BindRecord::alias('a')
                ->where("delete_time", 0)
                ->where('source', BindRecord::SOURCE_CASHIER)
                ->whereIn('brand', BindRecord::BRANDS_PRINTS)
                ->field("a.device_id as printer_id")
                ->field("CONCAT(if(remark='', a.device_id, remark), ' ($text)') printer_name")
                ->order('id')
                ->select()
                ->toArray();
        }
        
        //
        return array_merge($cashierDevices, $printerList);
    }

    /**
     * 获取小票打印机
     */
    public static function getNoteAll($shop_supplier_id = 0)
    {
        $list = (new static)->with([
            'printerType',
        ])->hasWhere('printerType', function($q) {
            $q->where('key', '<>', 'FEI_E_YUN_TAG');
        })->order(['sort' => 'asc'])->select();

        foreach ($list as $item) {
            $printerType = [ 'value' => '', 'text' => '' ];
            if ($item->printerType) {
                $printerType['value'] = $item->printerType['key'];
                $printerType['text'] = $item->printerType['name_text'];
            }
            $item->printer_type = $printerType;
        }

        return $list;
    }

    /**
     * 获取标签打印机
     */
    public static function getTagAll($shop_supplier_id = 0)
    {
        $list = (new static)->with([
            'printerType',
        ])->hasWhere('printerType', function($q) {
            $q->where('key', '<>', 'FEI_E_YUN_TAG');
        })->order(['sort' => 'asc'])->select();

        foreach ($list as $item) {
            $printerType = [ 'value' => '', 'text' => '' ];
            if ($item->printerType) {
                $printerType['value'] = $item->printerType['key'];
                $printerType['text'] = $item->printerType['name_text'];
            }
            $item->printer_type = $printerType;
        }

        return $list;
    }

    /**
     * 获取列表
     */
    public function getList($params, $shop_supplier_id = 0)
    {
        $printer = Setting::getSupplierItem(SettingEnum::PRINTER, $shop_supplier_id);
        $printerIds = array_column($printer['cashier_printer'] ?? [], 'printer_id');
        $supplierPrinterIds = PrintingItem::column('printer_uuid');
        foreach ($supplierPrinterIds as $supplierPrinterId) {
            $printerIds[] = $supplierPrinterId;
        }
        $printerIds = implode(',', $printerIds);
        //
        $paginate =  self::alias('a')
            ->field("a.*, IF(find_in_set(a.uuid, '$printerIds'), 1, 0) as is_use")
            ->with(['printerType'])
            ->order(['a.id' => 'desc'])
            ->paginate($params);

        $list = [];
        foreach ($paginate->items() as $item) {
            $list[] = [
                'printer_id' => $item['printer_id'],
                'printer_name' => $item['printer_name'],
                'printer_type' => [
                    'value' => $item['printerType']['key'] ?? '',
                    'text' => $item['printerType']['name_text'] ?? '',
                ],
                'sort' => $item['sort'],
                'is_usb' => $item['is_usb'],
                'create_time' => $item['create_time'],
                'printer_config' => $item['printer_config'],
                'is_use' => $item['is_use']
            ];
        }

        return [
            'current_page' => $paginate->currentPage(),
            'last_page' => $paginate->lastPage(),
            'per_page' => $paginate->listRows(),
            'total' => $paginate->total(),
            'data' => $list,
        ];
    }

    /**
     * 详情
     */
    public static function detail($printer_id)
    {
        return self::with(['printerType'])->where('uuid', $printer_id)->find();
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $id = null)
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
