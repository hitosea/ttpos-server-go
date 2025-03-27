<?php

namespace app\shop\model\product;

use app\common\model\product\ProductSku as ProductSkuModel;

/**
 * 商品规格模型
 */
class ProductSku extends ProductSkuModel
{
    /**
     * 批量添加商品sku记录
     */
    public function addSkuList($product_id, $sku_list, $productSkuIdList)
    {
        $updateData = [];
        $saveData = [];
        foreach ($sku_list as $item) {
            $data = $item;
            $data['product_id'] = $product_id;
            $data['spec_sku_id'] = $item['spec_id'];
            $data['line_price'] = $item['product_price'];
            $data['base_stock_num'] = $item['stock_num'];
            $data['material_stock'] = $item['material_stock'] ?? 0; // 材料库存
            $data['app_id'] = self::$app_id;

            if (isset($item['product_sku_id']) && $item['product_sku_id'] > 0) {
                $index = 0;
                foreach ($productSkuIdList as $skuId) {
                    if ($skuId == $item['product_sku_id']) {
                        array_splice($productSkuIdList, $index, 1);
                        break;
                    }
                    $index++;
                }
                $updateData[] = [
                    'data' => $data,
                    'where' => [
                        'product_sku_id' => $item['product_sku_id'],
                    ],
                ];
            } else {
                $saveData[] = $data;
            }
        }
        count($updateData) > 0 && $this->updateAll($updateData);
        count($saveData) > 0 && $this->saveAll($saveData);
        //
        if (count($productSkuIdList) > 0) {
            foreach ($this->where('product_sku_id', 'in', $productSkuIdList)->select() as $p) {
                $p->delete();
            }
        }
    }

}
