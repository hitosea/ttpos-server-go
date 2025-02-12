<?php

namespace app\shop\model\product;

use help\StringHelp;
use help\ValidateHelp;
use think\facade\Cache;
use app\common\library\helper;
use app\shop\service\CheckService;
use app\common\model\file\UploadFile;
use app\common\model\product\ProductTax;
use app\common\model\product\ProductFeed;
use app\common\model\erp\ErpInventoryRecord;
use app\common\model\product\ProductSkuMaterial;
use app\common\model\product\ProductFeedMaterial;
use app\common\model\erp\ErpMonthlyProductStatistics;
use app\common\model\product\Product as ProductModel;
use \app\common\model\buffet\BuffetProduct as BuffetProductModel;

/**
 * 商品模型
 */
class Product extends ProductModel
{
    /**
     * 添加商品
     */
    public function add($data)
    {
        if (!isset($data['type']) || !in_array($data['type'], [ProductModel::TYPE_PRODUCT, ProductModel::TYPE_MATERIAL])) {
            $this->error = '商品类型不能为空';
            return false;
        }
        $product_name = isset($data['product_name']) ? $data['product_name'] : '';
        if (ValidateHelp::hasEmptyValue($product_name)) {
            $this->error = '商品名称不能为空';
            return false;
        }
        //
        [$status, $msg] = ValidateHelp::hasExceedLength($product_name, 150);
        if ($status === true) {
            $this->error = '商品名称长度不能超过150个字符';
            $this->errorData = $msg;
            return false;
        }
        // 商品名称唯一性
        if (CheckService::checkNameExist('product', $product_name, $data['shop_supplier_id'])) {
            $this->error = '商品名称已存在';
            return false;
        }
        $data['content'] = isset($data['content']) ? $data['content'] : '';
        $data['alone_grade_equity'] = isset($data['alone_grade_equity']) ? json_decode($data['alone_grade_equity'], true) : '';
        $data['app_id'] = self::$app_id;
        // 规格
        if (isset($data['sku']) && is_array($data['sku']) && !empty($data['sku'])) {
            // 初始化条码错误数组和条码去重检查
            $firstError = [];
            $barcodes = array_column($data['sku'], 'barcode');
            $barcodeCounts = array_count_values(array_filter($barcodes));
            $duplicateBarcodes = array_filter($barcodeCounts, fn($count) => $count > 1);
            foreach ($data['sku'] as &$info) {
                // 如果同一个条码出现超过1次，标记为错误
                $errorMsg1 = '商品条码不能重复';
                $barcodeError1 = true;
                if (!isset($firstError[$errorMsg1])) {
                    $firstError[$errorMsg1] = [];
                }
                if ($info['barcode'] && isset($duplicateBarcodes[$info['barcode']])) {
                    $barcodeError1 = false;
                }
                $firstError[$errorMsg1][] = $barcodeError1;

                // 条码格式验证
                $errorMsg2 = '商品条码只能为数字、字母、数字+字母组合';
                $barcodeError2 = true;
                if (!isset($firstError[$errorMsg2])) {
                    $firstError[$errorMsg2] = [];
                }
                if ($info['barcode'] && !preg_match('/^[0-9a-zA-Z]+$/', $info['barcode'])) {
                    $barcodeError2 = false;
                }
                $firstError[$errorMsg2][] = $barcodeError2;

                // 条码唯一性验证
                $errorMsg3 = '商品条码已存在';
                $barcodeError3 = true;
                if (!isset($firstError[$errorMsg3])) {
                    $firstError[$errorMsg3] = [];
                }
                if ($info['barcode'] && CheckService::checkNameExist('product_barcode', $info['barcode'], 0)) {
                    $barcodeError3 = false;
                }
                $firstError[$errorMsg3][] = $barcodeError3;

                // 处理数据超过最大值时，返回提示信息
                if ($text = $this->alertProductData($info)) {
                    $this->error = $text;
                    return false;
                }

                // 处理数据格式
                $info = $this->sanitizeProductData($info);
            }
            unset($info);
            // 处理错误
            $firstError = array_filter($firstError, fn($item) => in_array(false, $item, true));
            if (!empty($firstError)) {
                $this->error = key($firstError);
                $this->errorData = reset($firstError);
                return false;
            }
        }
        // 属性
        if (isset($data['product_attr']) && is_array($data['product_attr']) && !empty($data['product_attr'])) {
            $attr = $data['product_attr'][0];
            // 最多默认勾选数量
            $defaultSelectCount = count(array_filter($attr['default_select'], function ($item) {
                return $item == 1;
            }));
            if (
                $attr['attribute_open_max_select'] == 1 &&
                isset($attr['attribute_value']) &&
                isset($attr['attribute_max_select']) &&
                $defaultSelectCount > $attr['attribute_max_select']
            ) {
                $this->error = '不能超过最多可选数量' . ' ' . $attr['attribute_max_select'];
                return false;
            }
        }
        $data = $this->sanitizeProductData($data);
        // 加料
        if (isset($data['product_feed']) && is_array($data['product_feed']) && !empty($data['product_feed'])) {
            // 最多默认勾选数量
            $defaultSelectCount = count(array_filter($data['product_feed'], function ($item) {
                return $item == 1;
            }));
            if (
                $data['feed_open_max_select'] == 1 &&
                isset($data['feed_max_select']) &&
                $defaultSelectCount > $data['feed_max_select']
            ) {
                $this->error = '不能超过最多可选数量' . ' ' . $data['feed_max_select'];
                return false;
            }
            foreach ($data['product_feed'] as &$item) {
                if (!isset($item['uuid']) || empty($item['uuid'])) {
                    $item['uuid'] = StringHelp::getGuidV4();
                }
                $item = $this->sanitizeProductData($item);
            }
            unset($item);
        }
        //
        if (isset($data['productTaxes']) && is_array($data['productTaxes']) && count($data['productTaxes']) > 2) {
            $this->error = '产品税类只能设置2条';
            return false;
        }
        // 判断图片id是否存在
        $images = isset($data['image']) ? $data['image'] : [];
        $imageIds = array_map(function ($image) {
            return isset($image['file_id']) ? $image['file_id'] : $image['image_id'];
        }, $images);
        $existingImageIds = UploadFile::whereIn('file_id', $imageIds)->column('file_id');
        $missingImageIds = array_diff($imageIds, $existingImageIds);
        if ($missingImageIds) {
            $this->error = '商品图片不存在';
            return false;
        }
        // 开启事务
        $this->startTrans();
        try {
            // 添加商品
            $data['product_attr'] = isset($data['product_attr']) ? $data['product_attr'] : '';
            $data['product_feed'] = isset($data['product_feed']) ? $data['product_feed'] : '';
            $this->save($data);
            // 更新产品关联税类
            $this->updateProductTaxes($this['product_id'], $data['productTaxes'] ?? []);
            // 商品规格
            $this->addProductSpec($data);
            // 商品图片
            $this->addProductImages($data['image']);
            // 更新属性
            (new Attribute)->updateAttr($this['product_id'], $data['product_attr'], $data['shop_supplier_id']);
            // 更新加料
            (new ProductFeed)->updateFeed($data['product_feed'], $this);
            // 更新单位
            (new Unit)->updateUnit($data['product_unit'], $data['shop_supplier_id']);
            // erp商品月初库存记录
            (new ErpMonthlyProductStatistics)->newProductRecord($this['product_id']);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 处理数据超过最大值时，返回提示信息
     */
    private function alertProductData($data)
    {
        $limits = [
            'price' => ['limit' => 1000000, 'message' => '价格不能超过1000000'],
            'product_price' => ['limit' => 1000000, 'message' => '价格不能超过1000000'],
            'stock_num' => ['limit' => 99999999, 'message' => '库存不能超过99999999']
        ];

        foreach ($limits as $key => $value) {
            if (array_key_exists($key, $data) && $data[$key] > $value['limit']) {
                return $value['message'];
            }
        }
        return '';
    }

    /**
     * 处理数据为负数时，自动转换为0
     */
    private function sanitizeProductData($data)
    {
        $keys = [
            'price',
            'product_price',
            'sales_initial',
            'product_sort',
            'line_price',
            'supplier_price',
            'bag_price',
            'cost_price',
            'min_buy',
            'limit_num',
            'purchase_price',
            'material_num',
            'stock_num',
        ];

        foreach ($keys as $key) {
            if (array_key_exists($key, $data)) {
                $data[$key] = max(0, $data[$key]);
            }
        }
        return $data;
    }

    /**
     * 添加商品图片
     */
    private function addProductImages($images)
    {
        foreach ($this->image()->select() as $image) {
            $image->delete();
        }
        $data = array_map(function ($images) {
            return [
                'image_id' => isset($images['file_id']) ? $images['file_id'] : $images['image_id'],
            ];
        }, $images);
        return $this->image()->saveAll($data);
    }

    /**
     * 编辑商品
     */
    public function edit($data)
    {
        if (!isset($data['type']) || !in_array($data['type'], [ProductModel::TYPE_PRODUCT, ProductModel::TYPE_MATERIAL])) {
            $this->error = '商品类型不能为空';
            return false;
        }
        $product_name = isset($data['product_name']) ? $data['product_name'] : '';
        if (ValidateHelp::hasEmptyValue($data['product_name'] ?? '')) {
            $this->error = '商品名称不能为空';
            return false;
        }
        //
        [$status, $msg] = ValidateHelp::hasExceedLength($product_name, 150);
        if ($status === true) {
            $this->error = '商品名称长度不能超过150个字符';
            $this->errorData = $msg;
            return false;
        }
        // 商品名称唯一性
        if (CheckService::checkNameExist('product', $product_name, $this['shop_supplier_id'] ?? 0, $this['product_id'] ?? 0)) {
            $this->error = '商品名称已存在';
            return false;
        }

        $data['spec_type'] = isset($data['spec_type']) ? $data['spec_type'] : $this['spec_type'];
        $data['content'] = isset($data['content']) ? $data['content'] : '';
        $data['alone_grade_equity'] = isset($data['alone_grade_equity']) ? json_decode($data['alone_grade_equity'], true) : '';
        $productSkuIdList = helper::getArrayColumn(($this['sku']), 'product_sku_id');
        // 规格
        if (isset($data['sku']) && is_array($data['sku']) && !empty($data['sku'])) {
            // 初始化条码错误数组和条码去重检查
            $firstError = [];
            $barcodes = array_column($data['sku'], 'barcode');
            $barcodeCounts = array_count_values(array_filter($barcodes));
            $duplicateBarcodes = array_filter($barcodeCounts, fn($count) => $count > 1);
            foreach ($data['sku'] as &$info) {
                // 如果同一个条码出现超过1次，标记为错误
                $errorMsg1 = '商品条码不能重复';
                $barcodeError1 = true;
                if (!isset($firstError[$errorMsg1])) {
                    $firstError[$errorMsg1] = [];
                }
                if ($info['barcode'] && isset($duplicateBarcodes[$info['barcode']])) {
                    $barcodeError1 = false;
                }
                $firstError[$errorMsg1][] = $barcodeError1;

                // 条码格式验证
                $errorMsg2 = '商品条码只能为数字、字母、数字+字母组合';
                $barcodeError2 = true;
                if (!isset($firstError[$errorMsg2])) {
                    $firstError[$errorMsg2] = [];
                }
                if ($info['barcode'] && !preg_match('/^[0-9a-zA-Z]+$/', $info['barcode'])) {
                    $barcodeError2 = false;
                }
                $firstError[$errorMsg2][] = $barcodeError2;

                // 条码唯一性验证
                $errorMsg3 = '商品条码已存在';
                $barcodeError3 = true;
                if (!isset($firstError[$errorMsg3])) {
                    $firstError[$errorMsg3] = [];
                }
                if ($info['barcode'] && CheckService::checkNameExist('product_barcode', $info['barcode'], 0, $info['product_sku_id'] ?? 0)) {
                    $barcodeError3 = false;
                }
                $firstError[$errorMsg3][] = $barcodeError3;

                // 处理数据超过最大值时，返回提示信息
                if ($text = $this->alertProductData($info)) {
                    $this->error = $text;
                    return false;
                }

                // 处理数据格式
                $info = $this->sanitizeProductData($info);
            }
            unset($info);
            // 处理错误
            $firstError = array_filter($firstError, fn($item) => in_array(false, $item, true));
            if (!empty($firstError)) {
                $this->error = key($firstError);
                $this->errorData = reset($firstError);
                return false;
            }
        }
        // 属性
        if (isset($data['product_attr']) && is_array($data['product_attr']) && !empty($data['product_attr'])) {
            $attr = $data['product_attr'][0];
            // 最多默认勾选数量
            $defaultSelectCount = count(array_filter($attr['default_select'], function ($item) {
                return $item == 1;
            }));
            if (
                $attr['attribute_open_max_select'] == 1 &&
                isset($attr['attribute_value']) &&
                isset($attr['attribute_max_select']) &&
                $defaultSelectCount > $attr['attribute_max_select']
            ) {
                $this->error = '不能超过最多可选数量' . ' ' . $attr['attribute_max_select'];
                return false;
            }
        }
        $data = $this->sanitizeProductData($data);
        // 加料
        if (isset($data['product_feed']) && is_array($data['product_feed']) && !empty($data['product_feed'])) {
            // 最多默认勾选数量
            $defaultSelectCount = count(array_filter($data['product_feed'], function ($item) {
                return $item == 1;
            }));
            if (
                $data['feed_open_max_select'] == 1 &&
                isset($data['feed_max_select']) &&
                $defaultSelectCount > $data['feed_max_select']
            ) {
                $this->error = '不能超过最多可选数量' . ' ' . $data['feed_max_select'];
                return false;
            }
            foreach ($data['product_feed'] as &$item) {
                if (!isset($item['uuid']) || empty($item['uuid'])) {
                    $item['uuid'] = StringHelp::getGuidV4();
                }
                $item = $this->sanitizeProductData($item);
            }
            unset($item);
        } else {
            $data['feed_required'] = 0;
            $data['feed_open_max_select'] = 0;
            $data['feed_max_select'] = 0;
        }
        //
        if (isset($data['productTaxes']) && is_array($data['productTaxes']) && count($data['productTaxes']) > 2) {
            $this->error = '产品税类只能设置2条';
            return false;
        }
        // 判断图片id是否存在
        $images = isset($data['image']) ? $data['image'] : [];
        $imageIds = array_map(function ($image) {
            return isset($image['file_id']) ? $image['file_id'] : $image['image_id'];
        }, $images);
        $existingImageIds = UploadFile::whereIn('file_id', $imageIds)->column('file_id');
        $missingImageIds = array_diff($imageIds, $existingImageIds);
        if ($missingImageIds) {
            $this->error = '商品图片不存在';
            return false;
        }
        //
        return $this->transaction(function () use ($data, $productSkuIdList) {
            $data['product_attr'] = isset($data['product_attr']) ? $data['product_attr'] : '';
            $data['product_feed'] = isset($data['product_feed']) ? $data['product_feed'] : '';
            $this->save($data);
            // 更新产品关联税类
            $this->updateProductTaxes($this['product_id'], $data['productTaxes'] ?? []);
            // 商品规格
            $this->addProductSpec($data, $productSkuIdList);
            // 商品图片
            $this->addProductImages($data['image']);
            // 更新属性
            (new Attribute)->updateAttr($this['product_id'], $data['product_attr'], $this['shop_supplier_id']);
            // 更新加料
            (new ProductFeed)->updateFeed($data['product_feed'], $this);
            // 更新单位
            (new Unit)->updateUnit($data['product_unit'], $this['shop_supplier_id']);
            //
            return true;
        });
    }

    /**
     * 更新产品关联税类
     */
    private function updateProductTaxes($productId, $productTax)
    {
        if (empty($productTax)) {
            return;
        }
        $model = new ProductTax;
        $model->destroy(['product_id' => $productId]);
        $data = [];
        foreach ($productTax as $item) {
            $data[] = [
                'product_id' => $productId,
                'product_tax_type' => $item['product_tax_type'] ?? 0,
                'tax_category_id' => $item['tax_category_id'] ?? 0,
            ];
        }
        $model->saveAll($data);
    }

    /**
     * 添加商品规格
     */
    private function addProductSpec($data, $productSkuIdList = [])
    {
        // 更新模式: 先删除所有规格
        $model = new ProductSku;
        $stock = 0; //总库存
        $materialStock = 0; //总材料库存
        $product_price = 0; //价格
        $cost_price = 0;
        $bag_price = 0;
        $shop_supplier_id = isset($data['shop_supplier_id']) ? $data['shop_supplier_id'] : $this['shop_supplier_id'];
        $productSkuOld = ProductSku::where('product_id', '=', $this['product_id'])->select()->toArray();

        // 添加规格数据
        if ($data['spec_type'] == '10') {
            // 规格名称新增或更新
            (new Spec)->updateSpec($data['sku']);
            // 单规格
            $sku = $data['sku'][0] ?? [];
            $sku['app_id'] = self::$app_id;
            $sku['line_price'] = $sku['product_price'] ?? 0;
            $sku['product_price'] = $sku['product_price'] ?? 0;
            $sku['stock_num'] = $sku['stock_num'] ?? 0;
            unset($sku['create_time']);
            unset($sku['update_time']);
            //
            $skus = $this->sku()->where('product_sku_id', '<>', $sku['product_sku_id'] ?? 0)->select();
            foreach ($skus as $skuDel) {
                $skuDel->delete();
            }
            $this->sku()->save($sku);
            //
            $stock = $sku['stock_num'] ?? 0;
            $materialStock = $sku['material_stock'] ?? 0;
            $product_price = $sku['product_price'] ?? 0;
            $cost_price = $sku['cost_price'] ?? 0;
            $bag_price = $sku['bag_price'] ?? 0;
        } else if ($data['spec_type'] == '20') {
            // 规格名称新增或更新
            (new Spec)->updateSpec($data['sku']);
            //
            $model->addSkuList($this['product_id'], $data['sku'], $productSkuIdList); // 添加商品sku
            //
            $product_price = $data['sku'][0]['product_price'] ?? 0;
            $cost_price = $data['sku'][0]['cost_price'] ?? 0;
            $bag_price = $data['sku'][0]['bag_price'] ?? 0;
            //
            foreach ($data['sku'] as $item) {
                $stock += (int)$item['stock_num'] ?? 0;
                $materialStock += (int)($item['material_stock'] ?? 0);
                if (($item['product_price'] ?? 0) < $product_price) {
                    $product_price = $item['product_price'] ?? 0;
                }
                if (($item['cost_price'] ?? 0) < $cost_price) {
                    $cost_price = $item['cost_price'] ?? 0;
                }
                if (($item['bag_price'] ?? 0) < $bag_price) {
                    $bag_price = $item['bag_price'] ?? 0;
                }
            }
        }
        // 判断是否开启授权进销存
        if ($this->hasInventoryAuth()) {
            $shopSupplierId = $data['shop_user_id'] ?? 0;
            $productSku = ProductSku::where('product_id', '=', $this['product_id'])->select()->toArray();
            // 关联产品规格材料
            foreach ($productSku as $item) {
                $materials = ProductSkuMaterial::where('product_sku_id', '=', $item['product_sku_id'] ?? 0)->select();
                foreach ($materials as $material) {
                    $material->delete();
                }
                if (isset($item['material']) && !empty($item['material'])) {
                    foreach ($item['material'] as $material) {
                        $material = [
                            'spec_id' => 0, // 产品规格不需要规格id
                            'product_sku_id' => $item['product_sku_id'] ?? 0,
                            'material_id' => $material['product_id'],
                            'material_num' => $material['material_num'] ?? 0,
                        ];
                        (new ProductSkuMaterial)->save($material);
                    }
                }
            }
            // 规格出入库记录
            $this->productInventoryRecord($productSku, $productSkuOld, $data['sku'], $this['product_id'], $this['type'], $data['stock_remark'] ?? '', $shop_supplier_id, $shopSupplierId);
        }
        //
        $this->save([
            'product_stock' => $stock,
            'product_material_stock' => $materialStock,
            'product_price' => $product_price,
            'line_price' => $product_price,
            'cost_price' => $cost_price,
            'bag_price' => $bag_price
        ]);
        // 如果是编辑材料，则更新所有产品的总库存、产品规格库存、加料库存
        if ($this['type'] == ProductModel::TYPE_MATERIAL && count($productSkuIdList) > 0) {
            $this->reCalProductStock(array_unique([$this['product_id']]));
        }
    }

    /**
     * 产品规格出入库记录
     */
    public function productInventoryRecord($productSku, $productSkuOld, $skuArrData, $productId, $productType, $stockRemark, $shopSupplierId, $shopUserId)
    {
        // 关联产品规格材料
        foreach ($skuArrData as $sku) {
            foreach ($productSku as &$product) {
                if (isset($sku['spec_name']) && $product['spec_name'] == $sku['spec_name']) {
                    $product['material'] = $sku['material'] ?? [];
                }
            }
        }
        //
        $addIds = [];
        $existingProductSkuOldIds = helper::getArrayColumn($productSkuOld, 'product_sku_id');
        foreach ($productSku as $item) {
            // 新增的规格
            if (!in_array($item['product_sku_id'] ?? 0, $existingProductSkuOldIds)) {
                $addIds[] = $item['product_sku_id'] ?? 0;
            }
        }
        // 删除的规格
        $existingProductSkuNewIds = helper::getArrayColumn($productSku, 'product_sku_id');
        foreach ($productSkuOld as $item) {
            if (!in_array($item['product_sku_id'] ?? 0, $existingProductSkuNewIds)) {
                $this->deleteInventoryRecord($item, $stockRemark, $shopSupplierId, $shopUserId);
            }
        }
        $productSkuOld = helper::arrayColumn2Key($productSkuOld, 'product_sku_id');
        foreach ($productSku as $item) {
            $productSkuId = $item['product_sku_id'] ?? 0;
            // 关联产品规格材料
            ProductSkuMaterial::where('product_sku_id', '=', $productSkuId)->delete();
            if (isset($item['material']) && !empty($item['material'])) {
                $materialIds = [];
                foreach ($item['material'] as $material) {
                    $materialIds = array_merge($materialIds, [$material['product_id']]);
                    $material = [
                        'spec_id' => 0, // 产品规格不需要规格id
                        'product_sku_id' => $productSkuId,
                        'material_id' => $material['product_id'],
                        'material_num' => $material['material_num'] ?? 0,
                    ];
                    (new ProductSkuMaterial)->save($material);
                }
                // 更新跟材料相关的所有产品总库存、产品规格库存、加料库存
                $this->reCalProductStock(array_unique($materialIds));
            }
            // 出入库记录，每个规格都记录
            $inventoryRecordData = [
                'product_id' => $productId,
                'product_sku_id' => $productSkuId,
                'product_sku_name' => $item['spec_name'],
                'operator_id' => $shopUserId ?? 0,
                'remark' => $stockRemark ?? '',
            ];

            $oldStockNum = 0;
            $newStockNum = 0;
            if ($productType == ProductModel::TYPE_PRODUCT) {
                $oldStockNum = $productSkuOld[$productSkuId]['stock_num'] ?? 0;
                $newStockNum = $item['stock_num'] ?? 0;
            } else {
                $oldStockNum = $productSkuOld[$productSkuId]['material_stock'] ?? 0;
                $newStockNum = $item['material_stock'] ?? 0;
            }

            if ($oldStockNum > $newStockNum) {
                // 减少库存 出库记录
                $inventoryRecordData['num'] = $oldStockNum - $newStockNum;
                $inventoryRecordData['type'] = ErpInventoryRecord::TYPE_ADJUST_OUT;
                $inventoryRecordData['status'] = ErpInventoryRecord::STATUS_OUT;
                (new ErpInventoryRecord)->addNew(ErpInventoryRecord::INVENTORY_TYPE_OUT, $inventoryRecordData);
            } else if ($oldStockNum < $newStockNum) {
                // 增加库存 入库记录
                $inventoryRecordData['num'] = $newStockNum - $oldStockNum;
                $inventoryRecordData['type'] = !in_array($item['product_sku_id'] ?? 0, $addIds) ? ErpInventoryRecord::TYPE_ADJUST_IN : ErpInventoryRecord::TYPE_ADJUST_IN_ADD; // 调整入库和添加入库
                $inventoryRecordData['status'] = ErpInventoryRecord::STATUS_IN;
                (new ErpInventoryRecord)->addNew(ErpInventoryRecord::INVENTORY_TYPE_IN, $inventoryRecordData);
            }
        }
    }

    /**
     * 删除出库记录
     */
    public function deleteInventoryRecord($productSku, $stockRemark, $shopSupplierId, $shopUserId)
    {
        $inventoryRecordData = [
            'product_id' => $productSku['product_id'],
            'product_sku_id' => $productSku['product_sku_id'] ?? 0,
            'product_sku_name' => $productSku['spec_name'],
            'operator_id' => $shopUserId,
            'remark' => $stockRemark ?? '',
        ];
        $inventoryRecordData['num'] = $productSku['stock_num'];
        $inventoryRecordData['type'] = ErpInventoryRecord::TYPE_ADJUST_OUT_DEL;
        $inventoryRecordData['status'] = ErpInventoryRecord::STATUS_OUT;
        (new ErpInventoryRecord)->addNew(ErpInventoryRecord::INVENTORY_TYPE_OUT, $inventoryRecordData);
    }

    /**
     * 修改商品状态
     */
    public function setStatus($state)
    {
        return $this->save(['product_status' => $state]) !== false;
    }

    /**
     * 软删除
     */
    public function setDelete($product_ids, $shop_user_id = 0)
    {
        $product_ids = explode(',', $product_ids);
        if (empty($product_ids)) return false;
        $products = $this->where('product_id', 'in', $product_ids)->field(['product_id', 'type', 'shop_supplier_id'])->select();
        // 开启事务
        $this->startTrans();
        try {
            foreach ($products as $product) {
                $product_id = $product['product_id'] ?? 0;
                $type = $product['type'] ?? 0;
                $shop_supplier_id = $product['shop_supplier_id'] ?? 0;
                //
                if ($type == ProductModel::TYPE_MATERIAL) {
                    $is_used = $this->checkMaterialUsed($product['product_id']) ? 1 : 0;
                    if ($is_used) {
                        $this->error = '该材料已被使用，无法删除';
                        return false;
                    }
                }
                //
                $this->where('product_id', '=', $product_id)->delete();
                // 删除自助餐关联产品
                (new BuffetProductModel)->where('product_id', '=', $product_id)->delete();
                // 删除的规格材料
                if ($type == ProductModel::TYPE_MATERIAL) {
                    $skuMaterials = ProductSkuMaterial::where('material_id', '=', $product_id)->select();
                    $productSkuIds = $skuMaterials->column('product_sku_id');
                    $newSkuMaterials = ProductSkuMaterial::where('product_sku_id', 'in', $productSkuIds)->select();
                    foreach ($skuMaterials as &$skuMaterial) {
                        $skuMaterial->delete();
                    }
                    $feedMaterials = ProductFeedMaterial::where('material_id', '=', $product_id)->select();
                    $productFeedIds = $feedMaterials->column('product_feed_id');
                    $newFeedMaterials = ProductFeedMaterial::where('product_feed_id', 'in', $productFeedIds)->select();
                    foreach ($feedMaterials as &$feedMaterial) {
                        $feedMaterial->delete();
                    }

                    // 更新跟材料相关的所有产品总库存、产品规格库存、加料库存
                    $newSkuMaterialIds = $newSkuMaterials->column('material_id');
                    $newFeedMaterialIds = $newFeedMaterials->column('material_id');
                    $materialIds = array_merge($newSkuMaterialIds, $newFeedMaterialIds);
                    $this->reCalProductStock(array_unique($materialIds));
                }
                // 新增规格删除出库记录
                $sku = ProductSku::with(['product'])->where('product_id', '=', $product_id)->select();
                foreach ($sku as $item) {
                    $this->deleteInventoryRecord($item, '', $shop_supplier_id, $shop_user_id);
                }
                // 删除产品关联规格
                ProductSku::where('product_id', '=', $product_id)->delete();
            }
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 获取商品告急数量总数
     */
    public function getProductStockTotal($shop_supplier_id = 0)
    {
        $query = $this->alias('product')
            ->join('product_sku sku', 'sku.product_id = product.product_id')
            ->where('sku.stock_num', '<', condition: 10)
            ->where('product.type', ProductModel::TYPE_PRODUCT)
            ->where('product.is_delete', '=', 0);

        if ($shop_supplier_id > 0) {
            $query = $query->where('product.shop_supplier_id', $shop_supplier_id);
        }

        return count($query->distinct(true)->column('product.product_id'));
    }

    public function getProductId($search)
    {
        $res = $this->like('product_name', $search)->select()->toArray();
        return array_column($res, 'product_id');
    }

    /**
     * 获取数量
     */
    public function getCount($type, $shop_supplier_id, $product_type = 0)
    {
        $model = $this;
        //已下架
        if ($type == 'lower') {
            $model = $model->where('product_status', '=', 20);
        }
        return $model->where('product_type', '=', $product_type)->count();
    }

    /**
     * 查询指定商品
     * @param $value
     */
    public function getProduct($value)
    {
        return $this->with(['image.file'])->where('product_id', 'in', $value)->hidden(['content'])->select();
    }

    /**
     * 同步商品到门店
     * @param $data
     */
    public function transmit($data)
    {
        $product_list = $this->where('product_id', 'in', $data['product_id'])->with(['image', 'sku'])->select();
        // 开启事务
        $this->startTrans();
        try {
            foreach ($product_list as $item) {
                foreach ($data['shop_supplier_id'] as $value) {
                    $product = $this->where('product_id', '=', $item['product_id'])->find();
                    if ($product) {
                        $product->update($item);
                        $this->addImages($item['image']);
                        (new ProductSku)->where('product_id', '=', $product['product_id'])->delete();
                        $this->addProductSku($data['sku']);
                    } else {
                        unset($item['product_id']);
                        unset($item['create_time']);
                        unset($item['update_time']);
                        $item['shop_supplier_id'] = $value;
                        $this->save($item);
                        $this->addImages($item['image']);
                        $this->addProductSku($data['sku']);
                    }
                }
            }
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 添加商品图片
     */
    private function addImages($images)
    {
        foreach ($this->image()->select() as $image) {
            $image->delete();
        }
        $data = array_map(function ($images) {
            return [
                'image_id' => $images['image_id'],
            ];
        }, $images);
        return $this->image()->saveAll($data);
    }

    /**
     * 添加商品规格
     */
    private function addProductSku($data, $isUpdate = false, $productSkuIdList = [])
    {
        // 更新模式: 先删除所有规格
        $model = new ProductSku;
        $stock = 0; //总库存
        $materialStock = 0; //总材料库存
        $product_price = 0; //价格
        $cost_price = 0;
        $bag_price = 0;
        // 添加规格数据
        if ($data['spec_type'] == '10') {
            // 单规格
            $this->sku()->save($data['sku']);
            $stock = $data['sku']['stock_num'] ?? 0;
            $materialStock = $data['sku']['material_stock'] ?? 0;
            $product_price = $data['sku']['product_price'] ?? 0;
            $cost_price = $data['sku']['cost_price'] ?? 0;
            $bag_price = $data['sku']['bag_price'] ?? 0;
        } else if ($data['spec_type'] == '20') {
            //更新规格
            (new Spec)->updateSpec($data['sku']);
            // 添加商品sku
            $model->addSkuList($this['product_id'], $data['sku'], $productSkuIdList);
            $product_price = $data['sku'][0]['product_price'] ?? 0;
            $cost_price = $data['sku'][0]['cost_price'] ?? 0;
            $bag_price = $data['sku'][0]['bag_price'] ?? 0;
            foreach ($data['sku'] as $item) {
                $stock += $item['stock_num'] ?? 0;
                $materialStock += $item['material_stock'] ?? 0;
                if ($item['product_price'] < $product_price) {
                    $product_price = $item['product_price'] ?? 0;
                }
                if ($item['cost_price'] < $cost_price) {
                    $cost_price = $item['cost_price'] ?? 0;
                }
                if ($item['bag_price'] < $bag_price) {
                    $bag_price = $item['bag_price'] ?? 0;
                }
            }
        }
        $this->save([
            'product_stock' => $stock,
            'product_material_stock' => $materialStock,
            'product_price' => $product_price,
            'cost_price' => $cost_price,
            'bag_price' => $bag_price
        ]);
    }
}
