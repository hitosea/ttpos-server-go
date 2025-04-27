<?php

namespace app\common\model_old\product;

use think\facade\Db;
use app\common\model_old\BaseModel;

/**
 * 单位模型
 */
class Unit extends BaseModel
{
    protected $name = 'product_unit';
    protected $pk   = 'unit_id';

    /**
     * 处理多语言
     */
    protected $append = ['unit_name_text'];

    /**
     * 单位名称
     */
    public function getUnitNameTextAttr($value, $data)
    {
        return extractLanguage($data['unit_name']);
    }

    /**
     * 关联产品ids
     */
    public function getProductIdsAttr($value, $data = [])
    {
        return $this->product()->where('is_delete', 0)->column('product_id');
    }

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->hasMany('app\\common\\model\\product\\Product', 'unit_id', 'unit_id');
    }

    /**
     * 更新单位
     * @param mixed $unit_name
     * @param mixed $shop_supplier_id
     * @return void
     */
    public function updateUnit($unit_name, $shop_supplier_id)
    {
        // todo 不用再次校验单位新增，单位添加是已固定单位，先注释掉
        // if ($unit_name) {
        //     $isExit = $this->where('unit_name', '=', $unit_name)
        //         ->where('shop_supplier_id', '=', $shop_supplier_id)
        //         ->count();
        //     if ($isExit == 0) {
        //         $addData = [
        //             'unit_name'        => $unit_name,
        //             'shop_supplier_id' => $shop_supplier_id,
        //             'app_id'           => self::$app_id
        //         ];
        //         $this->save($addData);
        //     }
        // }
    }

    /**
     * 获取列表数据
     */
    public function getAllList($shop_supplier_id)
    {
        return $this->where('shop_supplier_id', '=', $shop_supplier_id)->order(['sort' => 'asc', 'create_time' => 'desc'])->select()?->append(['product_ids'], true);
    }

    /**
     * 获取列表数据
     */
    public function getLists($shop_supplier_id)
    {
        return $this->where('shop_supplier_id', '=', $shop_supplier_id)->order(['sort' => 'asc', 'create_time' => 'desc'])->select();
    }

    /**
     * 详情
     */
    public static function detail($unit_id)
    {
        return self::find($unit_id);
    }

    /**
     * 检查是否被关联
     */
    public function isUseWithProduct($unit_id)
    {
        return Product::where('unit_id', 'in', $unit_id)->where('is_delete', 0)->count() > 0;
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null, $lang = 'zh')
    {
        $filter = [
            [Db::raw("JSON_UNQUOTE(JSON_EXTRACT(unit_name, '$.$lang'))"), '=', $name],
            'shop_supplier_id' => $shop_supplier_id
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['unit_id', '<>', $id];
        }
        return static::where($filter)->value('unit_id') ? true : false;
    }
}
