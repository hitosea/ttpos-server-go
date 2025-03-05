<?php

namespace app\shop\model\product;

use think\facade\Cache;
use app\common\model\store\MultiLanguageName;
use app\common\model\product\Category as CategoryModel;

/**
 * 商品分类模型
 */
class Category extends CategoryModel
{
    /**
     * 添加新记录
     */
    public function add($data)
    {
        $name = $data['name'] ?? '';
        // 判断分类名称不能为空
        if (empty($name)) {
            $this->error = '分类名称不能为空';
            return false;
        }
        // 判断父级分类是否存在
        $parentId = $data['parent_id'] ?? 0;
        $parent = $this->detail($parentId);
        if ($parentId > 0 && !$parent) {
            $this->error = '一级分类不存在';
            return false;
        }
        // 判断排序最大值为999
        if ($data['sort'] > 999) {
            $this->error = '排序最大值为999';
            return false;
        }
        //
        $data['parent_uuid'] = $parentId;
        $data['multi_language_name_uuid'] = (new MultiLanguageName())->saveNames($name);
        $this->save($data);
        $this->deleteCache(1, $data['is_special'] ?? 0, self::$app_id);
        return array_merge($data, ['category_id' => $this->uuid, 'name_text' => extractLanguage($name ?? '')]);
    }

    /**
     * 编辑记录
     */
    public function edit($data)
    {
        $name = $data['name'] ?? '';
        // 判断分类名称不能为空
        if (empty($name)) {
            $this->error = '分类名称不能为空';
            return false;
        }
        // 判断父级分类是否存在
        $parentId = $data['parent_id'] ?? 0;
        if ($parentId > 0 && !$this->detail($parentId)) {
            $this->error = '一级分类不存在';
            return false;
        }
        // 判断父级分类不能与当前分类相同
        if ($this['category_id'] == $parentId && $this['is_button'] == 0) {
            $this->error = '父级分类不能与当前分类相同';
            return false;
        }
        // 判断排序最大值为999
        if ($data['sort'] > 999) {
            $this->error = '排序最大值为999';
            return false;
        }
        //
        !array_key_exists('image_id', $data) && $data['image_id'] = 0;
        $data['parent_uuid'] = $parentId;
        $data['multi_language_name_uuid'] = (new MultiLanguageName())->saveNames($name, $this['multi_language_name_uuid']);
        $res = $this->save($data) !== false;
        $this->deleteCache(1, $this['is_special'], self::$app_id);
        return $res;
    }

    /**
     * 删除商品分类
     */
    public function remove()
    {
        if ($this['category_id'] == 0) {
            $this->error = '不可删除的分类';
            return false;
        }
        $where = $this->is_special == 1 ? ['special_category_uuid' => $this->uuid] : ['category_uuid' => $this->uuid];
        // 判断是否存在商品
        if ($productCount = (new Product)->getProductTotal($where)) {
            $this->error = '该分类下存在' . $productCount . '个商品，不允许删除';
            return false;
        }
        // 判断是否存在子分类
        if ($this->where('parent_uuid', $this['uuid'])->count()) {
            $this->error = '该分类下存在子分类，不允许删除';
            return false;
        }
        //
        $this->multiLanguageName->delete();
        $res = $this->delete();
        // 兼容
        $res && $this->deleteCache(1, $this['is_special'], self::$app_id);
        //
        if ($res && $this->uuid) {
            foreach (Product::where($where)->select() as $product) {
                if ($product->special_category_uuid == $this->uuid) {
                    $product->update(['special_category_uuid' => 0]);
                }
                if ($product->category_uuid == $this->uuid) {
                    $product->update(['category_uuid' => 0]);
                }
            }
        }
        //
        return $res;
    }

    /**
     * 编辑记录
     */
    public function setStatus($data)
    {
        if ($this['category_id'] <= 0) {
            $this->error = '不可操作的分类';
            return false;
        }
        $res = $this->save($data) !== false;
        $this->deleteCache(1, $this['is_special'], self::$app_id);
        return $res;
    }

    /**
     * 删除缓存
     */
    public function deleteCache($type, $is_special, $shop_supplier_id)
    {   
        Cache::tag('category' . $shop_supplier_id . (!$is_special ? 1 : 0) . $type)->clear();
        return true;
    }
}
