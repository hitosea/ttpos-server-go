<?php

namespace app\common\model\product;

use app\common\library\helper;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\product\Category;
use app\common\model\erp\ErpPurchaseDetail;
use app\common\model\erp\ErpDamagedProductRecord;

/**
 * 商品SKU模型
 */
class ProductBom extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_bom';
    protected $pk = 'id';

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
        return $this->belongsTo('app\\common\\model\\product\\Product', 'product_id', 'product_id')->with(['image', 'image.file', 'category', 'erpSupplier', 'erpSupplier.purchaser']);
    }

    /**
     * 产品规格关联材料（一对多）
     */
    public function material()
    {
        return $this->hasMany('app\\common\\model\\product\\Material', 'product_sku_id')->with(['materialProduct']);
    }
}
