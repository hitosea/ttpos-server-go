<?php

namespace app\shop\model_old\product;

use think\facade\Cache;
use app\common\model_old\product\Category as CategoryModel;

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
        // 判断分类名称不能为空
        if (empty($data['name'])) {
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
        $data['app_id'] = self::$app_id;
        $data['create_time'] = time();
        $data['update_time'] = time();
        $category_id = $this->insertGetId($data);
        $this->deleteCache($data['type'] ?? 0, $data['is_special'] ?? 0, $data['shop_supplier_id'] ?? 0);
        return array_merge($data, ['category_id' => $category_id, 'name_text' => extractLanguage($data['name'] ?? '')]);
    }

    /**
     * 编辑记录
     */
    public function edit($data)
    {
        // 判断分类名称不能为空
        if (empty($data['name'])) {
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
        $res = $this->save($data) !== false;
        $this->deleteCache($this['type'], $this['is_special'], $this['shop_supplier_id']);
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
        $where = $this->is_special == 1 ? ['special_id' => $this->category_id] : ['category_id' => $this->category_id];
        // 判断是否存在商品
        if ($productCount = (new Product)->getProductTotal($where)) {
            $this->error = '该分类下存在' . $productCount . '个商品，不允许删除';
            return false;
        }
        // 判断是否存在子分类
        if ($this->where('parent_id', $this['category_id'])->count()) {
            $this->error = '该分类下存在子分类，不允许删除';
            return false;
        }
        //
        $res = $this->delete();
        //
        $res && $this->deleteCache($this['type'], $this['is_special'], $this['shop_supplier_id']);
        //
        if ($res && $this->category_id) {
            foreach (Product::where($where)->where('is_delete', '=', 0)->select() as $product) {
                if ($product->is_special == 1 && $product->special_id == $this->category_id) {
                    $product->update(['special_id' => 0]);
                }
                if ($product->is_special != 1 && $product->category_id == $this->category_id) {
                    $product->update(['category_id' => 0]);
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
        $this->deleteCache($this['type'], $this['is_special'], $this['shop_supplier_id']);
        return $res;
    }

    /**
     * 删除缓存
     */
    public function deleteCache($type, $is_special, $shop_supplier_id)
    {
        Cache::tag('category' . $shop_supplier_id . $is_special . $type)->clear();
        Cache::tag('category' . $shop_supplier_id . (!$is_special ? 1 : 0) . $type)->clear();
        return true;
    }
}
