<?php

namespace app\common\model\product;

use app\common\library\helper;
use app\common\model\BaseModel;
use app\common\model\product\Feed;
use app\common\model\product\ProductFeedMaterial;

/**
 * 商品Feed模型
 */
class ProductFeed extends BaseModel
{
    protected $name = 'product_feed';
    protected $pk = 'product_feed_id';

    /**
     * 处理多语言
     */
    protected $append = ['feed_name_text'];
    public static function getFeedNameTextAttr($value, $data)
    {
        return extractLanguage($data['feed_name']);
    }

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->belongsTo('app\\common\\model\\product\\Product', 'product_id', 'product_id')->with(['image', 'image.file', 'erpSupplier', 'erpSupplier.purchaser']);
    }

    /**
     * 关联加料
     */
    public function feed()
    {
        return $this->belongsTo('app\\common\\model\\product\\Feed', 'feed_id', 'feed_id');
    }

    /**
     * 产品规格关联材料（一对多）
     */
    public function material()
    {
        return $this->hasMany('app\\common\\model\\product\\ProductFeedMaterial', 'product_feed_id')->with(['materialProduct']);
    }

    //更新加料库库
    public function updateFeed($feedList, $productObj)
    {
        $updateData = [];
        $productId = $productObj->product_id;
        $shopSupplierId = $productObj->shop_supplier_id;
        $existingProductFeedIds = helper::getArrayColumn(($productObj['feed']), 'product_feed_id');
        if ($feedList) {
            foreach ($feedList as $item) {
                $data = $item;
                $data['product_id'] = $productId;
                $data['feed_id'] = $item['feed_id'] ?? 0;
                $data['shop_supplier_id'] = $shopSupplierId;
                $data['app_id'] = self::$app_id;
                // todo 加料库， v1.0.8需求变更，暂时不需要
                // $feedExists = feed::where('feed_name', '=', $item['feed_name'])->count();
                // if ($feedExists == 0) {
                //     $newFeedData = [
                //         'feed_name' => $item['feed_name'],
                //         'price' => $item['price'],
                //     ];
                //     $feedModel = new Feed;
                //     $feedModel->save($newFeedData);
                //     $feedId = $feedModel->feed_id;
                //     $this->addMaterial($item['material'], $feedId, 0, $shopSupplierId);
                // }
                // 产品加料库
                if (isset($item['product_feed_id']) && $item['product_feed_id'] > 0) {
                    $index = 0;
                    foreach ($existingProductFeedIds as $feedId) {
                        if ($feedId == $item['product_feed_id']) {
                            array_splice($existingProductFeedIds, $index, 1);
                            break;
                        }
                        $index++;
                    }
                    $updateData[] = [
                        'data' => $data,
                        'where' => [
                            'product_feed_id' => $item['product_feed_id'],
                        ],
                    ];
                    $this->addMaterial($item['material'], 0, $item['product_feed_id'], $shopSupplierId);
                } else {
                    $productFeedModel = new self;
                    $productFeedModel->save($data);
                    $productFeedId = $productFeedModel->product_feed_id;
                    $this->addMaterial($item['material'], 0, $productFeedId, $shopSupplierId);
                }
            }
            count($updateData) > 0 && $this->updateAll($updateData);
        }
        if (count($existingProductFeedIds) > 0) {
            $feeds = $this->where('product_feed_id', 'in', $existingProductFeedIds)->select();
            foreach ($feeds as $feed) {
                $feed->delete();
            }
            $feedMaterials = ProductFeedMaterial::where('product_feed_id', 'in', $existingProductFeedIds)->select();
            foreach ($feedMaterials as $material) {
                $material->delete();
            }
        }
    }

    /**
     * 新增加料材料
     */
    public function addMaterial($material, $feedId = 0, $productFeedId = 0, $shopSupplierId = 0)
    {
        if ($feedId > 0) {
            $feedMaterials = ProductFeedMaterial::where('feed_id', '=', $feedId)->select();
            foreach ($feedMaterials as $m) {
                $m->delete();
            }
        } else {
            $feedMaterials = ProductFeedMaterial::where('product_feed_id', '=', $productFeedId)->select();
            foreach ($feedMaterials as $m) {
                $m->delete();
            }
        }
        if (isset($material) && !empty($material)) {
            $materialData = [];
            foreach ($material as $item) {
                $materialData[] = [
                    'feed_id' => $feedId,
                    'product_feed_id' => $productFeedId,
                    'material_id' => $item['material_id'] ?? $item['materialProduct']['product_id'] ?? 0,
                    'material_num' => $item['material_num'] ?? 0,
                ];
            }

            if (!empty($materialData)) {
                (new ProductFeedMaterial)->saveAll($materialData);
            }
        }
    }

    /**
     * 详情
     */
    public static function detail($id)
    {
        return self::with(['product', 'material'])->find($id);
    }
}
