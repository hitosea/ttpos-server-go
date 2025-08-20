<?php

namespace app\shop\model\product;

use app\common\model\product\ProductAttributeGroup as ProductAttributeGroupModel;
use app\common\model\product\ProductAttribute as ProductAttributeModel;

/**
 * 产品属性模型
 */
class ProductAttribute
{
    /**
     * 添加产品属性
     */
    public static function addAttribute($data, Product $product)
    {
        $attributeList = $data['product_attr'] ?? [];
        foreach ($attributeList as $item) {
            // 新增属性组
            $group = ProductAttributeGroupModel::create([
                'is_must' => $item['attribute_required'] ?? 0, // 是否必选
                'max_selection' => $item['attribute_max_select'] ?? 0, // 最大选择数量
                'product_package_uuid' => $product['uuid'], // 产品包uuid
                'product_attribute_group_uuid' => $item['parent_id'] // 属性组uuid
            ]);
            // 新增属性值
            foreach ($item['attribute_ids'] as $key => $attributeId) {
                ProductAttributeModel::create([
                    'product_package_attribute_group_uuid' => $group['uuid'],
                    'attribute_uuid' => $attributeId,
                    'is_default_selected' => $item['default_select'][$key] ?? 0,
                ]);
            }
        }
    }

    /**
     * 更新属性库
     */
    public static function updateAttribute($data, Product $product)
    {
        $groupUuidList = [];
        $attributeUuidList = [];
        $attributeList = $data['product_attr'] ?? [];
        if (empty($attributeList)) {
            self::deleteAttribute($product);
            return;
        }
        foreach ($attributeList as $item) {
            // 属性组
            $group = ProductAttributeGroupModel::where('product_attribute_group_uuid', $item['parent_id'])
                ->where('product_package_uuid', $product['uuid'])
                ->find();
            if (!$group) {
                $group = ProductAttributeGroupModel::create([
                    'is_must' => $item['attribute_required'] ?? 0,
                    'max_selection' => $item['attribute_max_select'] ?? 0,
                    'product_package_uuid' => $product['uuid'],
                    'product_attribute_group_uuid' => $item['parent_id']
                ]);
            } else {
                $attributeOpenMaxSelect = $item['attribute_open_max_select'] ?? 0;
                $attributeMaxSelect = $item['attribute_max_select'] ?? 0;
                if ($attributeOpenMaxSelect == 0) {
                    $attributeMaxSelect = 0;
                }
                $group->save([
                    'is_must' => $item['attribute_required'] ?? 0,
                    'max_selection' => $attributeMaxSelect,
                ]);
            }
            $groupUuidList[] = $group['uuid'];
            $attributeUuidList[$group['uuid']] = [];
            // 属性值
            foreach ($item['attribute_ids'] as $key => $attributeId) {
                $attribute = ProductAttributeModel::where('product_package_attribute_group_uuid', $group['uuid'])
                    ->where('attribute_uuid', $attributeId)
                    ->find();
                if (!$attribute) {
                    $attribute = ProductAttributeModel::create([
                        'product_package_attribute_group_uuid' => $group['uuid'],
                        'attribute_uuid' => $attributeId,
                        'is_default_selected' => $item['default_select'][$key] ?? 0,
                    ]);
                } else {
                    $attribute->save([
                        'is_default_selected' => $item['default_select'][$key] ?? 0,
                    ]);
                }
                $attributeUuidList[$group['uuid']][] = $attribute['uuid'];
            }
        }
        // 查询需要删除的属性组, 并删除属性组和属性值
        $groupList = ProductAttributeGroupModel::whereNotIn('uuid', $groupUuidList)
            ->with(['productAttribute'])
            ->where('product_package_uuid', $product['uuid'])
            ->select();
        foreach ($groupList as $group) {
            foreach ($group->productAttribute as $attribute) {
                $attribute->delete();
            }
            $group->delete();
        }
        // 删除属性值
        foreach ($attributeUuidList as $groupUuid => $attributeUuids) {
            $attributeList = ProductAttributeModel::where('product_package_attribute_group_uuid', $groupUuid)
                ->whereNotIn('uuid', $attributeUuids)
                ->select();
            foreach ($attributeList as $attribute) {
                $attribute->delete();
            }
        }
    }

    /**
     * 删除产品属性
     */
    public static function deleteAttribute(Product $product)
    {
        foreach ($product->productAttributeGroup as $group) {
            foreach ($group->productAttribute as $attribute) {
                $attribute->delete();
            }
            $group->delete();
        }
    }
}
