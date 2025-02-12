<?php

namespace app\shop\model\product;

use think\facade\Env;
use help\ValidateHelp;
use app\common\model\product\ProductAttribute;
use app\common\model\product\ProductAttributeGroup;
use app\common\model\product\Attribute as AttributeModel;

/**
 * 规格/属性(组)模型
 */
class Attribute extends AttributeModel
{
    /**
     * 获取列表数据
     */
    public function getList($data, $shop_supplier_id)
    {
        $prefix = Env::get('DB_PREFIX');
        $model = $this->alias('a')
            ->field('a.*')
            ->field("IF(pa.attribute_count IS NULL, 0, 1) AS is_used")
            ->field("IFNULL(pa.product_ids, '') AS product_ids")
            ->leftJoin("
                (
                    SELECT pa.attribute_id, GROUP_CONCAT(DISTINCT product.product_id) AS product_ids, COUNT(DISTINCT pa.attribute_id) AS attribute_count
                    FROM {$prefix}product_attribute pa
                    LEFT JOIN {$prefix}product product ON pa.product_id = product.product_id
                    WHERE product.is_delete = 0
                    GROUP BY pa.attribute_id
                ) pa
            ", 'a.attribute_id = pa.attribute_id');
        // 名称
        if (isset($data['attribute_name']) && $data['attribute_name'] != '') {
            $model = $model->jsonLike('a.attribute_name', $data['attribute_name']);
        }
        // 类型
        if (isset($data['type']) && $data['type']) {
            if ($data['type'] == 1) {
                $model = $model->where('a.parent_id', 0);
            } else {
                $model = $model->where('a.parent_id', '>', 0);
            }
        }
        // 父级ids
        if (isset($data['parent_ids']) && $data['parent_ids'] !== '') {
            $model = $model->where('a.parent_id', 'in', $data['parent_ids']);
        }
        // 关联查询父级名称
        $model = $model->alias('a')
            ->with(['children'])
            ->leftJoin('attribute parent', 'a.parent_id = parent.attribute_id')
            ->field('a.*, parent.attribute_name as parent_attribute_name')
            ->where('a.shop_supplier_id', $shop_supplier_id)
            ->order(['a.create_time' => 'desc']);
        return $model->paginate($data);
    }

    /**
     * 添加
     */
    public function add($data, $shop_supplier_id)
    {
        if (ValidateHelp::hasEmptyValue($data['attribute_name'] ?? '')) {
            $this->error = '属性名称不能为空';
            return false;
        }
        $parent_id = $data['parent_id'] ?? 0;
        $attribute_name = is_array($data['attribute_name']) ? json_encode($data['attribute_name']) : ($data['attribute_name'] ?: '');
        $isExist        = $this->where('parent_id', $parent_id > 0 ? '>' : '=', 0)->where('attribute_name', $attribute_name)->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        $data['shop_supplier_id'] = $shop_supplier_id;
        $data['app_id']           = self::$app_id;
        return $this->save($data);
    }

    /**
     * 修改
     */
    public function edit($data)
    {
        if (ValidateHelp::hasEmptyValue($data['attribute_name'] ?? '')) {
            $this->error = '属性名称不能为空';
            return false;
        }
        $parent_id = $this['parent_id'] ?? 0;
        $attribute_name = is_array($data['attribute_name']) ? json_encode($data['attribute_name']) : ($data['attribute_name'] ?: '');
        $isExist        = $this->where('parent_id', $parent_id)->where('attribute_name', $attribute_name)->where('attribute_id', '<>', $this['attribute_id'])->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        $this->save($data);
        // 更新关联产品表中的属性数组
        $attribute_id = $this['attribute_id'];
        $this->maintainProductAttribute($this->productAttribute($attribute_id)->column('product_id'));
        return true;
    }

    /**
     * 删除
     */
    public function setDelete($attribute_id)
    {
        if ($this->isUseWithAttributeValue($attribute_id)) {
            $this->error = '该属性下存在属性值，不允许删除';
            return false;
        }
        if ($this->isUseWithProduct($attribute_id)) {
            $this->error = '该属性下存在商品，不允许删除';
            return false;
        }
        return $this->where('attribute_id', 'in', $attribute_id)->delete();
    }

    /**
     * 关联菜品
     *
     * @param int $attribute_id ID
     * @param array $product_ids 产品ID数组
     * @return bool
     */
    public function relatedProduct($attribute_id, $product_ids)
    {
        $this->startTrans();
        try {
            // 获取当前关联的产品ID
            $current_product_ids = $this->productAttribute($attribute_id)->column('product_id') ?: [];
            // 计算需要删除的产品ID
            $delete_product_ids = array_diff($current_product_ids, $product_ids) ?: [];
            // 计算需要新增的产品ID
            $add_product_ids = array_diff($product_ids, $current_product_ids) ?: [];
            // 删除变动的关系
            if (!empty($delete_product_ids)) {
                $chunks = array_chunk($delete_product_ids, 1000);
                foreach ($chunks as $chunk) {
                    // 删除属性值
                    ProductAttribute::where('attribute_id', $attribute_id)->whereIn('product_id', $chunk)->delete();
                    // 删除无属性值的属性组
                    $prefix = env('DB_PREFIX');
                    ProductAttributeGroup::whereIn('product_id', $chunk)
                        ->whereRaw('group_attribute_id NOT IN (SELECT DISTINCT group_attribute_id FROM ' . $prefix . 'product_attribute WHERE product_id IN (' . implode(',', $chunk) . '))')
                        ->delete();
                }
            }
            // 添加新关系
            if (!empty($add_product_ids)) {
                $insert_data = [];
                $parent_attribute_id = $this['parent_id'] ?? 0; // 父级属性ID
                foreach ($add_product_ids as $product_id) {
                    // 先查看是否存在属性组
                    $group_attribute = ProductAttributeGroup::where('product_id', $product_id)->where('attribute_id', $parent_attribute_id)->field('group_attribute_id')->find();
                    if (empty($group_attribute)) {
                        $group_attribute_insert_data = [
                            'product_id'       => $product_id,
                            'attribute_id'     => $parent_attribute_id,
                            'create_time'      => time(),
                            'update_time'      => time(),
                        ];
                        $group_attribute_id = ProductAttributeGroup::insertGetId($group_attribute_insert_data);
                    } else {
                        $group_attribute_id = $group_attribute['group_attribute_id'];
                    }
                    //
                    $insert_data[] = [
                        'product_id'         => $product_id,
                        'group_attribute_id' => $group_attribute_id,
                        'attribute_id'       => $attribute_id,
                        'shop_supplier_id'   => $this['shop_supplier_id'],
                        'app_id'             => $this['app_id'],
                        'create_time'        => time(),
                        'update_time'        => time(),
                    ];
                }
                ProductAttribute::where('product_id', $product_id)->insertAll($insert_data);
            }
            // 维护产品表中的属性数组
            $total_product_ids = array_unique(array_merge($product_ids, $current_product_ids)) ?: [];
            $this->maintainProductAttribute($total_product_ids, $delete_product_ids);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }
}
