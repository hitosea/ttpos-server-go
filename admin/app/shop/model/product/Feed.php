<?php

namespace app\shop\model\product;

use think\facade\Env;
use help\ValidateHelp;
use app\common\model\product\ProductBom;
use app\common\model\product\ProductFeed;
use app\common\model\store\MultiLanguageName;
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
            ->field("IF(pb.sku_count IS NULL, 0, 1) AS is_used")
            ->field("IFNULL(pb.product_ids, '') AS product_ids")
            ->leftJoin("
                (
                    SELECT pb.product_sauce_uuid, GROUP_CONCAT(DISTINCT pb.product_package_uuid) AS product_ids, COUNT(DISTINCT pb.product_sauce_uuid) AS sku_count
                    FROM {$prefix}product_bom pb
                    WHERE pb.delete_time = 0 AND pb.product_sauce_uuid > 0
                    GROUP BY pb.product_sauce_uuid
                ) pb
            ", 'feed.uuid = pb.product_sauce_uuid');
        //
        if (isset($data['feed_name']) && $data['feed_name'] != '') {
            $model = $model->jsonLike('feed.name', trim($data['feed_name']));
        }
        // todo 兼容
        $list = $model->with(['material'])->order(['feed.create_time' => 'desc'])->paginate($data);
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
        $isExist = $this->where('name', '=', $data['feed_name'])->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        //
        $data['name'] = $data['feed_name'] ?? '';
        $data['multi_language_name_uuid'] = (new MultiLanguageName)->saveNames($data['feed_name']);
        $this->save($data);
        // todo 关联加料材料
        // $feedId = $this->feed_id;
        // if (isset($data['material']) && !empty($data['material'])) {
        //     ProductFeedMaterial::destroy(['feed_id' => $feedId]);
        //     foreach ($data['material'] as $data) {
        //         // 数量超过处理
        //         $material_num = $data['material_num'] ?? 0;
        //         if ($material_num > self::MAX_MATERIAL_NUM) {
        //             $material_num = self::MAX_MATERIAL_NUM;
        //         }
        //         $material = [
        //             'feed_id'          => $feedId,
        //             'product_feed_id'  => 0,
        //             'material_id'      => $data['product_id'],
        //             'material_num'     => $material_num,
        //         ];
        //         (new ProductFeedMaterial)->save($material);
        //     }
        // }
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
        $isExist = $this->where('name', '=', $data['feed_name'])
            ->where('uuid', '<>', $this['uuid'])
            ->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        //
        $data['name'] = $data['feed_name'] ?? '';
        (new MultiLanguageName)->saveNames($data['feed_name'], $this['multi_language_name_uuid']);
        $this->save($data);
        // todo 兼容 关联加料材料
        // $feedId = $this['feed_id'];
        // ProductFeedMaterial::destroy(['feed_id' => $feedId]);
        // if (isset($data['material']) && !empty($data['material'])) {
        //     foreach ($data['material'] as $item) {
        //         // 数量超过处理
        //         $material_num = $item['material_num'] ?? 0;
        //         if ($material_num > self::MAX_MATERIAL_NUM) {
        //             $material_num = self::MAX_MATERIAL_NUM;
        //         }
        //         $material = [
        //             'feed_id'          => $feedId,
        //             'product_feed_id'  => 0,
        //             'material_id'      => $item['product_id'],
        //             'material_num'     => $material_num,
        //         ];
        //         (new ProductFeedMaterial)->save($material);
        //     }
        // }
        // // 同步产品加料表名称
        // ProductFeed::where('feed_id', $feedId)->update(['feed_name' => $data['feed_name']]);
        // // 同步产品表中的加料数组
        // $this->maintainProductFeed($this->productFeed($feedId)->column('product_id'));
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
        $this->startTrans();
        try {
            // 删除多语言数据
            $models = $this->whereIn('uuid', $feed_id)->select();
            foreach ($models as $model) {
                if ($model['multi_language_name_uuid']) {
                    (new MultiLanguageName)->where('uuid', $model['multi_language_name_uuid'])->find()?->delete();
                }
                $model->delete();
            }
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
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
            // todo 兼容 获取材料最小库存
            // $stock = ProductFeedMaterial::alias('pfm')
            //     ->join('product p', 'pfm.material_id = p.product_id')
            //     ->field('pfm.feed_id, LEAST(FLOOR(MIN(p.product_material_stock / pfm.material_num)), 99999999) AS min_stock_num')
            //     ->where('pfm.feed_id', $feed_id)
            //     ->find();
            // $min_stock_num = $stock['min_stock_num'] ?: self::MAX_MATERIAL_NUM;
            $min_stock_num = self::MAX_MATERIAL_NUM;
            // 删除变动的关系
            if (!empty($delete_product_ids)) {
                $chunks = array_chunk($delete_product_ids, 1000);
                foreach ($chunks as $chunk) {
                    $list = ProductBom::where('product_sauce_uuid', $feed_id)->whereIn('product_package_uuid', $chunk)->select();
                    foreach ($list as $item) {
                        $item->force()->delete();
                    }
                }
            }
            // 添加新关系
            if (!empty($add_product_ids)) {
                $insert_data = [];
                foreach ($add_product_ids as $product_id) {
                    $insert_data[] = [
                        'uuid' => createUuid(),
                        'product_package_uuid'       => $product_id,
                        'product_sauce_uuid'          => $feed_id,
                        'name'        => $this['name'],
                        'price'            => $this['price'],
                        'stock_num'        => $min_stock_num,
                        'create_time'      => time(),
                        'update_time'      => time(),
                    ];
                }
                ProductBom::insertAll($insert_data);
            }
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }
}
