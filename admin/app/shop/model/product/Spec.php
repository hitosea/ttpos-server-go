<?php

namespace app\shop\model\product;

use help\ValidateHelp;
use app\common\model\product\ProductBom;
use app\common\model\product\ProductPackageGroupItem;
use app\common\service\websocket\Websocket;
use app\common\model\store\MultiLanguageName;
use app\common\model\product\Spec as SpecModel;
use app\shop\model\product\ProductBom as ProductProductBom;

/**
 * 规格/属性(组)模型
 */
class Spec extends SpecModel
{
    /**
     * 规格更新后推送通知
     */
    public static function onAfterWrite(Spec $model)
    {
        $msgData = [
            'type' => 'update',
            'product_uuid' => 0,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_PRODUCT, 0, $msgData);
    }

    /**
     * 规格删除后推送通知
     */
    public static function onAfterDelete(Spec $model)
    {
        $msgData = [
            'type' => 'update',
            'product_uuid' =>0,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_PRODUCT, 0, $msgData);
    }
    
    /**
     * 获取列表数据
     */
    public function getList($data)
    {
        $prefix = env('DB_PREFIX');
        $model = $this->alias('sku')
            ->field('sku.*')
            ->field("IF(pb.sku_count IS NULL, 0, 1) AS is_used")
            ->field("IFNULL(pb.product_ids, '') AS product_ids")
            ->leftJoin("
                (
                    SELECT pb.product_flavor_uuid, GROUP_CONCAT(DISTINCT pb.product_package_uuid) AS product_ids, COUNT(DISTINCT pb.product_flavor_uuid) AS sku_count
                    FROM {$prefix}product_bom pb
                    WHERE pb.delete_time = 0 AND pb.product_flavor_uuid > 0
                    GROUP BY pb.product_flavor_uuid
                ) pb
            ", 'sku.uuid = pb.product_flavor_uuid');

        //
        if (isset($data['spec_name']) && $data['spec_name'] != '') {
            $model = $model->jsonLike('sku.name', $data['spec_name']);
        }
        $list = $model->order(['sku.create_time' => 'desc'])->paginate($data);
        return $list;
    }

    /**
     * 新增规格组
     */
    public function add($data)
    {
        $name = $data['spec_name'] ?? '';
        if (ValidateHelp::hasEmptyValue($name)) {
            $this->error = '规格名称不能为空';
            return false;
        }
        //
        $data['name'] = $name;
        $data['multi_language_name_uuid'] = (new MultiLanguageName)->saveNames($name);
        $this->save($data);
        $specId = $this->spec_id;
        return array_merge($data, ['spec_id' => $specId]);
    }

    /**
     * 修改
     */
    public function edit($data)
    {
        $name = $data['spec_name'] ?? '';
        if (ValidateHelp::hasEmptyValue($name)) {
            $this->error = '规格名称不能为空';
            return false;
        }
        //
        $data['name'] = $name;
        $data['multi_language_name_uuid'] = (new MultiLanguageName)->saveNames($name, $this['multi_language_name_uuid']);
        $this->save($data);
        // 修改关联商品的规格名称
        ProductProductBom::where('product_flavor_uuid', $this['uuid'])->update(['name' => $name]);
        return true;
    }

    /**
     * 删除
     */
    public function setDelete($spec_id)
    {
        // 判断是否关联产品
        if ($this->isUseWithProduct($spec_id)) {
            $this->error = '该规格下存在商品，不允许删除';
            return false;
        }
        //
        $this->startTrans();
        try {
            // 删除多语言数据
            $models = $this->whereIn('uuid', $spec_id)->select();
            foreach ($models as $model) {
                if ($model['multi_language_name_uuid']) {
                    $model->multiLanguageName->delete();
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
     * @param int $spec_id ID
     * @param array $product_ids 产品ID数组
     * @return bool
     */
    public function relatedProduct($spec_id, $product_ids)
    {
        $this->startTrans();
        try {
            // 获取当前关联的产品ID
            $current_product_ids = $this->productSku($spec_id)->column('product_id');
            // 计算需要删除的产品ID
            $delete_product_ids = array_diff($current_product_ids, $product_ids);
            // 计算需要新增的产品ID
            $add_product_ids = array_diff($product_ids, $current_product_ids);
            // 获取材料最小库存
            // $stock = RelatedMaterial::alias('r')
            //     ->join('material m', 'r.material_uuid = m.uuid')
            //     ->field('r.related_uuid, LEAST(FLOOR(MIN(m.stock_num / r.num)), 99999999) AS min_stock_num')
            //     ->where('r.related_uuid', $spec_id)
            //     ->find();
            // $min_stock_num = $stock['min_stock_num'] ?: Feed::MAX_MATERIAL_NUM;
            // 删除变动的关系
            if (!empty($delete_product_ids)) {
                $packageNames = [];
                $bomIds = ProductBom::where('product_flavor_uuid', $spec_id)->whereIn('product_package_uuid', $delete_product_ids)->column('uuid');
                $packageGroupItems = ProductPackageGroupItem::with([
                    'productPackageGroup' => [
                        'product'
                    ]
                ])->whereIn('product_bom_uuid', $bomIds)->select();
                foreach ($packageGroupItems as $packageGroupItem) {
                    $name = $packageGroupItem['productPackageGroup']['product']['product_name_text'];
                    if (!in_array($name, $packageNames)) {
                        $packageNames[] = "【{$name}】";
                    }
                }
                if (!empty($packageNames)) {
                    $this->error = sprintf(__('套餐%s已使用此规格，不可删除'), implode('', $packageNames));
                    return false;
                }
                if (!ProductBom::destroy(function ($query) use ($spec_id, $delete_product_ids) {
                    $query->where('product_flavor_uuid', $spec_id)->whereIn('product_package_uuid', $delete_product_ids);
                })) {
                    return false;
                }
            }
            // 添加新关系
            if (!empty($add_product_ids)) {
                $insert_data = [];
                foreach ($add_product_ids as $product_id) {
                    $insert_data[] = [
                        'uuid' => createUuid(),
                        'product_package_uuid' => $product_id,
                        'product_flavor_uuid' => $spec_id,
                        'name' => $this['name'],
                        'status' => 1,
                        'create_time' => time(),
                        'update_time' => time(),
                    ];
                }
                ProductBom::insertAll($insert_data);
            }
            // 提交事务
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

    /**
     * 修改价格
     *
     * @param array $data
     * @return bool
     */
    public function batchPrice($data)
    {
        $this->startTrans();
        try {
            foreach ($data['products'] as $product) {
                ProductBom::where('product_package_uuid', $product['product_id'])
                    ->where('product_flavor_uuid', $this['spec_id'])
                    ->update(['price' => $product['product_price']]);
            }
            $this->commit();

            // 推送
            if (!empty($data['products'])) {
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

    /**
     * 规格产品列表
     *
     * @param int $spec_id
     * @return array
     */
    public function skuProduct($spec_id)
    {
        $list = $this->alias('a')
            ->leftJoin('product_bom ps', 'a.uuid = ps.product_flavor_uuid')
            ->leftJoin('product_package p', 'ps.product_package_uuid = p.uuid')
            ->leftJoin('product_category c', 'p.category_uuid = c.uuid')
            ->where('a.uuid', $spec_id)
            ->where('ps.delete_time', '=', 0)
            ->field('a.uuid, a.name, ps.product_package_uuid as product_id, ps.price as product_sku_price, p.name as product_name, p.category_uuid as category_id, c.name as category_name')
            ->select() ?: [];
        // 检查product_id是否为空
        if (empty($list) || empty($list[0]['product_id'])) {
            return [];
        }
        //
        foreach ($list as &$item) {
            $item['product_name_text'] = extractLanguage($item['product_name']);
            unset($item['product_name']);
            $item['category_name_text'] = extractLanguage($item['category_name']);
            unset($item['category_name']);
        }
        return $list;
    }
}
