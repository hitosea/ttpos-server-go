<?php

namespace app\shop\model\product;

use think\facade\Env;
use help\ValidateHelp;
use app\common\model\product\ProductFeed;
use app\common\model\product\Feed as FeedModel;
use app\common\model\product\ProductFeedMaterial;

/**
 * 加料模型
 */
class Feed extends FeedModel
{
    /**
     * 获取列表数据
     */
    public function getList($data, $shop_supplier_id)
    {
        $prefix = Env::get('DB_PREFIX');
        $model = $this->alias('feed')
            ->field('feed.*')
            ->field("IF(pf.feed_count IS NULL, 0, 1) AS is_used")
            ->field("IFNULL(pf.product_ids, '') AS product_ids")
            ->leftJoin("
                (
                    SELECT pf.feed_id, GROUP_CONCAT(DISTINCT product.product_id) AS product_ids, COUNT(DISTINCT pf.feed_id) AS feed_count
                    FROM {$prefix}product_feed pf
                    LEFT JOIN {$prefix}product product ON pf.product_id = product.product_id
                    WHERE product.is_delete = 0
                    GROUP BY pf.feed_id
                ) pf
            ", 'feed.feed_id = pf.feed_id');
        //
        if (isset($data['feed_name']) && $data['feed_name'] != '') {
            $model = $model->jsonLike('feed.feed_name', trim($data['feed_name']));
        }
        $list = $model->with(['material'])
            ->where('feed.shop_supplier_id', '=', $shop_supplier_id)
            ->order(['feed.create_time' => 'desc'])
            ->paginate($data);
        return $list;
    }

    /**
     * 添加
     */
    public function add($data, $shop_supplier_id)
    {
        if (ValidateHelp::hasEmptyValue($data['feed_name'] ?? '')) {
            $this->error = '加料名称不能为空';
            return false;
        }
        $isExist = $this->where('feed_name', '=', $data['feed_name'])
            ->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        $data['shop_supplier_id'] = $shop_supplier_id;
        $data['app_id']           = self::$app_id;
        $this->save($data);
        // 关联加料材料
        $feedId = $this->feed_id;
        if (isset($data['material']) && !empty($data['material'])) {
            ProductFeedMaterial::destroy(['feed_id' => $feedId]);
            foreach ($data['material'] as $data) {
                // 数量超过处理
                $material_num = $data['material_num'] ?? 0;
                if ($material_num > self::MAX_MATERIAL_NUM) {
                    $material_num = self::MAX_MATERIAL_NUM;
                }
                $material = [
                    'feed_id'          => $feedId,
                    'product_feed_id'  => 0,
                    'material_id'      => $data['product_id'],
                    'material_num'     => $material_num,
                    'shop_supplier_id' => $shop_supplier_id,
                    'app_id'           => self::$app_id,
                ];
                (new ProductFeedMaterial)->save($material);
            }
        }
        return true;
    }

    /**
     * 修改
     */
    public function edit($data)
    {
        if (ValidateHelp::hasEmptyValue($data['feed_name'] ?? '')) {
            $this->error = '加料名称不能为空';
            return false;
        }
        $isExist = $this->where('feed_name', '=', $data['feed_name'])
            ->where('feed_id', '<>', $this['feed_id'])
            ->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        $this->save($data);
        // 关联加料材料
        $feedId = $this['feed_id'];
        ProductFeedMaterial::destroy(['feed_id' => $feedId]);
        if (isset($data['material']) && !empty($data['material'])) {
            foreach ($data['material'] as $item) {
                // 数量超过处理
                $material_num = $item['material_num'] ?? 0;
                if ($material_num > self::MAX_MATERIAL_NUM) {
                    $material_num = self::MAX_MATERIAL_NUM;
                }
                $material = [
                    'feed_id'          => $feedId,
                    'product_feed_id'  => 0,
                    'material_id'      => $item['product_id'],
                    'material_num'     => $material_num,
                    'shop_supplier_id' => $this['shop_supplier_id'],
                    'app_id'           => self::$app_id,
                ];
                (new ProductFeedMaterial)->save($material);
            }
        }
        // 同步产品加料表名称
        ProductFeed::where('feed_id', $feedId)->update(['feed_name' => $data['feed_name']]);
        // 同步产品表中的加料数组
        $this->maintainProductFeed($this->productFeed($feedId)->column('product_id'));
        return true;
    }

    /**
     * 删除
     */
    public function setDelete($feed_id)
    {
        // 判断是否关联产品
        if ($this->isUseWithProduct($feed_id)) {
            $this->error = '该加料下存在商品，不允许删除';
            return false;
        }
        return $this->where('feed_id', 'in', $feed_id)->delete();
    }

    /**
     * 关联菜品
     *
     * @param int $feed_id ID
     * @param array $product_ids 产品ID数组
     * @return bool
     */
    public function relatedProduct($feed_id, $product_ids)
    {
        // 开始事务
        $this->startTrans();
        try {
            // 获取当前关联的产品ID
            $current_product_ids = $this->productFeed($feed_id)->column('product_id') ?: [];
            // 计算需要删除的产品ID
            $delete_product_ids = array_diff($current_product_ids, $product_ids) ?: [];
            // 计算需要新增的产品ID
            $add_product_ids = array_diff($product_ids, $current_product_ids) ?: [];
            // 获取材料最小库存
            $stock = ProductFeedMaterial::alias('pfm')
                ->join('product p', 'pfm.material_id = p.product_id')
                ->field('pfm.feed_id, LEAST(FLOOR(MIN(p.product_material_stock / pfm.material_num)), 99999999) AS min_stock_num')
                ->where('pfm.feed_id', $feed_id)
                ->find();
            $min_stock_num = $stock['min_stock_num'] ?: self::MAX_MATERIAL_NUM;
            // 删除变动的关系
            if (!empty($delete_product_ids)) {
                $chunks = array_chunk($delete_product_ids, 1000);
                foreach ($chunks as $chunk) {
                    ProductFeed::where('feed_id', $feed_id)->whereIn('product_id', $chunk)->delete();
                }
            }
            // 添加新关系
            if (!empty($add_product_ids)) {
                $insert_data = [];
                foreach ($add_product_ids as $product_id) {
                    $insert_data[] = [
                        'product_id'       => $product_id,
                        'feed_id'          => $feed_id,
                        'feed_name'        => $this['feed_name'],
                        'price'            => $this['price'],
                        'stock_num'        => $min_stock_num,
                        'shop_supplier_id' => $this['shop_supplier_id'],
                        'app_id'           => $this['app_id'],
                        'create_time'      => time(),
                        'update_time'      => time(),
                    ];
                }
                productFeed::where('product_id', $product_id)->insertAll($insert_data);
            }
            // 维护产品表中的加料数组
            $total_product_ids = array_unique(array_merge($product_ids, $current_product_ids)) ?: [];
            $this->maintainProductFeed($total_product_ids, $delete_product_ids);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }
}
