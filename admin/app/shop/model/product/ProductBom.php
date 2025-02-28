<?php

namespace app\shop\model\product;

use app\common\model\product\ProductBom as ProductBomModel;

/**
 * 商家-商品BOM模型
 */
class ProductBom extends ProductBomModel
{
    /**
     * 添加规格
     */
    public static function addFlavor($data, Product $product)
    {
        $flavors = $data['sku'];
        foreach ($flavors as $item) {
            $flavor = self::create([
                'purchase_price' => $item['purchase_price'] ?? 0, // 采购单价
                'price' => $item['product_price'], // 销售单价
                'name' => $item['spec_name'], // 规格名称
                'product_flavor_uuid' => $item['spec_id'], // 规格uuid
                'product_package_uuid' => $product['uuid'], // 产品包uuid
                'stock_num' => $item['stock_num'], // 库存数量
                'barcode_value' => $item['barcode'], // 条码值
                'status' => $product['status'], // 状态
            ]);
            // 判断是否开启授权进销存
            if ($product->hasInventoryAuth()) {
                // 添加规格关联材料
                $materialList = $item['material'] ?? [];
                RelatedMaterial::addRelatedMaterial($materialList, $flavor['uuid']);
            }
        }
    }

    /**
     * 更新规格
     */
    public static function updateFlavor($data, Product $product)
    {
        $flavorUuidList = [];
        // 新增或编辑规格
        $flavorList = $data['sku'];
        foreach ($flavorList as $item) {
            $flavorUuidList[] = $item['product_sku_id'];
            $flavorData = [
                'purchase_price' => $item['purchase_price'] ?? 0, // 采购单价
                'price' => $item['price'], // 销售单价
                'name' => $item['name'], // 规格名称,
                'product_flavor_uuid' => $item['spec_sku_id'], // 规格uuid
                'product_package_uuid' => $product['uuid'], // 产品包uuid
                'stock_num' => $item['stock_num'], // 库存数量
                'barcode_value' => $item['barcode'], // 条码值
                'status' => $data['product_status'] == 10 ? 1 : 0 // 状态: 10-上架, 20-下架
            ];
            $flavor = $product->sku()->where('uuid', $item['product_sku_id'])->find();
            if (!$flavor) {
                $flavor = self::create($flavorData);
            } else {
                $flavor->save($flavorData);
            }
            // 判断是否开启授权进销存
            if ($product->hasInventoryAuth()) {
                $materialList = $item['material'] ?? [];
                // 更新规格关联材料
                RelatedMaterial::updateRelatedMaterial($materialList, $flavor['uuid']);
                // 规格出入库记录
                // $this->productInventoryRecord($productSku, $productSkuOld, $data['sku'], $this['product_id'], $this['type'], $data['stock_remark'] ?? '', $shop_supplier_id, $shopSupplierId);
            }
        }
        // 删除规格
        if (!empty($flavorUuidList)) {
            $flavorList = self::whereNotIn('uuid', $flavorUuidList)
                ->where('product_flavor_uuid', '>', 0)
                ->select();
            foreach ($flavorList as $flavor) {
                // 删除规格关联材料
                RelatedMaterial::deleteRelatedMaterial($flavor['uuid']);
                $flavor->delete();
            }
        }
    }

    /**
     * 添加加料
     */
    public static function addFeed($data, Product $product)
    {
        $feeds = $data['product_feed'] ?? [];
        foreach ($feeds as $feed) {
            self::create([
                'price' => $feed['price'],
                'name' => $feed['feed_name'],
                'product_sauce_uuid' => $feed['feed_id'],
                'product_package_uuid' => $product['uuid'],
                'stock_num' => $feed['stock_num'],
                'status' => $product['status'],
                'is_default_select' => $feed['default_select'],
            ]);
        }
    }

    /**
     * 更新加料
     */
    public static function updateFeed($data, Product $product)
    {
        // 加料uuid列表
        $feedUuidList = [];
        // 新增或编辑加料
        $feedList = $data['product_feed'];
        if (empty($feedList)) {
            self::deleteFeed($product);
            return;
        }
        foreach ($feedList as $item) {
            $feedUuidList[] = $item['feed_id'];
            $feedData = [
                'price' => $item['price'], // 价格
                'name' => $item['feed_name'], // 名称
                'product_sauce_uuid' => $item['feed_id'], // 加料uuid
                'product_package_uuid' => $product['uuid'], // 产品包uuid
                'stock_num' => $item['stock_num'], // 库存数量
                'is_default_select' => $item['default_select'], // 是否默认勾选
                'status' => $data['product_status'] == 10 ? 1 : 0, // 状态: 10-上架, 20-下架
            ];
            $feed = $product->feed()->where('product_sauce_uuid', $item['feed_id'])->find();
            if (!$feed) {
                self::create($feedData);
            } else {
                $feed->save($feedData);
            }
        }
        // 删除加料
        if (!empty($feedUuidList)) {
            $feedList = self::where('product_package_uuid', $product['uuid'])
                ->whereNotIn('product_sauce_uuid', $feedUuidList)
                ->where('product_sauce_uuid', '>', 0)
                ->select();
            foreach ($feedList as $feed) {
                $feed->delete();
            }
        }
    }

    /**
     * 删除加料
     */
    public static function deleteFeed(Product $product)
    {
        $feedList = self::where('product_package_uuid', $product['uuid'])
            ->where('product_sauce_uuid', '>', 0)
            ->select();
        foreach ($feedList as $feed) {
            $feed->delete();
        }
    }
}