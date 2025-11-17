<?php

namespace app\common\model\product;

use think\facade\Db;
use think\Collection;
use help\ValidateHelp;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\service\websocket\Websocket;
use app\common\model\store\MultiLanguageName;
use app\common\model\product\ProductAttribute;
use app\common\model\product\ProductAttributeGroup;

/**
 * 属性模型
 */
class AttributeGroup extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_attribute_group';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 处理多语言
     */
    protected $append = ['attribute_id', 'attribute_name', 'parent_id', 'attribute_name_text', 'parent_attribute_name_text'];

    /**
     * 分类更新后推送通知
     */
    public static function onAfterWrite(AttributeGroup $model)
    {
        $msgData = [
            'type' => 'update',
            'product_uuid' => 0,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_CATEGORY, 0, $msgData);
    }

    /**
     * 分类删除后推送通知
     */
    public static function onAfterDelete(AttributeGroup $model)
    {
        $msgData = [
            'type' => 'delete',
            'product_uuid' => 0,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_CATEGORY, 0, $msgData);
    }
    
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
    public function getParentIdAttr($value, $data = [])
    {
        return 0;
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
            ->field('product.product_id')
            ->leftJoin('product_attribute pa', 'attr.attribute_id = pa.attribute_id')
            ->leftJoin('product product', 'product.product_id = pa.product_id')
            ->where('product.delete_time', 0)
            ->where('attr.attribute_id', $attribute_id)
            ->select();
    }

    /**
     * 子属性
     */
    public function children()
    {
        return $this->hasMany(Attribute::class, 'attribute_group_uuid', 'uuid');
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
                    $attribute = ProductAttributeGroup::where('attribute_id', $item['parent_id'])->where('product_id', $product_id)->find();
                } else {
                    $parent_attribute = Attribute::where('attribute_name', $item['attribute_name'])->where('parent_id', 0)->find();
                    $item['parent_id'] = $parent_attribute ? $parent_attribute['attribute_id'] : 0;
                    $attribute = ProductAttributeGroup::where('attribute_id', $item['parent_id'])->where('product_id', $product_id)->find();
                }
                $updateData = [
                    'attribute_required' => $item['attribute_required'] ?? 0,
                    'attribute_open_max_select' => $item['attribute_open_max_select'] ?? 0,
                    'attribute_max_select' => $item['attribute_max_select'] ?? 0,
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
                        $child = ProductAttribute::where('attribute_id', $child_attribute_id)->where('product_id', $product_id)->find();
                        $childUpdateData = [
                            'group_attribute_id' => $group_attribute_id,
                            'default_select' => $item['default_select'][$key] ?? 0,
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
        $groupQuery = ProductAttributeGroup::where('product_id', $product_id);
        if (!empty($del_group_attribute_ids)) {
            $groupQuery->whereNotIn('group_attribute_id', $del_group_attribute_ids);
        }
        $groupQuery->delete();

        // 删除多余的产品关联属性值
        $attributeQuery = ProductAttribute::where('product_id', $product_id);
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
        return $this->alias('a')
            ->field('a.*')
            ->order(['a.create_time' => 'desc'])
            ->field('a.*, b.attribute_name as parent_attribute_name')
            ->select();
    }

    /**
     * 详情
     */
    public static function detail($attribute_id)
    {
        return self::where('uuid', $attribute_id)->find();
    }

    /**
     * 检查是否关联属性值
     */
    public function isUseWithAttributeValue($attribute_id)
    {
        return Attribute::where('attribute_group_uuid', 'in', $attribute_id)->count() > 0;
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
     * 修改
     */
    public function edit($data)
    {
        if (ValidateHelp::hasEmptyValue($data['attribute_name'] ?? '')) {
            $this->error = '属性名称不能为空';
            return false;
        }
        $attribute_name = is_array($data['attribute_name']) ? json_encode($data['attribute_name']) : ($data['attribute_name'] ?: '');
        //
        $data['name'] = $attribute_name;
        $data['multi_language_name_uuid'] = (new MultiLanguageName())->saveNames($attribute_name, $this['multi_language_name_uuid']);
        $this->save($data);
        return true;
    }

    /**
     * 删除
     */
    public function setDelete($attribute_id)
    {
        // 检查是否存在属性值
        if ($this->isUseWithAttributeValue($attribute_id)) {
            $this->error = '该属性下存在属性值，不允许删除';
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
}
