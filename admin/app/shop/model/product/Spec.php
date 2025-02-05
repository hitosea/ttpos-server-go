<?php

namespace app\shop\model\product;

use help\ValidateHelp;
use app\common\model\product\ProductSku;
use app\common\model\product\Spec as SpecModel;
use app\common\model\product\ProductSkuMaterial;

/**
 * 规格/属性(组)模型
 */
class Spec extends SpecModel
{
    /**
     * 获取列表数据
     */
    public function getList($data, $shop_supplier_id)
    {
        $prefix = env('DB_PREFIX');
        $model = $this->alias('sku')
            ->field('sku.*')
            ->field("IF(psku.sku_count IS NULL, 0, 1) AS is_used")
            ->field("IFNULL(psku.product_ids, '') AS product_ids")
            ->leftJoin("
                (
                    SELECT psku.spec_sku_id, GROUP_CONCAT(DISTINCT product.product_id) AS product_ids, COUNT(DISTINCT psku.spec_sku_id) AS sku_count
                    FROM {$prefix}product_sku psku
                    LEFT JOIN {$prefix}product product ON psku.product_id = product.product_id
                    WHERE product.is_delete = 0
                    GROUP BY psku.spec_sku_id
                ) psku
            ", 'sku.spec_id = psku.spec_sku_id');
        //
        if (isset($data['spec_name']) && $data['spec_name'] != '') {
            $model = $model->jsonLike('sku.spec_name', $data['spec_name']);
        }
        $list = $model
            ->with('material')
            ->where('sku.shop_supplier_id', '=', $shop_supplier_id)
            ->order(['sku.create_time' => 'desc'])
            ->paginate($data);

        return $list;
    }

    /**
     * 新增规格组
     */
    public function add($data, $shop_supplier_id)
    {
        if (ValidateHelp::hasEmptyValue($data['spec_name'] ?? '')) {
            $this->error = '规格名称不能为空';
            return false;
        }
        $isExist = $this->where('shop_supplier_id', '=', $shop_supplier_id)
            ->where('spec_name', '=', $data['spec_name'])
            ->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        $data['shop_supplier_id'] = $shop_supplier_id;
        $data['app_id'] = self::$app_id;
        $this->save($data);

        // 关联规格材料
        $specId = $this->spec_id;
        if (isset($data['material']) && !empty($data['material'])) {
            ProductSkuMaterial::where('spec_id', '=', $specId)->delete();
            foreach ($data['material'] as $data) {
                $material = [
                    'spec_id' => $specId,
                    'product_sku_id' => 0,
                    'material_id' => $data['product_id'],
                    'material_num' => $data['material_num'] ?? 0,
                    'shop_supplier_id' => $shop_supplier_id,
                    'app_id' => self::$app_id,
                ];
                (new ProductSkuMaterial)->save($material);
            }
        }
        return array_merge($data, ['spec_id' => $specId]);
    }

    /**
     * 修改
     */
    public function edit($data)
    {
        if (ValidateHelp::hasEmptyValue($data['spec_name'] ?? '')) {
            $this->error = '规格名称不能为空';
            return false;
        }
        $isExist = $this->where('shop_supplier_id', '=', $this['shop_supplier_id'])
            ->where('spec_name', '=', $data['spec_name'])
            ->where('spec_id', '<>', $this['spec_id'])
            ->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        $this->save($data);
        // 关联规格材料
        $specId = $this['spec_id'];
        ProductSkuMaterial::where('spec_id', '=', $specId)->delete();
        if (isset($data['material']) && !empty($data['material'])) {
            foreach ($data['material'] as $data) {
                $material = [
                    'spec_id' => $specId,
                    'product_sku_id' => 0,
                    'material_id' => $data['product_id'],
                    'material_num' => $data['material_num'] ?? 0,
                    'shop_supplier_id' => $this['shop_supplier_id'],
                    'app_id' => self::$app_id,
                ];
                (new ProductSkuMaterial)->save($material);
            }
        }
        // 同步产品规格表名称
        ProductSku::where('spec_sku_id', $specId)->update(['spec_name' => $data['spec_name']]);
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
        return $this->where('spec_id', 'in', $spec_id)->delete();
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
            // 删除变动的关系
            if (!empty($delete_product_ids)) {
                ProductSku::where('spec_sku_id', $spec_id)->whereIn('product_id', $delete_product_ids)->delete();
            }
            // 添加新关系
            if (!empty($add_product_ids)) {
                $insert_data = [];
                foreach ($add_product_ids as $product_id) {
                    $insert_data[] = [
                        'product_id' => $product_id,
                        'spec_sku_id' => $spec_id,
                        'spec_name' => $this['spec_name'],
                        'stock_num' => 99999999,
                        'app_id' => $this['app_id'],
                        'create_time' => time(),
                        'update_time' => time(),
                    ];
                }
                ProductSku::where('product_id', $product_id)->insertAll($insert_data);
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
                ProductSku::where('product_sku_id', $product['product_sku_id'])->where('spec_sku_id', $this['spec_id'])->update(['product_price' => $product['product_price']]);
                // 获取产品规格最低价
                $product_id = $product['product_id'] ?? 0;
                $product_price = ProductSku::where('product_id', $product_id)->min('product_price');
                Product::update(['product_price' => $product_price], ['product_id' => $product_id]);
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
     * 规格产品列表
     *
     * @param int $spec_id
     * @return array
     */
    public function skuProduct($spec_id)
    {
        $list = $this->alias('a')
            ->leftJoin('product_sku ps', 'a.spec_id = ps.spec_sku_id')
            ->leftJoin('product p', 'ps.product_id = p.product_id')
            ->leftJoin('category c', 'p.category_id = c.category_id')
            ->where('a.spec_id', $spec_id)
            ->field('a.spec_id, a.spec_name, ps.product_sku_id, ps.product_id, ps.product_price as product_sku_price, p.product_name, p.category_id, c.name as category_name')
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
