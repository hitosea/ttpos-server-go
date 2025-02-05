<?php

namespace app\common\service\product\factory;

use app\common\model\product\ProductSoldOut;
use app\common\enum\product\DeductStockTypeEnum;
use app\common\model\product\Product as ProductModel;
use app\common\model\product\ProductSku as ProductSkuModel;

/**
 * 商品来源-普通商品扩展类
 */
class MasterProductService extends ProductService
{
    /**
     * 更新商品库存 (针对下单减库存的商品)
     */
    public function updateProductStock($productList)
    {
        $productData = [];
        $productSkuData = [];
        foreach ($productList as $product) {
            // 下单减库存
            if ($product['product']['deduct_stock_type'] == DeductStockTypeEnum::CREATE) {
                // 总库存
                $productData[] = [
                    'data' => ['product_stock' => ['dec', $product['product_num']]],
                    'where' => [
                        'product_id' => $product['product_id'],
                    ],
                ];
                $productSkuData[] = [
                    'data' => ['stock_num' => ['dec', $product['product_num']]],
                    'where' => [
                        'product_id' => $product['product_id'],
                        'product_sku_id' => $product['product_sku_id'],
                    ],
                ];
            }
        }
        try {
            // 更新商品sku库存
            !empty($productSkuData) && $this->updateProductSku($productSkuData);
        } catch (\Exception $e) {
            log_write('master updateProductStock' . $e->getMessage());
        }
        return true;
    }

    /**
     * 更新商品库存销量（订单付款后）
     * 关联点：结账后更新库存
     */
    public function updateStockSales($sourceProductList, $type = 'dec', $order_id = 0)
    {
        $license = request()->licenses;
        //
        $productList = $sourceProductList['orderProductList'];
        $allProductSkuList = $sourceProductList['allProductSkuList'];
        //
        $mergeProductList = [];
        foreach ($productList as $product) {
            if ($product['is_return'] != 1) {
                $cloneProduct = (clone $product);
                $key = $cloneProduct->product_id . '_' . $cloneProduct->product_sku_id . '_' . $cloneProduct->deduct_stock_type;
                $cloneProduct->total_num = ($mergeProductList[$key]['total_num'] ?? 0) + $cloneProduct->total_num;
                $mergeProductList[$key] = $cloneProduct;
            }
        }
        //
        $productData = [];
        $productSkuData = [];
        foreach ($mergeProductList as $product) {
            $productSku = $allProductSkuList[$product['product_sku_id']] ?? [];
            // 记录商品的销量
            $product_data = [
                'data' => ['sales_actual' => [$type == 'dec' ? 'inc' : 'dec', $product['total_num']]],
                'where' => [
                    'product_id' => $product['product_id']
                ],
            ];
            // 付款减库存
            if ((($license['sale'] ?? 0) != 1 || empty($productSku['material'])) && $product['deduct_stock_type'] == DeductStockTypeEnum::PAYMENT) {
                //总库存
                $product_data['data']['product_stock'] = [$type, $product['total_num']];
                //sku库存
                $productSkuData[] = [
                    'data' => [
                        'stock_num' => [$type, $product['total_num']],
                        'product_sales' => [$type == 'dec' ? 'inc' : 'dec', $product['total_num']]
                    ],
                    'where' => [
                        'product_id' => $product['product_id'],
                        'product_sku_id' => $product['product_sku_id'],
                    ],
                ];
            }
            $productData[] = $product_data;
        }

        try {
            // 更新商品销量
            !empty($productData) && $this->updateProduct($productData);
            // 更新商品sku库存
            !empty($productSkuData) && $this->updateProductSku($productSkuData);
        } catch (\Exception $e) {
            log_write('master updateStockSales' . $e->getMessage());
        }
        return true;
    }

    /**
     * 回退商品库存
     */
    public function backProductStock($productList, $isPayOrder = false)
    {
        $productData = [];
        $productSkuData = [];
        foreach ($productList as $product) {
            $product_item = [
                'where' => [
                    'product_id' => $product['product_id'],
                ],
                'data' => ['product_stock' => ['inc', $product['total_num']]],
            ];
            $sku_item = [
                'where' => [
                    'product_id' => $product['product_id'],
                    'product_sku_id' => $product['product_sku_id'],
                ],
                'data' => ['stock_num' => ['inc', $product['total_num']]],
            ];
            if ($isPayOrder == true) {
                // 付款订单全部库存
                $productData[] = $product_item;
                $productSkuData[] = $sku_item;
            } else {
                // 未付款订单，判断必须为下单减库存时才回退
                $product['deduct_stock_type'] == DeductStockTypeEnum::CREATE && $productData[] = $product_item;
                $product['deduct_stock_type'] == DeductStockTypeEnum::CREATE && $productSkuData[] = $sku_item;
            }
        }
        // 更新商品库存
        !empty($productData) && $this->updateProduct($productData);
        // 更新商品sku库存
        !empty($productSkuData) && $this->updateProductSku($productSkuData);
        return true;
    }

    /**
     * 更新商品信息
     */
    private function updateProduct($data)
    {
        return (new ProductModel)->updateAll($data);
    }

    /**
     * 更新商品sku信息
     */
    private function updateProductSku($data)
    {
        return (new ProductSkuModel)->updateAll($data);
    }

    /**
     * 送厨更新商品库存 (针对下单减库存的商品)
     * 关联点：收银送厨、平板下单
     */
    public function updateOrderProductStock($sourceProductList, $type = 'dec')
    {
        // 授权信息
        $license = request()->licenses;
        //
        $productList = $sourceProductList['orderProductList'];
        $allProductSkuList = $sourceProductList['allProductSkuList'];
        //
        $mergeProductList = [];
        foreach ($productList as $product) {
            $cloneProduct = (clone $product);
            $key = $cloneProduct->product_id . '_' . $cloneProduct->product_sku_id . '_' . $cloneProduct->deduct_stock_type;
            $cloneProduct['total_num'] = ($mergeProductList[$key]['total_num'] ?? 0) + $cloneProduct->total_num;
            $mergeProductList[$key] = $cloneProduct;
        }
        //
        $productData = [];
        $productSkuData = [];
        $error = [];
        //
        foreach ($mergeProductList as $product) {
            // 是否放入已售罄
            if (ProductSoldOut::where('product_id', $product['product_id'])->where('product_sku_id', $product['product_sku_id'])->find()) {
                $error[] = [
                    'order_product_id' => $product['order_product_id'],
                    'product_id' => $product['product_id'],
                    'product_sku_id' => $product['product_sku_id'],
                    'total_num' => $product['total_num'],
                    'product_name_text' => ProductSkuModel::getNameById($product['product_sku_id']),
                    'tablet_product_name_text' => ProductSkuModel::getNameById($product['product_sku_id']),
                ];
                continue;
            }
            // 下单减库存
            if ($product['deduct_stock_type'] == DeductStockTypeEnum::CREATE) {
                $productSku = $allProductSkuList[$product['product_sku_id']] ?? [];
                if ((($license['sale'] ?? 0) != 1 || empty($productSku['material'])) && $type == 'dec' && $product['total_num'] > $productSku['stock_num']) {
                    $error[] = [
                        'order_product_id' => $product['order_product_id'],
                        'product_id' => $product['product_id'],
                        'product_sku_id' => $product['product_sku_id'],
                        'total_num' => $product['total_num'],
                        'product_name_text' => ProductSkuModel::getNameById($product['product_sku_id']),
                        'tablet_product_name_text' => ProductSkuModel::getNameById($product['product_sku_id']),
                    ];
                    continue;
                }
                // 总库存
                $productData[] = [
                    'data' => [
                        'product_stock' => [
                            $type,
                            $product['total_num']
                        ]
                    ],
                    'where' => [
                        'product_id' => $product['product_id'],
                    ],
                ];
                $productSkuData[] = [
                    'data' => [
                        'stock_num' => [$type, $product['total_num']],
                        'product_sales' => [$type == 'dec' ? 'inc' : 'dec', $product['total_num']]
                    ],
                    'where' => [
                        'product_id' => $product['product_id'],
                        'product_sku_id' => $product['product_sku_id'],
                    ],
                ];
            }
        }

        if (!empty($error)) {
            return $error;
        }

        try {
            // 更新商品销量
            !empty($productData) && $this->updateProduct($productData);
            // 更新商品sku库存
            !empty($productSkuData) && $this->updateProductSku($productSkuData);
        } catch (\Exception $e) {
            log_write('master updateProductStock' . $e->getMessage());
        }
        return true;
    }
}
