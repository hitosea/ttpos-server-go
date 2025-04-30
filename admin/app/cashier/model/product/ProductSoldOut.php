<?php

namespace app\cashier\model\product;

use app\common\model\product\ProductSoldOut as ProductSoldOutModel;

/**
 * 商品售罄模型
 */
class ProductSoldOut extends ProductSoldOutModel
{

    /**
     * 获取商品简易列表
     */
    public function productSkuList($params)
    {
        $search = $params['search'] ?? "";
        //
        return $this->alias('a')
            ->leftJoin('product product', 'product.product_id = a.product_id')
            ->leftJoin('product_sku sku', 'sku.product_sku_id = a.product_sku_id')
            ->field([
                'product.product_id',
                'product.product_name',
                'sku.product_sku_id',
                'sku.spec_name',
            ])
            ->when($search, function ($q) use ($search) {
                $q->jsonLike('product.product_name', $search);
            })
            ->where('product.type', '=', 10)
            ->where('product.is_delete', '=', 0)
            ->where('product.product_type', '=', 1)
            ->where('product.shop_supplier_id', '=', $params['shop_supplier_id'])
            ->where('product.product_status', '=', 10)
            ->whereNotNull('sku.product_sku_id')
            ->order(['product.product_sort', 'product.product_id' => 'desc'])
            ->paginate($params)
            ->append(['product_name_text', 'spec_name_text'])
            ->hidden(['product_name', 'product_unit', 'spec_name'])
            ->toArray();
    }

    /**
     * 售罄数量
     * @param $shop_supplier_id
     * @return mixed
     */
    public function soldOutCount($shop_supplier_id)
    {
        //
        return $this->alias('a')
            ->leftJoin('product product', 'product.product_id = a.product_id')
            ->leftJoin('product_sku sku', 'sku.product_sku_id = a.product_sku_id')
            ->where('product.type', '=', 10)
            ->where('product.is_delete', '=', 0)
            ->where('product.product_type', '=', 1)
            ->where('product.shop_supplier_id', '=', $shop_supplier_id)
            ->where('product.product_status', '=', 10)
            ->whereNotNull('sku.product_sku_id')
            ->count();
    }
}
