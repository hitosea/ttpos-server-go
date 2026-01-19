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
        foreach ($attributeList as $key => $item) {
            // 计算默认最大可选（如果未设置）
            $attrValueCount = count($item['attribute_ids'] ?? []);
            $maxSelect = $item['attribute_max_select'] ?? $attrValueCount;
            
            // 新增属性组
            $group = ProductAttributeGroupModel::create([
                'min_selection' => $item['attribute_min_select'] ?? 0, // 最小选择数量
                'max_selection' => $maxSelect, // 最大选择数量
                // 保留 is_must 字段用于兼容旧数据
                'is_must' => ($item['attribute_min_select'] ?? 0) > 0 ? 1 : 0,
                'product_package_uuid' => $product['uuid'], // 产品包uuid
                'product_attribute_group_uuid' => $item['parent_id'], // 属性组uuid
                'sort' => $key // 排序
            ]);
            // 新增属性值
            foreach ($item['attribute_ids'] as $key => $attributeId) {
                ProductAttributeModel::create([
                    'product_package_attribute_group_uuid' => $group['uuid'],
                    'attribute_uuid' => $attributeId,
                    'is_default_selected' => $item['default_select'][$key] ?? 0,
                    'sort' => $key // 排序
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
        foreach ($attributeList as $key => $item) {
            // 属性组
            $group = ProductAttributeGroupModel::where('product_attribute_group_uuid', $item['parent_id'])
                ->where('product_package_uuid', $product['uuid'])
                ->find();
            if (!$group) {
                // 计算默认最大可选（如果未设置）
                $attrValueCount = count($item['attribute_ids'] ?? []);
                $maxSelect = $item['attribute_max_select'] ?? $attrValueCount;
                
                $group = ProductAttributeGroupModel::create([
                    'min_selection' => $item['attribute_min_select'] ?? 0,
                    'max_selection' => $maxSelect,
                    // 保留 is_must 字段用于兼容旧数据
                    'is_must' => ($item['attribute_min_select'] ?? 0) > 0 ? 1 : 0,
                    'product_package_uuid' => $product['uuid'],
                    'product_attribute_group_uuid' => $item['parent_id'],
                    'sort' => $key
                ]);
            } else {
                // 计算默认最大可选（如果未设置）
                $attrValueCount = count($item['attribute_ids'] ?? []);
                $maxSelect = $item['attribute_max_select'] ?? $attrValueCount;
                
                $group->save([
                    'min_selection' => $item['attribute_min_select'] ?? 0,
                    'max_selection' => $maxSelect,
                    // 保留 is_must 字段用于兼容旧数据
                    'is_must' => ($item['attribute_min_select'] ?? 0) > 0 ? 1 : 0,
                    'sort' => $group['sort'] != $key ? $key : $group['sort']
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
                        'sort' => $key
                    ]);
                } else {
                    $attribute->save([
                        'is_default_selected' => $item['default_select'][$key] ?? 0,
                        'sort' => $group['sort'] != $key ? $key : $group['sort']
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
