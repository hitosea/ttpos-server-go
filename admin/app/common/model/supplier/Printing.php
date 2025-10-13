<?php

namespace app\common\model\supplier;

use think\facade\Cache;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 * 菜品打印模型
 */
class Printing extends BaseModel
{
    use SoftDelete;

    protected $name = 'product_printer';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $autoWriteTimestamp = true;
    protected $defaultSoftDelete = 0;

    // 打印类型
    const PRINT_TYPE_BACK_FOOD = 0;    // 退菜打印
    const PRINT_TYPE_PAY = 10;         // 付款打印
    const PRINT_TYPE_ADD_ORDER = 20;   // 下单打印
    const PRINT_TYPE_KITCHEN = 30;     // 送厨打印

    // 打印模式字段映射
    const PRINT_MODE_MAP = [
        self::PRINT_TYPE_PAY => 0,
        self::PRINT_TYPE_KITCHEN => 1,
    ];

    // 打印模式字段反向映射
    const PRINT_MODE_REVERSE_MAP = [
        self::PRINT_TYPE_PAY,
        self::PRINT_TYPE_KITCHEN,
    ];

    // 打印方式
    const PRINT_METHOD = -1; // 全选
    const PRINT_METHOD_ALL = 10; // 整单打印
    const PRINT_METHID_ONE = 40; // 按一菜一单打印

    // 打印方式字段映射
    const PRINT_METHOD_MAP = [
        self::PRINT_METHOD => -1,
        self::PRINT_METHOD_ALL => 0,
        self::PRINT_METHID_ONE => 1,
    ];

    // 打印方式字段反向映射
    const PRINT_METHOD_REVERSE_MAP = [
        -1 => self::PRINT_METHOD,
        0 => self::PRINT_METHOD_ALL,
        1 => self::PRINT_METHID_ONE,
    ];

    // 打印商品选择
    const PRINT_PRODUCT_SELECT_CATEGORY = 1; // 按商品分类
    const PRINT_PRODUCT_SELECT_TAG = 2; // 按打印标签

    // 打印商品选择字段映射
    const PRINT_PRODUCT_SELECT_MAP = [
        self::PRINT_PRODUCT_SELECT_CATEGORY => 0,
        self::PRINT_PRODUCT_SELECT_TAG => 1,
    ];

    // 打印商品选择字段反向映射
    const PRINT_PRODUCT_SELECT_REVERSE_MAP = [
        self::PRINT_PRODUCT_SELECT_CATEGORY,
        self::PRINT_PRODUCT_SELECT_TAG,
    ];

    // 打印场景
    const PRINT_MODE_SCENE_MERGE = 1; // 合并
    const PRINT_MODE_SCENE_SEPARATE = 2; // 分开

    // 打印场景字段映射
    const PRINT_MODE_SCENE_MAP = [
        self::PRINT_MODE_SCENE_MERGE => 0,
        self::PRINT_MODE_SCENE_SEPARATE => 1,
    ];

    // 打印场景字段反向映射
    const PRINT_MODE_SCENE_REVERSE_MAP = [
        self::PRINT_MODE_SCENE_MERGE,
        self::PRINT_MODE_SCENE_SEPARATE,
    ];

    /**
     * 分类更新后推送通知
     */
    public static function onAfterWrite()
    {
        self::clearProductPrinterListCache(self::$app_id);
    }

    /**
     * 分类删除后推送通知
     */
    public static function onAfterDelete()
    {
        self::clearProductPrinterListCache(self::$app_id);
    }

    /**
     * 关联商品打印详情
     */
    public function printingItem()
    {
        return $this->hasMany(PrintingItem::class, 'product_printer_uuid', 'uuid');
    }

    /**
     * 关联商品打印区域
     */
    public function printingRegion()
    {
        return $this->hasMany(PrintingRegion::class, 'product_printer_uuid', 'uuid');
    }

    /**
     * 关联商品打印-商品列表
     */
    public function printingProductItem()
    {
        return $this->hasMany(PrintingProduct::class, 'product_printer_uuid', 'uuid');
    }

    /**
     * 获取打印机
     */
    public static function getPrinterList($printingItemList)
    {
        if (empty($printingItemList)) {
            return '';
        }

        $printerList = [];
        foreach ($printingItemList as $item) {
            $printer = $item['printer'] ?? null;
            if ($printer) {
                $printerList[] = $printer;
            }
        }

        return $printerList;
    }

    /**
     * 获取打印机名称
     */
    public static function getPrinterNameText($printingItemList)
    {
        if (empty($printingItemList)) {
            return '';
        }

        $printerList = [];
        foreach ($printingItemList as $item) {
            $printer = $item['printer'] ?? null;
            if (!$printer) {
                $printerList[] = '-';
            } else {
                $printerList[] = $printer['printer_name'].'('.($printer['is_usb'] == 1 ? 'USB' : '').')';
            }
        }

        return implode(',', $printerList);
    }


    /**
     * 详情
     */
    public static function detail($id, $with = [])
    {
        return static::with($with)->where('id', $id)->find();
    }

    /**
     * 构建详情数据
     */
    public function buildDetailData()
    {
        // 区域uuid列表
        $areaUuidList = [];
        foreach ($this['printingRegion'] as $region) {
            $areaUuidList[] = "{$region['desk_region_uuid']}";
        }
        // 商品列表
        $productList = [];
        foreach ($this['printingProductItem'] as $productItem) {
            $productList[] = [
                'product_id' => $productItem['product_package_uuid'] ?? 0,
                'category_id' => $productItem['product']['category_uuid'] ?? 0,
                'label_id' => $productItem['product']['label_id'] ?? 0,
                'parent_category_id' => $productItem['product']['category']['parent_id'] ?? 0,
            ];
        }
        // 打印机uuid列表
        $printerUuidList = [];
        foreach ($this['printingItem'] as $printingItem) {
            $printerUuidList[] = "{$printingItem['printer_uuid']}";
        }

        return [
            'id' => $this['id'],
            'name' => $this['name'],
            'is_open' => $this['status'],
            'copies' => $this['copies'],
            'print_type' => self::PRINT_MODE_REVERSE_MAP[$this['print_mode']],
            'print_method' => self::PRINT_METHOD_REVERSE_MAP[$this['print_method']],
            'product_method' => self::PRINT_PRODUCT_SELECT_REVERSE_MAP[$this['print_product_select']],
            'print_select' => self::PRINT_MODE_SCENE_REVERSE_MAP[$this['print_mode_scene']],
            'area_id' => $areaUuidList,
            'product_ids' => $productList,
            'printer_id' => $printerUuidList,
        ];
    }

    /**
     * 列表
     */
    public function getList($print_type, $shop_supplier_id, $product_type)
    {
        return $this->where('print_type', '=', $print_type)->where('is_open', '=', 1)->select();
    }

    /**
     * 获取打印档口列表
     */
    public function getPrintPortList($shop_supplier_id = 0)
    {
        return $this->where('is_open', '1')->order(['create_time' => 'desc'])->select();
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
            $filter[] = ['id', '<>', $id];
        }
        return static::where($filter)->value('id') ? true : false;
    }

    /**
     * 删除商品打印机列表缓存
     */
    public static function clearProductPrinterListCache($appid)
    {
        Cache::set(sprintf("PRODUCT_PRINTER_LIST_v2:%d:%d", $appid, 0), null);
        Cache::set(sprintf("PRODUCT_PRINTER_LIST_v2:%d:%d", $appid, 1), null);
        Cache::set(sprintf("PRODUCT_PRINTER_LIST_v2:%d:%d", $appid, -1), null);
        Cache::set(sprintf("PRODUCT_PRINTER_LIST_v2:%d:%d", $appid, -2), null);
    }
}
