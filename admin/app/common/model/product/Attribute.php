<?php

namespace app\common\model\product;

use think\facade\Db;
use think\Collection;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\product\AttributeGroup;
use app\common\model\product\ProductAttribute;
use app\common\model\product\ProductAttributeGroup;
use app\common\model\product\Attribute as AttributeModel;

/**
 * 属性模型
 */
class Attribute extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_attribute';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     */
    protected $append = ['attribute_id', 'attribute_name', 'attribute_name_text', 'parent_attribute_name_text', 'parent_id'];

    /**
     * 兼容字段
     */
    public function getAttributeIdAttr($value, $data = [])
    {
        return $this->uuid ?: 0;
    }
    public function getAttributeNameAttr($value, $data = [])
    {
        return $this->getData('name') ?: 0;
    }

    /**
     * 父级ID
     */
    public function getParentIdAttr($value, $data = [])
    {
        return $this->getData('attribute_group_uuid') ?: 0;
    }

    /**
     * 属性名称
     */
    public function getAttributeNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['name']);
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
            ->field('product.uuid as product_id')
            ->leftJoin('product_package_attribute pa', 'attr.uuid = pa.attribute_uuid')
            ->leftJoin('product_package_attribute_group pag', 'pa.product_package_attribute_group_uuid = pag.uuid')
            ->leftJoin('product_package product', 'product.uuid = pag.product_package_uuid')
            ->where('product.delete_time', 0)
            ->where('attr.uuid', $attribute_id)
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
     * 获取列表数据
     */
    public function getAllList($shop_supplier_id)
    {
        return $this->order(['create_time' => 'desc'])->select();

        // todo 兼容
        // $prefix = env('DB_PREFIX');
        // return $this->alias('a')
        //     ->field('a.*')
        //     ->field("IF(pa.attribute_count IS NULL, 0, 1) AS is_used")
        //     ->field("IFNULL(pa.product_ids, '') AS product_ids")
        //     ->leftJoin("
        //         (
        //             SELECT pa.attribute_id, GROUP_CONCAT(DISTINCT product.product_id) AS product_ids, COUNT(DISTINCT pa.attribute_id) AS attribute_count
        //             FROM {$prefix}product_attribute pa
        //             LEFT JOIN {$prefix}product_package product ON pa.product_id = product.product_id
        //             GROUP BY pa.attribute_id
        //         ) pa
        //     ", 'a.attribute_id = pa.attribute_id')
        //     ->leftJoin('attribute b', 'a.parent_id = b.attribute_id')
        //     ->order(['a.create_time' => 'desc'])
        //     ->field('a.*, b.attribute_name as parent_attribute_name')
        //     ->select();
    }

    /**
     * 详情
     * @param string $uuid
     */
    public static function detail($uuid)
    {
        return static::where('uuid', $uuid)->find();
    }

    /**
     * 检查是否被关联
     */
    public function isUseWithProduct($attribute_id)
    {
        return ProductAttribute::where('attribute_uuid', 'in', $attribute_id)->count() > 0;
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null, $lang = 'zh')
    {
        $filter = [
            [Db::raw("JSON_UNQUOTE(JSON_EXTRACT(name, '$.$lang'))"), '=', $name],
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['uuid', '<>', $id];
        }
        return static::where($filter)->value('uuid') ? true : false;
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
