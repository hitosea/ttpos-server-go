<?php

namespace app\common\model\erp;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\product\Product;
use app\common\model\product\Material;
use app\common\model\product\ProductBom;

/**
 * 采购单明细模型
 */
class ErpPurchaseDetail extends BaseModel
{
    use SoftDelete;
    protected $name = 'purchase_form_item';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;

    /**
     * material_type 物料类型 0-商品 1-原料
     */
    const MATERIAL_TYPE_PRODUCT = 0;
    const MATERIAL_TYPE_MATERIAL = 1;

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['estimate_purchase_num', 'estimate_purchase_price', 'estimate_total_amount', 'actual_purchase_num', 'actual_purchase_price', 'actual_total_amount'];

    /**
     * 兼容字段
     */
    public function getEstimatePurchaseNumAttr($value, $data)
    {
        return floatval($this->estimate_num ?: 0);
    }
    public function getEstimatePurchasePriceAttr($value, $data)
    {
        return floatval($this->estimate_price ?: 0);
    }
    public function getEstimateTotalAmountAttr($value, $data)
    {
        return floatval($this->estimate_amount ?: 0);
    }
    public function getActualPurchaseNumAttr($value, $data)
    {
        return floatval($this->num ?: 0);
    }
    public function getActualPurchasePriceAttr($value, $data)
    {
        return floatval($this->price ?: 0);
    }
    public function getActualTotalAmountAttr($value, $data)
    {
        return floatval($this->amount ?: 0);
    }

    /**
     * 获取商品名称
     */
    public function getProductNameAttr($value)
    {
        return extractLanguage($value ?: '');
    }

    /**
     * 获取商品规格名称
     */
    public function getProductSkuNameAttr($value)
    {
        return extractLanguage($value ?: '');
    }

    /**
     * 产品规格信息
     */
    public function sku()
    {
        return $this->belongsTo(ProductBom::class, 'material_uuid', 'uuid');
    }

    /**
     * 产品信息
     */
    public function product()
    {
        return $this->belongsTo(Product::class, 'material_uuid', 'uuid')->field('uuid, name, category_uuid, supplier_uuid')->with(['sku', 'image', 'erpSupplier', 'erpSupplier.purchaser']);
    }

    /**
     * 原料信息
     */
    public function material()
    {
        return $this->belongsTo(Material::class, 'material_uuid', 'uuid');
    }

    /**
     * 产品图片
     */
    public function productImage()
    {
        return $this->hasOne('app\\common\\model\\file\\UploadFile', 'file_id', 'product_image_id');
    }

    /**
     * 根据规格汇总采购数量
     */
    public function sumActualPurchaseNum($material_uuid)
    {
        return (new ErpPurchaseDetail())->where('material_uuid', $material_uuid)->sum('actual_purchase_num');;
    }

    /**
     * 全部列表
     */
    public function getListAll($purchase_order_uuid)
    {
        return $this->with(['product', 'sku'])->where('purchase_form_uuid', '=', $purchase_order_uuid)->select();
    }

    /**
     * 详情
     */
    public static function detail($id)
    {
        return self::where('uuid', $id)->find();
    }
}
