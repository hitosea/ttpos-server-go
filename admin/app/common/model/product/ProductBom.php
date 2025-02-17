<?php

namespace app\common\model\product;

use app\common\library\helper;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\product\Category;
use app\common\model\erp\ErpPurchaseDetail;
use app\common\model\erp\ErpDamagedProductRecord;

/**
 * 商品BOM模型
 */
class ProductBom extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_bom';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     */
    protected $append = ['product_sku_id', 'spec_name_text'];

    /**
     * 兼容字段
     */
    public function getProductSkuIdAttr($value, $data = [])
    {
        return $this->uuid ?: 0;
    }
    public static function getSpecNameTextAttr($value, $data)
    {
        return extractLanguage($data['name']);
    }

    /**
     * 成品库存
     */
    public function getStockNumAttr($value)
    {
        return floatval($value);
    }

    /**
     * 材料库存
     */
    public function getMaterialStockAttr($value)
    {
        return floatval($value);
    }

    /**
     * 销量
     */
    public function getProductSalesAttr($value)
    {
        return floatval($value);
    }

    /**
     * 规格图片
     */
    public function image()
    {
        return $this->hasOne('app\\common\\model\\file\\UploadFile', 'file_id', 'image_id');
    }

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->belongsTo('app\\common\\model\\product\\Product', 'product_package_uuid', 'uuid')->with(['image', 'category', 'erpSupplier', 'erpSupplier.purchaser']);
    }

    /**
     * 产品规格关联材料（一对多）
     */
    public function material()
    {
        return $this->hasMany('app\\common\\model\\product\\Material', 'product_sku_id')->with(['materialProduct']);
    }

    /**
     * 通过规格获取商品SKU列表
     */
    public static function getProductBomList($params, $filterHavingMaterial = 0)
    {
       return (new self())->alias('bom')
            ->with(['product'])
            ->field(['bom.*'])
            ->where('bom.product_flavor_uuid', '>', 0)
            ->paginate($params);

        // todo 兼容
        // $model = (new ProductSku())->alias('ps')
        //     ->field('ps.*, p.category_id, p.erp_supplier_id, IF(p.type=10, ps.stock_num, ps.material_stock) as stock_num')
        //     ->leftJoin('product p', 'ps.product_id = p.product_id')
        //     ->where('p.is_delete', '=', 0)
        //     ->with(['product' => function ($query) {
        //         $query->with(['image', 'image.file', 'category', 'erpSupplier', 'erpSupplier.purchaser']);
        //     }]);

        // // 过滤产品规格有材料关联的数据
        // if ($filterHavingMaterial) {
        //     $model = $model->leftJoin('product_sku_material psm', 'psm.product_sku_id = ps.product_sku_id')
        //         ->where('psm.product_sku_id', null);
        // }

        // // 规格库存
        // if (isset($params['stock_num']) && $params['stock_num'] > 0) {
        //     $model = $model->where(function ($query) use ($params) {
        //         $query->where(function ($query) use ($params) {
        //             $query->where('p.type', '=', 10)
        //                 ->where('ps.stock_num', '<', $params['stock_num']);
        //         })->whereOr(function ($query) use ($params) {
        //             $query->where('p.type', '<>', 10)
        //                 ->where('ps.material_stock', '<', $params['stock_num']);
        //         });
        //     });
        // }

        // // 类型
        // if (isset($params['material_type']) && $params['material_type'] > 0) {
        //     $model->where('p.type', '=', $params['material_type']);
        // }

        // // 状态
        // if (isset($params['product_status']) && $params['product_status'] > 0) {
        //     $model->where('p.product_status', '=', $params['product_status']);
        // }

        // // 分类
        // if (!empty($params['category_id'])) {
        //     $categoryIds = (new Category)
        //         ->where('category_id', $params['category_id'])
        //         ->whereOr('parent_id', $params['category_id'])
        //         ->column('category_id');
        //     $model->where('p.category_id', 'IN', $categoryIds);
        // }

        // // 商品名称/条形码
        // if (isset($params['keyword']) && $params['keyword'] != '') {
        //     $keyword = trim($params['keyword']);
        //     $model->where(function ($query) use ($keyword) {
        //         $query->jsonLike('p.product_name', $keyword);
        //         $query->orLike('ps.barcode', $keyword);
        //     });
        // }

        // // 规格库存排序
        // if (isset($params['sort']) && $params['sort'] != '') {
        //     $model = $model->order('stock_num', $params['sort']);
        // }
        // $model = $model->order('ps.update_time', 'desc')->order('ps.product_sku_id');
        // $list = $model->paginate($params);

        // //
        // foreach ($list as &$item) {
        //     // 历史进货数
        //     $item['history_purchase_num'] = (new ErpPurchaseDetail())->sumActualPurchaseNum($item['product_sku_id']);
        //     // 历史报损数
        //     $item['history_loss_num'] = (new ErpDamagedProductRecord())->sumDamagedProductNum($item['product_sku_id']);
        // }
        // return $list;
    }
}
