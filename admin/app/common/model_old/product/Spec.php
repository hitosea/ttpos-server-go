<?php

namespace app\common\model_old\product;

use think\facade\Db;
use app\common\model_old\BaseModel;

/**
 * 规格/属性(组)模型
 */
class Spec extends BaseModel
{
    protected $name = 'spec';
    protected $pk = 'spec_id';
    protected $updateTime = false;

    /**
     * 处理多语言
     */
    protected $append = ['spec_name_text'];

    /**
     * 规格名称
     * @param mixed $value
     * @param mixed $data
     * @return array|string
     */
    public function getSpecNameTextAttr($value, $data)
    {
        return extractLanguage($data['spec_name']);
    }

    /**
     * 关联产品ids
     */
    public function getProductIdsAttr($value, $data = [])
    {
        $product_ids = $data['product_ids'] ?? $value ?? '';
        if (empty($product_ids)) {
            return [];
        }
        $arr = array_map('intval', explode(',', $product_ids));
        return array_values($arr);
    }

    /**
     * 关联产品
     */
    public function productSku($spec_id)
    {
        return $this->alias('spec')
            ->field('product.product_id')
            ->leftJoin('product_sku sku', 'sku.spec_sku_id = spec.spec_id')
            ->leftJoin('product product', 'product.product_id = sku.product_id')
            ->where('product.is_delete', 0)
            ->where('spec.spec_id', $spec_id)
            ->select();
    }

    /**
     * 关联材料
     */
    public function material()
    {
        return $this->hasMany('app\\common\\model_old\\product\\ProductSkuMaterial', 'spec_id')->where('product_sku_id', '=', 0)->with(['materialProduct']);
    }

    /**
     * 更新规格组
     * @param mixed $data
     * @return void
     */
    public function updateSpec($data)
    {
        // todo v1.0.8需求变更，暂时不需要
        // if ($data) {
        //     $addData = [];
        //     foreach ($data as $item) {
        //         if ($item['spec_name']) {
        //             $isExit = $this->where('spec_name', '=', $item['spec_name'])->count();
        //             if ($isExit == 0) {
        //                 $addData[] = [
        //                     'spec_name' => $item['spec_name'],
        //                     'app_id' => self::$app_id,
        //                     'shop_supplier_id' => self::$app_id,
        //                 ];
        //             }
        //         }
        //     }
        //     $addData && $this->saveAll($addData);
        // }
    }

    /**
     * 获取列表数据
     */
    public function getAllList($shop_supplier_id)
    {
        $prefix = 'jjjfood_';
        return $this->alias('sku')
        ->with(['material'])
        ->field('sku.*')
        ->field("IF(psku.sku_count IS NULL, 0, 1) AS is_used")
        ->field("IFNULL(psku.product_ids, '') AS product_ids")
        ->leftJoin("
            (
                SELECT psku.spec_sku_id, GROUP_CONCAT(DISTINCT product.product_id) AS product_ids, COUNT(DISTINCT psku.spec_sku_id) AS sku_count
                FROM {$prefix}product_sku psku
                LEFT JOIN {$prefix}product product ON psku.product_id = product.product_id
                WHERE product.is_delete = 0
                GROUP BY psku.spec_sku_id
            ) psku
        ", 'sku.spec_id = psku.spec_sku_id')
        ->where('shop_supplier_id', '=', $shop_supplier_id)
        ->order(['create_time' => 'desc'])
        ->select();
    }

    /**
     * 获取列表数据
     */
    public function getLists($shop_supplier_id)
    {
        return $this->where('shop_supplier_id', '=', $shop_supplier_id)->order(['create_time' => 'desc'])->select();
    }

    /**
     * 详情
     */
    public static function detail($spec_id)
    {
        return self::find($spec_id);
    }

    /**
     * 检查是否被关联
     */
    public function isUseWithProduct($spec_id)
    {
        // 兼容旧数据，先删除产品已删除的关联数据
        ProductSku::where('product_id', 'in', function ($query) {
            $query->name('product')->where('is_delete', '=', 1)->field('product_id');
        })->delete();
        return ProductSku::where('spec_sku_id', 'in', $spec_id)->count() > 0;
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null, $lang = 'zh')
    {
        $filter = [
            [Db::raw("JSON_UNQUOTE(JSON_EXTRACT(spec_name, '$.$lang'))"), '=', $name],
            'shop_supplier_id' => $shop_supplier_id
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['spec_id', '<>', $id];
        }
        return static::where($filter)->value('spec_id') ? true : false;
    }
}
