<?php

namespace app\common\model\supplier;

use app\common\model\BaseModel;
use app\common\model\product\Product;
use app\common\model\settings\Printer;

/**
 * 菜品打印模型
 */
class Printing extends BaseModel
{
    protected $name = 'supplier_printing';
    protected $pk = 'id';

    // 打印类型
    const PRINT_TYPE_BACK_FOOD = 0;    // 退菜打印
    const PRINT_TYPE_PAY = 10;         // 付款打印
    const PRINT_TYPE_ADD_ORDER = 20;   // 下单打印
    const PRINT_TYPE_KITCHEN = 30;     // 送厨打印

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['printer_name_text'];

    /**
     * 获取打印机名称
     */
    public function getPrinterNameTextAttr($value, $data)
    {
        $printerIds = $data['printer_id'] ? json_decode($data['printer_id'], true) : [];
        $printer = Printer::where('printer_id', 'in', $printerIds)->where('is_delete', 0)->select()->toArray();
        $printer = array_column($printer, 'printer_name');
        if (count($printer) !== count($printerIds)) {
            $printer[] = '-';
        }
        return implode(',', $printer);
    }

    /**
     * 获取商品ids列表
     */
    public function getProductIdsAttr($value, $data)
    {
        $product_ids = $data['product_ids'] ? json_decode($data['product_ids'], true) : [];
        $list = Product::alias('product')
            ->field(['product.product_id', 'product.category_id', 'product.label_id', 'pc.category_id as parent_category_id'])
            ->leftJoin('category c', 'c.category_id = product.category_id')
            ->leftJoin('category pc', 'pc.category_id = c.parent_id')
            ->where('product_id', 'in', $product_ids)
            ->where('is_delete', 0)
            ->select()?->append([]);
        return $list->toArray();
    }

    /**
     * 获取商品分类
     */
    public function getCategoryIdAttr($value, $data)
    {
        return $value ? json_decode($value, true) : [];
    }

    /**
     * 设置商品分类
     */
    public function setProductIdsAttr($value, $data)
    {
        return json_encode($value) ?: '';
    }

    /**
     * 设置商品分类
     */
    public function setCategoryIdAttr($value, $data)
    {
        return json_encode($value) ?: '';
    }

    /**
     * 关联打印机表
     */
    public function printers()
    {
        return Printer::whereIn('printer_id', $this->printer_id)->select();
    }

    /**
     * 获取打印标签
     */
    public function getLabelIdAttr($value, $data)
    {
        return $value ? json_decode($value, true) : [];
    }

    /**
     * 设置打印标签
     */
    public function setLabelIdAttr($value, $data)
    {
        return json_encode($value) ?: '';
    }

    /**
     * 获取区域
     */
    public function getAreaIdAttr($value, $data)
    {
        return $value ? json_decode($value, true) : [];
    }

    /**
     * 设置区域
     */
    public function setAreaIdAttr($value, $data)
    {
        return json_encode($value) ?: '';
    }

    /**
     * 获取打印机ids
     */
    public function getPrinterIdAttr($value, $data)
    {
        return $value ? json_decode($value, true) : [];
    }

    /**
     * 设置打印机ids
     */
    public function setPrinterIdAttr($value, $data)
    {
        return json_encode($value) ?: '';
    }

    /**
     * 关联供应商表
     */
    public function supplier()
    {
        return $this->belongsTo('app\\common\\model\\supplier\\Supplier', 'shop_supplier_id', 'shop_supplier_id');
    }


    /**
     * 详情
     */
    public static function detail($id, $with = [])
    {
        return static::with($with)->find($id);
    }

    /**
     * 列表
     */
    public function getList($print_type, $shop_supplier_id, $product_type)
    {
        return $this->where('print_type', '=', $print_type)
            ->where('shop_supplier_id', '=', $shop_supplier_id)
            // ->where('product_type', '=', $product_type)
            ->where('is_open', '=', 1)
            ->where('is_delete', '=', 0)
            ->select();
    }

    /**
     * 获取打印档口列表
     */
    public function getPrintPortList($shop_supplier_id = 0)
    {
        return $this->where('is_open', '1')->where('is_delete', '0')->where('shop_supplier_id', $shop_supplier_id)->order(['create_time' => 'desc'])->select();
    }

    /**
     * 设置状态
     */
    public function setStatus($status)
    {
        return $this->save(['is_open' => $status ? 1 : 0]);
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null)
    {
        $filter = [
            'name' => $name,
            'is_delete' => 0,
            'shop_supplier_id' => $shop_supplier_id
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['id', '<>', $id];
        }
        return static::where($filter)->where('is_delete', 0)->value('id') ? true : false;
    }
}
