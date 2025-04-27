<?php

namespace app\common\model_old\erp;

use app\common\model_old\BaseModel;
use app\common\model_old\product\Product;
use app\common\model_old\product\ProductSku;

/**
 * 采购单明细模型
 */
class ErpPurchaseDetail extends BaseModel
{
    protected $name = 'erp_purchase_detail';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [];

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
     * 去零
     */
    public function getActualPurchasePriceAttr($value)
    {
        return floatval($value ?: 0);
    }

    /**
     * 去零
     */
    public function getActualPurchaseNumAttr($value)
    {
        return floatval($value ?: 0);
    }

    /**
     * 去零
     */
    public function getActualTotalAmountAttr($value)
    {
        return floatval($value ?: 0);
    }

    /**
     * 去零
     */
    public function getEstimatePurchaseNumAttr($value)
    {
        return floatval($value ?: 0);
    }

    /**
     * 去零
     */
    public function getEstimatePurchasePriceAttr($value)
    {
        return floatval($value ?: 0);
    }

    /**
     * 去零
     */
    public function getEstimateTotalAmountAttr($value)
    {
        return floatval($value ?: 0);
    }

    /**
     * 去零
     */
    public function getProductStockAttr($value)
    {
        return floatval($value ?: 0);
    }

    /**
     * 产品规格信息
     */
    public function sku()
    {
        return $this->belongsTo(ProductSku::class, 'product_sku_id', 'product_sku_id')->with(['product']);
    }

    /**
     * 产品信息
     */
    public function product()
    {
        return $this->belongsTo(Product::class, 'product_id', 'product_id')->field('product_id, product_name, type, category_id, erp_supplier_id')->with(['sku', 'image', 'image.file', 'erpSupplier', 'erpSupplier.purchaser']);
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
    public function sumActualPurchaseNum($product_sku_id)
    {
        return (new ErpPurchaseDetail())->where('product_sku_id', $product_sku_id)->sum('actual_purchase_num');;
    }

    /**
     * 全部列表
     */
    public function getListAll($purchase_order_id)
    {
        return $this->with(['product', 'sku'])->where('purchase_order_id', '=', $purchase_order_id)->select();
    }

    /**
     * 详情
     */
    public static function detail($id)
    {
        return self::find($id);
    }
}
