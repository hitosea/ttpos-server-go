<?php

namespace app\shop\model\product;

use think\facade\Env;
use help\ValidateHelp;
use app\common\service\websocket\Websocket;
use app\common\model\product\AttributeGroup;
use app\common\model\store\MultiLanguageName;
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
        // 属性组
        $type = $data['type'] ?? 1;
        if ($type == 1) {
            return AttributeGroup::with(['children' => function ($query) {
                $query->withAttr('id', function ($value, $data) {
                    return $data['uuid'] ?: 0;
                });
            }])
            ->withAttr('id', function ($value, $data) {
                return $data['uuid'] ?: 0;
            })
            ->order(['sort' => 'asc', 'create_time' => 'asc'])
            ->paginate($data);
        }

        // 属性值
        $prefix = Env::get('DB_PREFIX');
        $model = $this->alias('a')
            ->withAttr('id', function ($value, $data) {
                return $data['uuid'] ?: 0;
            })
            ->field('a.*')
            ->field("IFNULL(pa.product_ids, '') AS product_ids")
            ->field("IF(pa.attribute_count IS NULL, 0, 1) AS is_used")
            ->leftJoin("
                (
                    SELECT pp.attribute_uuid, pp.attribute_uuid as attribute_id, GROUP_CONCAT(DISTINCT product.uuid) AS product_ids, COUNT(DISTINCT pp.attribute_uuid) AS attribute_count
                    FROM {$prefix}product_package_attribute pp
                    LEFT JOIN {$prefix}product_package_attribute_group pag ON pp.product_package_attribute_group_uuid = pag.uuid
                    LEFT JOIN {$prefix}product_package product ON pag.product_package_uuid = product.uuid
                    WHERE product.delete_time = 0 AND pp.delete_time = 0 AND pag.delete_time = 0
                    GROUP BY pp.attribute_uuid
                ) pa
            ", 'a.uuid = pa.attribute_uuid');

        // 名称
        if (isset($data['attribute_name']) && $data['attribute_name'] != '') {
            $model = $model->jsonLike('a.name', $data['attribute_name']);
        }
        // 父级ids
        if (isset($data['parent_ids']) && $data['parent_ids'] !== '') {
            $model = $model->where('a.attribute_group_uuid', 'in', $data['parent_ids']);
        }
        // 关联查询父级名称
        $model = $model->alias('a')
            ->leftJoin('product_attribute_group parent', 'a.attribute_group_uuid = parent.uuid')
            ->field('a.*, parent.name as parent_attribute_name')
            ->order(['parent.sort' => 'asc', 'parent.create_time' => 'asc', 'a.sort' => 'asc', 'a.create_time' => 'asc']);
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
        $model = null;
        if ($parent_id > 0) {
            // 获取最大排序
            $maxSort = $this->where('attribute_group_uuid', $parent_id)->max('sort');
            $data['sort'] = $maxSort + 1;
            $model = $this;
            $isExist = $this->where('attribute_group_uuid', '>', 0)->where('name', $attribute_name)->count();
            if ($isExist) {
                $this->error = '名称已存在';
                return false;
            }
        } else {
            $model = new AttributeGroup;
            $maxSort = $model->max('sort');
            $data['sort'] = $maxSort + 1;
            $isExist = $model->where('name', $attribute_name)->count();
            if ($isExist) {
                $this->error = '名称已存在';
                return false;
            }
        }
        //
        $data['name'] = $attribute_name;
        $data['multi_language_name_uuid'] = (new MultiLanguageName())->saveNames($attribute_name);
        $data['attribute_group_uuid'] = $parent_id;
        return $model->save($data);
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
        $parent_id = $this['attribute_group_uuid'] ?? 0;
        $attribute_name = is_array($data['attribute_name']) ? json_encode($data['attribute_name']) : ($data['attribute_name'] ?: '');
        $isExist = $this->where('name', $attribute_name)->where('uuid', '<>', $this['uuid'])->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        //
        $data['name'] = $attribute_name;
        $data['multi_language_name_uuid'] = (new MultiLanguageName())->saveNames($attribute_name, $this['multi_language_name_uuid']);
        $data['attribute_group_uuid'] = $parent_id;
        $this->save($data);
        return true;
    }

    /**
     * 删除
     */
    public function setDelete($attribute_id)
    {
        if ($this->isUseWithProduct($attribute_id)) {
            $this->error = '该属性下存在商品，不允许删除';
            return false;
        }

        $this->startTrans();
        try {
            // 删除多语言数据
            $models = $this->whereIn('uuid', $attribute_id)->select();
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
     * @param int $attribute_id ID
     * @param array $product_ids 产品ID数组
     * @return bool
     */
    public function relatedProduct($attribute_id, $product_ids)
    {
        $prefix = Env::get('DB_PREFIX');
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
                    $attributeList =  ProductAttribute::alias('pa')
                        ->field('pa.*')
                        ->leftJoin('product_package_attribute_group pag', 'pa.product_package_attribute_group_uuid = pag.uuid')
                        ->leftJoin('product_package product', 'pag.product_package_uuid = product.uuid')
                        ->where('product.delete_time', 0)
                        ->where('pa.attribute_uuid', $attribute_id)
                        ->whereIn('product.uuid', $chunk)
                        ->select();

                    foreach ($attributeList as $attribute) {
                        $attribute->force()->delete();
                    }

                    // 删除无属性值的属性组
                    $attributeGroupList = ProductAttributeGroup::alias('pag')
                        ->whereNotExists(function ($query) use ($prefix) {
                            $query->table($prefix . 'product_package_attribute')
                                ->where('product_package_attribute_group_uuid = pag.uuid');
                        })
                        ->whereIn('pag.product_package_uuid', $chunk)
                        ->select();
                    foreach ($attributeGroupList as $attributeGroup) {
                        $attributeGroup->force()->delete();
                    }
                }
            }
            // 添加新关系
            if (!empty($add_product_ids)) {
                $insert_data = [];
                $parent_attribute_id = $this['attribute_group_uuid'] ?? 0; // 父级属性ID
                foreach ($add_product_ids as $product_id) {
                    // 先查看是否存在属性组
                    $group_attribute = ProductAttributeGroup::where('product_package_uuid', $product_id)->where('product_attribute_group_uuid', $parent_attribute_id)->field('uuid')->find();
                    if (empty($group_attribute)) {
                        $group_attribute_insert_data = [
                            'product_package_uuid' => $product_id,
                            'product_attribute_group_uuid' => $parent_attribute_id,
                            'create_time'          => time(),
                            'update_time'          => time(),
                        ];
                        $group_attribute = new ProductAttributeGroup();
                        $group_attribute->save($group_attribute_insert_data);
                        $group_attribute_id = $group_attribute->uuid;
                    } else {
                        $group_attribute_id = $group_attribute['uuid'];
                    }
                    //
                    $insert_data[] = [
                        'product_package_attribute_group_uuid' => $group_attribute_id,
                        'attribute_uuid'       => $attribute_id,
                        'create_time'        => time(),
                        'update_time'        => time(),
                    ];
                }
                (new ProductAttribute())->saveAll($insert_data);
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
