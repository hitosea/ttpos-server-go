<?php

namespace app\shop\model\product;

use app\common\library\helper;
use app\common\model\erp\ErpWarehouseForm;
use app\common\model\erp\ErpWarehouseOutForm;
use app\common\model\product\ProductBom as ProductBomModel;
use app\common\model\erp\ErpMonthlyProductStatistics as ErpMonthlyProductStatisticsModel;

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
                'is_open_stock' => $item['is_open_stock'] ?? 1, // 是否开启库存, 兼容商品默认开启库存
            ]);
            // 判断是否开启授权进销存
            if ($product->hasInventoryAuth()) {
                // 添加规格关联材料
                $materialList = $item['material'] ?? [];
                RelatedMaterial::updateRelatedMaterial($materialList, $flavor['uuid']);

                // 创建"添加入库"记录
                self::addWarehouseInForm($flavor, 1, $data['shop_user_id'], $flavor['stock_num']);
                // erp商品月初库存记录
                ErpMonthlyProductStatisticsModel::newProductRecord($flavor['uuid']);
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
            $isAdd = true; // 是否新增规格
            $oldStockNum = 0; // 旧库存
            $newStockNum = $item['stock_num']; // 新库存
            $flavorData = [
                'purchase_price' => $item['purchase_price'] ?? 0, // 采购单价
                'price' => $item['product_price'], // 销售单价
                'name' => $item['spec_name'], // 规格名称,
                'product_flavor_uuid' => $item['spec_id'], // 规格uuid
                'product_package_uuid' => $product['uuid'], // 产品包uuid
                'stock_num' => $item['stock_num'], // 库存数量
                'barcode_value' => $item['barcode'], // 条码值
                'status' => $data['product_status'] == 10 ? 1 : 0, // 状态: 10-上架, 20-下架
                'is_open_stock' => $item['is_open_stock'] ?? 1, // 是否开启库存, 兼容商品默认开启库存
            ];
            $flavorUuid = $item['product_sku_id'] ?? 0;
            if ($flavorUuid == 0) {
                /** @var ProductBom $flavor */
                $flavor = self::create($flavorData);
            } else {
                $flavor = $product->sku()->where('uuid', $flavorUuid)->find();
                if (!$flavor) {
                    /** @var ProductBom $flavor */
                    $flavor = self::create($flavorData);
                } else {
                    $oldStockNum = $flavor['stock_num'];
                    /** @var ProductBom $flavor */
                    $flavor->save($flavorData);
                    $isAdd = false;
                }
            }
            //
            $flavorUuidList[] = $flavor['uuid'];
            // 判断是否开启授权进销存
            if ($product->hasInventoryAuth()) {
                $materialList = $item['material'] ?? [];
                if (empty($materialList)) {
                    RelatedMaterial::deleteRelatedMaterial($flavor['uuid']);
                } else {
                    // 更新规格关联材料
                    RelatedMaterial::updateRelatedMaterial($materialList, $flavor['uuid']);
                }

                // 如果新增规格，则创建"添加入库"记录
                if ($isAdd) {
                    self::addWarehouseInForm($flavor, 1, $data['shop_user_id'], $flavor->stock_num, $data['stock_remark']);
                    // erp商品月初库存记录
                    ErpMonthlyProductStatisticsModel::newProductRecord($flavor['uuid']);
                } else {
                    $diffStockNum = abs(floatval(helper::bcsub($newStockNum, $oldStockNum, 4)));
                    // 创建"调整入库"记录
                    if ($newStockNum > $oldStockNum) {
                        self::addWarehouseInForm($flavor, 2, $data['shop_user_id'], $diffStockNum, $data['stock_remark']);
                    }
                    // 创建"调整出库"记录
                    if ($newStockNum < $oldStockNum) {
                        self::addWarehouseOutForm($flavor, 1, $data['shop_user_id'], $diffStockNum, $data['stock_remark']);
                    }
                }
            }
        }
        // 删除规格
        if (!empty($flavorUuidList)) {
            $flavorList = self::whereNotIn('uuid', $flavorUuidList)
                ->where('product_package_uuid', $product['uuid'])
                ->where('product_flavor_uuid', '>', 0)
                ->select();
            /** @var ProductBom $flavor */
            foreach ($flavorList as $flavor) {
                // 删除规格关联材料
                RelatedMaterial::deleteRelatedMaterial($flavor['uuid']);
                if ($product->hasInventoryAuth()) {
                    // 创建"删除出库"记录
                    self::addWarehouseOutForm($flavor, 4, $data['shop_user_id'], $flavor['stock_num'], $data['stock_remark']);
                }
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
                $feed = self::create($feedData);
            } else {
                $feed->save($feedData);
            }
            // 加料uuid列表
            $feedUuidList[] = $feed['product_sauce_uuid'];
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

    /**
     * 添加规格入库记录
     * 
     * @param $flavor 规格
     * @param $scene 入库场景: 0-purchase采购入库 1-add添加入库 2-adjust调整入库
     * @param $operatorUuid 入库操作人
     * @param $num 入库数量
     * @param $remark 入库备注
     */
    public static function addWarehouseInForm($flavor, $scene, $operatorUuid, $num, $remark = '', $purchaseOrderUuid = 0)
    {
        $formModel = new ErpWarehouseForm();
        return $formModel->save([
            'form_no' => $formModel->generateInCode(),
            'scene' => $scene,
            'num' => $num,
            'product_bom_uuid' => $flavor['uuid'],
            'operator_uuid' => $operatorUuid,
            'remark' => $remark,
            'purchase_order_uuid' => $purchaseOrderUuid,
        ]);
    }

    /**
     * 添加规格出库记录
     * 
     * @param $flavor 规格
     * @param $scene 出库场景: 0-sales销售出库 1-adjust调整出库 2-loss损耗出库 3-lost丢失出库 4-delete删除出库
     * @param $operatorUuid 操作人uuid
     * @param $num 出库数量
     * @param $remark $num 出库备注
     */
    public static function addWarehouseOutForm($flavor, $scene, $operatorUuid, $num, $remark = '')
    {
        $outFormModel = new ErpWarehouseOutForm();
        return $outFormModel->addOutForm($scene, $operatorUuid, [
            'product_bom_uuid' => $flavor['uuid'],
            'num' => $num,
            'remark' => $remark
        ]);
    }
}