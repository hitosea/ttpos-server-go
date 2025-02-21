<?php

namespace app\common\model\product;

use think\facade\Db;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 * 规格/属性(组)模型
 */
class Spec extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_flavor';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     */
    protected $append = ['spec_id', 'spec_name', 'spec_name_text'];

    /**
     * 兼容字段
     */
    public function getSpecIdAttr($value, $data)
    {
        return $this->uuid . "" ?: "0";
    }
    public function getSpecNameAttr($value)
    {
        return $this->getData('name') ?: '';
    }

    /**
     * 多语言关联
     */
    public function multiLanguageName()
    {
        return $this->hasOne('app\common\model\store\MultiLanguageName', 'uuid', 'multi_language_name_uuid');
    }
    /**
     * 规格名称
     * @param mixed $value
     * @param mixed $data
     * @return array|string
     */
    public function getSpecNameTextAttr($value, $data)
    {
        return extractLanguage($data['name']);
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
            ->field('pb.product_package_uuid as product_id')
            ->leftJoin('product_bom pb', 'pb.product_flavor_uuid = spec.uuid')
            ->where('pb.delete_time', 0)
            ->where('pb.product_flavor_uuid', $spec_id)
            ->select();
    }

    /**
     * 关联材料
     */
    public function material()
    {
        return $this->hasMany('app\\common\\model\\product\\ProductSkuMaterial', 'spec_id')->where('product_sku_id', '=', 0)->with(['materialProduct']);
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
        return $this->order(['create_time' => 'desc'])->select();

        // todo 兼容
        // $prefix = env('DB_PREFIX');
        // return $this->alias('sku')
        //     ->with(['material'])
        //     ->field('sku.*')
        //     ->field("IF(psku.sku_count IS NULL, 0, 1) AS is_used")
        //     ->field("IFNULL(psku.product_ids, '') AS product_ids")
        //     ->leftJoin("
        //     (
        //         SELECT psku.spec_sku_id, GROUP_CONCAT(DISTINCT product.product_id) AS product_ids, COUNT(DISTINCT psku.spec_sku_id) AS sku_count
        //         FROM {$prefix}product_sku psku
        //         LEFT JOIN {$prefix}product_package product ON psku.product_id = product.product_id
        //         WHERE product.is_delete = 0
        //         GROUP BY psku.spec_sku_id
        //     ) psku
        // ", 'sku.spec_id = psku.spec_sku_id')
        //     ->order(['create_time' => 'desc'])
        //     ->select();
    }

    /**
     * 获取列表数据
     */
    public function getLists($shop_supplier_id)
    {
        return $this->order(['create_time' => 'desc'])->select();
    }

    /**
     * 详情
     */
    public static function detail($spec_id)
    {
        return self::where('uuid', $spec_id)->find();
    }

    /**
     * 检查是否被关联
     */
    public function isUseWithProduct($spec_id)
    {
        // todo 兼容 兼容旧数据，先删除产品已删除的关联数据
        // ProductSku::where('product_id', 'in', function ($query) {
        //     $query->name('product')->field('product_id');
        // })->delete();
        // return ProductSku::where('spec_sku_id', 'in', $spec_id)->count() > 0;
        return false;
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null, $lang = 'zh')
    {
        $filter = [
            [Db::raw("JSON_UNQUOTE(JSON_EXTRACT(name, '$.$lang'))"), '=', $name],
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['uuid', '<>', $id];
        }
        return static::where($filter)->value('uuid') ? true : false;
    }
}
