<?php

namespace app\common\model\product;

use app\common\library\helper;
use app\common\model\BaseModel;
use app\common\model\product\Category;
use app\common\model\erp\ErpPurchaseDetail;
use app\common\model\erp\ErpDamagedProductRecord;

/**
 * 商品SKU模型
 */
class ProductSku extends BaseModel
{
    protected $name = 'product_sku';
    protected $pk = 'product_sku_id';

    /**
     * 处理多语言
     */
    protected $append = ['spec_name_text'];
    public static function getSpecNameTextAttr($value, $data)
    {
        return extractLanguage($data['spec_name']);
    }

    /**
     * 规格图片
     */
    public function image()
    {
        return $this->hasOne('app\\common\\model\\file\\UploadFile', 'file_id', 'image_id');
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
        return $this->hasMany('app\\common\\model\\product\\ProductSkuMaterial', 'product_sku_id')->with(['materialProduct']);
    }

    /**
     * 通过规格获取商品SKU列表
     */
    public static function getSkuProductList($params, $filterHavingMaterial = 0)
    {
        $model = (new ProductSku())->alias('ps')
            ->field('ps.*, p.category_id, p.erp_supplier_id, IF(p.type=10, ps.stock_num, ps.material_stock) as stock_num')
            ->leftJoin('product p', 'ps.product_id = p.product_id')
            ->where('p.is_delete', '=', 0)
            ->with(['product' => function ($query) {
                $query->with(['image', 'image.file', 'category', 'erpSupplier', 'erpSupplier.purchaser']);
            }]);

        // 过滤产品规格有材料关联的数据
        if ($filterHavingMaterial) {
            $model = $model->leftJoin('product_sku_material psm', 'psm.product_sku_id = ps.product_sku_id')
                ->where('psm.product_sku_id', null);
        }

        // 规格库存
        if (isset($params['stock_num']) && $params['stock_num'] > 0) {
            $model = $model->where(function ($query) use ($params) {
                $query->where(function ($query) use ($params) {
                    $query->where('p.type', '=', 10)
                        ->where('ps.stock_num', '<', $params['stock_num']);
                })->whereOr(function ($query) use ($params) {
                    $query->where('p.type', '<>', 10)
                        ->where('ps.material_stock', '<', $params['stock_num']);
                });
            });
        }

        // 类型
        if (isset($params['material_type']) && $params['material_type'] > 0) {
            $model->where('p.type', '=', $params['material_type']);
        }

        // 状态
        if (isset($params['product_status']) && $params['product_status'] > 0) {
            $model->where('p.product_status', '=', $params['product_status']);
        }

        // 分类
        if (!empty($params['category_id'])) {
            $categoryIds = (new Category)
                ->where('category_id', $params['category_id'])
                ->whereOr('parent_id', $params['category_id'])
                ->column('category_id');
            $model->where('p.category_id', 'IN', $categoryIds);
        }

        // 商品名称/条形码
        if (isset($params['keyword']) && $params['keyword'] != '') {
            $keyword = trim($params['keyword']);
            $model->where(function ($query) use ($keyword) {
                $query->jsonLike('p.product_name', $keyword);
                $query->orLike('ps.barcode', $keyword);
            });
        }

        // 规格库存排序
        if (isset($params['sort']) && $params['sort'] != '') {
            $model = $model->order('stock_num', $params['sort']);
        }
        $model = $model->order('ps.update_time', 'desc')->order('ps.product_sku_id');
        $list = $model->paginate($params);

        //
        foreach ($list as &$item) {
            // 历史进货数
            $item['history_purchase_num'] = (new ErpPurchaseDetail())->sumActualPurchaseNum($item['product_sku_id']);
            // 历史报损数
            $item['history_loss_num'] = (new ErpDamagedProductRecord())->sumDamagedProductNum($item['product_sku_id']);
        }
        return $list;
    }

    /**
     * 获取sku信息详情
     */
    public static function skuDetail($productSkuId)
    {
        return static::with(['product'])->where('product_sku_id', '=', $productSkuId)->find();
    }

    /**
     * 获取sku信息详情
     */
    public static function detail($productId, $specSkuId)
    {
        return static::where('product_id', '=', $productId)
            ->where('spec_sku_id', '=', $specSkuId)->find();
    }

    /**
     * 获取sku信息详情
     */
    public static function getDetailSku($productId, $productSkuId)
    {
        return static::where('product_id', '=', $productId)
            ->where('product_sku_id', '=', $productSkuId)->find();
    }

    // 所有商品库存(不含关联材料的商品)
    public static function getTotalStock()
    {
        $product_stock = (new ProductSku)->alias('ps')
            ->leftJoin('product_sku_material psm', 'psm.product_sku_id = ps.product_sku_id')
            ->leftJoin('product p', 'ps.product_id = p.product_id')
            ->where('psm.product_sku_id', null)
            ->where('p.product_type', 1)
            ->where('p.type', Product::TYPE_PRODUCT)
            ->where('p.is_delete', 0)
            ->sum('ps.stock_num');
        $product_material_stock = (new ProductSku)->alias('ps')
            ->join('product p', 'ps.product_id = p.product_id')
            ->where('p.product_type', 1)
            ->where('p.type', Product::TYPE_MATERIAL)
            ->where('p.is_delete', 0)
            ->sum('ps.material_stock');
        $total_stock = helper::bcadd($product_stock, $product_material_stock, 4);
        return floatval($total_stock);
    }

    // 通过sku_id获取商品名+规格
    public static function getNameById($product_sku_id)
    {
        $sku = self::where('product_sku_id', $product_sku_id)->find();
        if ($sku) {
            $product = Product::where('product_id', $sku->product_id)->find();
            if ($product->product_name_text && $sku->spec_name_text) {
                return $product->product_name_text . ' （' . $sku->spec_name_text . '）';
            }
            return $product->product_name_text;
        }
        return '';
    }

    /**
     * 检查产品条码唯一性
     */
    public function checkProductBarcodeExist($name, $id = null)
    {
        $filter = [
            'sku.barcode' => $name,
            'p.is_delete' => 0,
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['sku.product_sku_id', '<>', $id];
        }
        return static::alias('sku')
            ->leftJoin('product p', 'sku.product_id = p.product_id')
            ->where($filter)
            ->where('sku.barcode', '<>', '') // 兼容导入条码可为空
            ->value('sku.product_sku_id') ? true : false;
    }
}
