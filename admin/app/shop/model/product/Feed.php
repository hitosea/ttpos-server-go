<?php

namespace app\shop\model\product;

use think\facade\Env;
use help\ValidateHelp;
use app\common\model\product\ProductBom;
use app\common\service\websocket\Websocket;
use app\common\model\store\MultiLanguageName;
use app\common\model\product\Feed as FeedModel;

/**
 * 加料模型
 */
class Feed extends FeedModel
{
    /**
     * 获取列表数据
     */
    public function getList($data)
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
        // 关联加料材料
        $list = $model->with([
            'relatedMaterial' => [ 'material' ]
        ])->order(['feed.sort' => 'asc', 'feed.create_time' => 'asc'])->paginate($data);

        foreach ($list as $item) {
            $materialList = [];
            foreach ($item['relatedMaterial'] as $relatedMaterial) {
                $materialList[] = [
                    'feed_id' => $item['feed_id'],
                    'material_num' => $relatedMaterial['num'],
                    'materialProduct' => [
                        'product_id' => $relatedMaterial['material_uuid'],
                        'product_material_stock' => $relatedMaterial['material']['stock_num'],
                        'product_name_text' => $relatedMaterial['material']['product_name_text'],
                        'product_unit_text' => $relatedMaterial['material']['product_unit_text'],
                    ],
                ];
            }
            $item['material'] = $materialList;
        }

        return $list;
    }

    /**
     * 添加
     */
    public function add($data)
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
        $this->startTrans();
        try {
            $data['name'] = $data['feed_name'] ?? '';
            $data['multi_language_name_uuid'] = (new MultiLanguageName)->saveNames($data['feed_name']);
            // 获取当前最大的排序值
            $maxSort = $this->where('uuid', '<>', $this['uuid'])->max('sort');
            $data['sort'] = $maxSort + 1;
            // 保存加料
            $this->save($data);
            // 关联加料材料
            $materialList = $data['material'] ?? [];
            foreach ($materialList as $key => $item) {
                $num = $item['material_num'] ?? 0;
                if ($num > self::MAX_MATERIAL_NUM) {
                    $num = self::MAX_MATERIAL_NUM;
                }
                $materialList[$key]['material_num'] = $num;
            }
            RelatedMaterial::updateRelatedMaterial($materialList, $this['uuid']);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
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
        $this->startTrans();
        try {
            $data['name'] = $data['feed_name'] ?? '';
            (new MultiLanguageName)->saveNames($data['feed_name'], $this['multi_language_name_uuid']);
            // 更新加料
            $this->save($data);

            // 关联加料材料
            $materialList = $data['material'] ?? [];
            if (empty($materialList)) {
                RelatedMaterial::deleteRelatedMaterial($this['uuid']);
            } else {
                RelatedMaterial::updateRelatedMaterial($materialList, $this['uuid']);
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
                // 删除关联加料材料
                RelatedMaterial::deleteRelatedMaterial($model['uuid']);
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
            // 获取材料最小库存
            $stock = RelatedMaterial::alias('r')
                ->join('material m', 'r.material_uuid = m.uuid')
                ->field('r.related_uuid, LEAST(FLOOR(MIN(m.stock_num / r.num)), 99999999) AS min_stock_num')
                ->where('r.related_uuid', $feed_id)
                ->find();
            $min_stock_num = $stock['min_stock_num'] ?: self::MAX_MATERIAL_NUM;
            // 删除变动的关系
            if (!empty($delete_product_ids)) {
                $chunks = array_chunk($delete_product_ids, 1000);
                foreach ($chunks as $chunk) {
                    $list = ProductBom::where('product_sauce_uuid', $feed_id)->whereIn('product_package_uuid', $chunk)->select();
                    foreach ($list as $item) {
                        $item->delete();
                    }
                }
            }
            // 添加新关系
            if (!empty($add_product_ids)) {
                $insert_data = [];
                foreach ($add_product_ids as $product_id) {
                    $insert_data[] = [
                        'uuid' => createUuid(),
                        'product_package_uuid' => $product_id,
                        'product_sauce_uuid' => $feed_id,
                        'name'  => $this['name'],
                        'price' => $this['price'],
                        'stock_num' => $min_stock_num,
                        'create_time' => time(),
                        'update_time' => time(),
                        'status' => 1,
                    ];
                }
                ProductBom::insertAll($insert_data);
            }
            $this->commit();
            // 推送
            if (!empty($delete_product_ids) || !empty($add_product_ids)) {
                $msgData = [
                    'type' => 'update',
                    'product_uuid' => 0,
                    'update_time' => time()
                ];
                Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_PRODUCT, 0, $msgData);
            }
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }
}
