<?php

namespace app\common\model_old\product;

use think\facade\Db;
use think\Collection;
use app\common\model_old\BaseModel;
use app\common\model_old\product\ProductAttribute;
use app\common\model_old\product\ProductAttributeGroup;

/**
 * 属性模型
 */
class Attribute extends BaseModel
{
    protected $name = 'attribute';
    protected $pk = 'attribute_id';

    /**
     * 处理多语言
     */
    protected $append = ['attribute_name_text', 'parent_attribute_name_text'];

    /**
     * 属性名称
     */
    public function getAttributeNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['attribute_name']);
    }

    /**
     * 父级属性名称
     */
    public function getParentAttributeNameTextAttr($value, $data = [])
    {
        if (isset($data['parent_attribute_name']) && $data['parent_attribute_name']) {
            return extractLanguage($data['parent_attribute_name']);
        }
        return '';
    }

    /**
     * 设置属性值
     */
    public function setAttributeNameAttr($value)
    {
        return $value && is_array($value) ? json_encode($value) : ($value ?: '');
    }

    /**
     * 关联产品ids
     */
    public function getProductIdsAttr($value, $data = [])
    {
        $product_ids = $data['product_ids'] ?? $value ?? '';
        if (empty($product_ids)) {
            return [];
        }
        $arr = array_map('intval', explode(',', $product_ids));
        return array_values($arr);
    }

    /**
     * 关联产品
     */
    public function productAttribute($attribute_id): Collection
    {
        return $this->alias('attr')
            ->field('product.product_id')
            ->leftJoin('product_attribute pa', 'attr.attribute_id = pa.attribute_id')
            ->leftJoin('product product', 'product.product_id = pa.product_id')
            ->where('product.is_delete', 0)
            ->where('attr.attribute_id', $attribute_id)
            ->select();
    }

    /**
     * 子属性
     */
    public function children()
    {
        return $this->hasMany(Attribute::class, 'parent_id', 'attribute_id');
    }

    /**
     * 父属性
     */
    public function parent()
    {
        return $this->belongsTo(Attribute::class, 'parent_id', 'attribute_id');
    }

    /**
     * 更新属性库
     *
     */
    public function updateAttr($product_id, $data, $shop_supplier_id)
    {
        $del_group_attribute_ids = [];
        $del_attribute_ids = [];
        if ($data) {
            foreach ($data as $item) {
                // 属性组
                $group_attribute_id = 0;
                if (isset($item['parent_id']) && $item['parent_id']) {
                    $attribute = ProductAttributeGroup::where('attribute_id', $item['parent_id'])->where('product_id', $product_id)->where('shop_supplier_id', $shop_supplier_id)->find();
                } else {
                    $parent_attribute = Attribute::where('attribute_name', $item['attribute_name'])->where('parent_id', 0)->where('shop_supplier_id', $shop_supplier_id)->find();
                    $item['parent_id'] = $parent_attribute ? $parent_attribute['attribute_id'] : 0;
                    $attribute = ProductAttributeGroup::where('attribute_id', $item['parent_id'])->where('product_id', $product_id)->where('shop_supplier_id', $shop_supplier_id)->find();
                }
                $updateData = [
                    'attribute_required' => $item['attribute_required'] ?? 0,
                    'attribute_open_max_select' => $item['attribute_open_max_select'] ?? 0,
                    'attribute_max_select' => $item['attribute_max_select'] ?? 0,
                    'shop_supplier_id' => $shop_supplier_id,
                    'app_id' => self::$app_id
                ];
                if ($attribute) {
                    $attribute->save($updateData);
                    $group_attribute_id = $attribute['group_attribute_id'];
                } else {
                    $newAttribute = ProductAttributeGroup::create(array_merge($updateData, [
                        'product_id' => $product_id,
                        'attribute_id' => $item['parent_id']
                    ]));
                    $group_attribute_id = $newAttribute['group_attribute_id'];
                }
                $del_group_attribute_ids[] = $group_attribute_id;

                // 属性值
                if (isset($item['attribute_ids']) && $item['attribute_ids']) {
                    $del_attribute_ids = array_merge($del_attribute_ids, $item['attribute_ids']);
                    foreach ($item['attribute_ids'] as $key => $child_attribute_id) {
                        $child = ProductAttribute::where('attribute_id', $child_attribute_id)->where('product_id', $product_id)->where('shop_supplier_id', $shop_supplier_id)->find();
                        $childUpdateData = [
                            'group_attribute_id' => $group_attribute_id,
                            'default_select' => $item['default_select'][$key] ?? 0,
                            'shop_supplier_id' => $shop_supplier_id,
                            'app_id' => self::$app_id
                        ];

                        if ($child) {
                            $child->where('product_id', $product_id)->save($childUpdateData);
                        } else {
                            ProductAttribute::create(array_merge($childUpdateData, [
                                'product_id' => $product_id,
                                'attribute_id' => $child_attribute_id
                            ]));
                        }
                    }
                }
            }
        }
        // 删除多余的产品关联属性组
        $groupQuery = ProductAttributeGroup::where('product_id', $product_id)->where('shop_supplier_id', $shop_supplier_id);
        if (!empty($del_group_attribute_ids)) {
            $groupQuery->whereNotIn('group_attribute_id', $del_group_attribute_ids);
        }
        $groupQuery->delete();

        // 删除多余的产品关联属性值
        $attributeQuery = ProductAttribute::where('product_id', $product_id)->where('shop_supplier_id', $shop_supplier_id);
        if (!empty($del_attribute_ids)) {
            $attributeQuery->whereNotIn('attribute_id', $del_attribute_ids);
        }
        $attributeQuery->delete();
    }

    /**
     * 获取列表数据
     */
    public function getAllList($shop_supplier_id)
    {
        $prefix = env('DB_PREFIX');
        return $this->alias('a')
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
            ", 'a.attribute_id = pa.attribute_id')
            ->leftJoin('attribute b', 'a.parent_id = b.attribute_id')
            ->where('a.shop_supplier_id', '=', $shop_supplier_id)
            ->order(['a.create_time' => 'desc'])
            ->field('a.*, b.attribute_name as parent_attribute_name')
            ->select();
    }

    /**
     * 详情
     */
    public static function detail($attribute_id)
    {
        return self::find($attribute_id);
    }

    /**
     * 检查是否关联属性值
     */
    public function isUseWithAttributeValue($attribute_id)
    {
        return Attribute::where('parent_id', 'in', $attribute_id)->count() > 0;
    }

    /**
     * 检查是否被关联
     */
    public function isUseWithProduct($attribute_id)
    {
        // 兼容旧数据，先删除产品已删除的关联数据
        ProductAttribute::where('product_id', 'in', function ($query) {
            $query->name('product')->where('is_delete', '=', 1)->field('product_id');
        })->delete();
        // 兼容旧数据，先删除产品已删除的关联属性组
        ProductAttributeGroup::where('product_id', 'in', function ($query) {
            $query->name('product')->where('is_delete', '=', 1)->field('product_id');
        })->delete();
        return ProductAttribute::where('attribute_id', 'in', $attribute_id)->count() > 0;
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null, $lang = 'zh', $parent_id = 0)
    {
        $filter = [
            [Db::raw("JSON_UNQUOTE(JSON_EXTRACT(attribute_name, '$.$lang'))"), '=', $name],
            'shop_supplier_id' => $shop_supplier_id
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['attribute_id', '<>', $id];
        }
        $filter[] = ['parent_id', $parent_id > 0 ? '>' : '=',0];
        return static::where($filter)->value('attribute_id') ? true : false;
    }

    /**
     * 维护产品表中的属性数组
     *
     * @param $total_product_ids 产品ID数组
     */
    public function maintainProductAttribute($total_product_ids, $delete_product_ids = []): bool
    {
        if (!empty($total_product_ids)) {
            $chunks = array_chunk($total_product_ids, 1000);
            foreach ($chunks as $chunk) {
                // 查询属性表
                $product_attributes = ProductAttribute::with([
                    'productAttributeGroup' => function ($query) {
                        $query->field(['group_attribute_id', 'attribute_open_max_select', 'attribute_required', 'attribute_max_select']);
                    },
                    'attribute' => function ($query) {
                        $query->field(['attribute_id', 'attribute_name', 'parent_id'])->with([
                            'parent' => function ($query) {
                                $query->field(['attribute_id', 'attribute_name']);
                            }
                        ]);
                    }
                ])->whereIn('product_id', $chunk)
                    ->field(['product_attribute_id', 'product_id', 'attribute_id', 'default_select', 'group_attribute_id'])
                    ->select()
                    ->toArray();
                // 格式化数据
                $product_attribute_map = [];
                foreach ($product_attributes as $item) {
                    $product_id = $item['product_id'];
                    $attribute = $item['attribute'];
                    $parent_id = $attribute['parent']['attribute_id'] ?? 0;
                    //
                    if (!isset($product_attribute_map[$product_id])) {
                        $product_attribute_map[$product_id] = [];
                    }
                    if (!isset($product_attribute_map[$product_id][$parent_id])) {
                        $product_attribute_map[$product_id][$parent_id] = [
                            'attribute_open_max_select' => $item['productAttributeGroup']['attribute_open_max_select'] ?? 0,
                            'attribute_required'        => $item['productAttributeGroup']['attribute_required'] ?? 0,
                            'attribute_max_select'      => $item['productAttributeGroup']['attribute_max_select'] ?? 0,
                            'parent_id'                 => $parent_id,
                            'attribute_name'            => $attribute['parent']['attribute_name'] ?? '',
                            'attribute_value'           => [],
                            'default_select'            => [],
                            'attribute_ids'             => []
                        ];
                    }
                    $product_attribute_map[$product_id][$parent_id]['attribute_value'][] = $attribute['attribute_name'] ?? '';
                    $product_attribute_map[$product_id][$parent_id]['default_select'][]  = $item['default_select'] ?? 0;
                    $product_attribute_map[$product_id][$parent_id]['attribute_ids'][]   = $item['attribute_id'] ?? 0;
                }
                // 更新产品表
                $prefix = env('DB_PREFIX');
                $product = new Product;
                $product_ids = array_keys($product_attribute_map);
                $product_attributes = array_values($product_attribute_map);
                if (!empty($product_ids)) {
                    $update_sql = "UPDATE {$prefix}product SET product_attr = CASE product_id ";
                    foreach ($product_ids as $index => $product_id) {
                        $product_attr = json_encode(array_values($product_attributes[$index] ?? []));
                        $product_attr = addslashes($product_attr); // 防止SQL注入并确保JSON数据正确转义
                        $update_sql .= "WHEN $product_id THEN '$product_attr' ";
                    }
                    $update_sql .= "END WHERE product_id IN (" . implode(',', $product_ids) . ")";
                    Db::connect($product->getConnection())->execute($update_sql);
                }
                // 如果有全部删除的产品ID，则清空对应的属性数组
                if (!empty($delete_product_ids)) {
                    $delete_product_ids = array_diff($delete_product_ids, $product_ids);
                    if (!empty($delete_product_ids)) {
                        $product->where('product_id', 'in', $delete_product_ids)->update(['product_attr' => '[]']);
                    }
                }
            }
        }
        return true;
    }
}
