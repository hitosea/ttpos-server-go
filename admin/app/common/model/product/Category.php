<?php

namespace app\common\model\product;

use help\StringHelp;
use think\facade\Db;
use think\facade\Env;
use think\facade\Cache;
use app\common\model\BaseModel;
use app\common\model\order\Order;
use think\model\concern\SoftDelete;
use app\common\model\product\Product;
use app\common\model\supplier\Supplier as SupplierModel;

/**
 * 产品分类模型
 */
class Category extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_category';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     */
    protected $append = ['category_id', 'parent_id', 'name_text', 'path_name_text'];

    /**
     * 兼容字段
     */
    public function getCategoryIdAttr($value, $data = [])
    {
        return $this->uuid ?: 0;
    }
    public function getParentIdAttr($value, $data = [])
    {
        return $this->parent_uuid ?: 0;
    }

    /**
     * 多语言关联
     */
    public function multiLanguageName()
    {
        return $this->hasOne('app\common\model\store\MultiLanguageName', 'uuid', 'multi_language_name_uuid');
    }

    /**
     * 处理多语言
     */
    public static function getNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['name'] ?? '');
    }

    /**
     * 全路径分类名称
     */
    public static function getPathNameTextAttr($value, $data = [])
    {
        $text = extractLanguage($value ?: $data['name'] ?? '');
        if (isset($data['parent_uuid']) && $data['parent_uuid'] > 0) {
            try {
                $parentName = self::where('uuid', $data['parent_uuid'])->value('name');
                $parentText = extractLanguage($parentName ?? '');
                $text = $parentText . '-' . $text;
            } catch (\Exception $e) {
            }
        }
        return $text;
    }

    /**
     * 分类图片
     */
    public function images()
    {
        return $this->hasOne('app\\common\\model\\file\\UploadFile', 'file_id', 'image_id');
    }

    public function child()
    {
        return $this->hasMany('app\\common\\model\\product\\Category', 'parent_uuid', 'uuid')->with(['images']);
    }

    /**
     * 详情
     */
    public static function detail($category_id)
    {
        return self::where('uuid', $category_id)->find();
    }


    public function detailWithImage($where)
    {
        return $this->with(['image'])->where($where)->find();
    }

    /**
     * 所有分类
     */
    public static function getAll($type, $is_special, $store = [], $name = '', $is_sort = 1)
    {
        $request = request();
        $page = $request->param('page');
        $list_rows = $request->param('list_rows');
        $isPaginate = ($request->is_paginate !== false && $page != null && $list_rows != null);
        $order_conditions = $is_sort ? ['c.sort' => 'asc', 'c.create_time' => 'desc'] : ['c.create_time' => 'desc'];
        $child_order_conditions = $is_sort ? ['sort' => 'asc', 'create_time' => 'desc'] : ['create_time' => 'desc'];

        $model = new static;
        $cacheKey = 'category_' . $model::$app_id . $type . $is_special . $is_sort . '_' . checkDetect();
        if ($name != '' || $isPaginate || !($result = Cache::get($cacheKey))) {
            $prefix = Env::get('DB_PREFIX');
            $data = $model->alias('c')->with(['images', 'child' => function ($q) use ($name, $child_order_conditions) {
                $q->jsonLike('name', $name);
                $q->order($child_order_conditions);
            }])
                ->where('c.parent_uuid', '=', 0)
                ->where('c.is_special', '=', $is_special)
                ->order($order_conditions)
                ->when($name != '', function ($q) use ($prefix, $name) {
                    $q->jsonLike('c.name', $name);
                    $key = '1';
                    $lang = checkDetect();                                                                                                                                                                          
                    foreach (getSettingLanguages() as $language) {
                        if (($language['name'] ?? '') == $lang) {
                            $key = $language['key'];
                        }
                    }
                    $value = StringHelp::escapeLikeStr($name);
                    $q->whereOrRaw("EXISTS(SELECT * FROM {$prefix}product_category as cc WHERE cc.parent_uuid = c.uuid and JSON_EXTRACT(cc.name , '$.\"$key\"') LIKE '%$value%' )");
                });

            $data = $isPaginate ? $data->paginate(compact('page', 'list_rows')) : $data->select();
            $all = !empty($data) ? $data->toArray() : [];
            $result = $all;
            if ($name == '' && !$isPaginate) {
                Cache::tag('category' . $model::$app_id . (!$is_special ? 1 : 0) . $type)->set($cacheKey, $all);
            }
        }
        return $result;
    }

    /**
     * 所有父级分类
     */
    public static function getAllParent($type, $is_special, $store = '', $filter_button = false)
    {
        $model = new static;
        $model = $model->with(['images'])
            ->where('parent_uuid', '=', 0)
            ->where('is_special', '=', $is_special)
            ->order(['create_time' => 'desc']);
        // 过滤按钮分类
        if ($filter_button) {
            $model = $model->where('uuid', '>', 0);
        }
        //
        $data = $model->select();
        $all = !empty($data) ? $data->toArray() : [];
        return $all;
    }

    /**
     * 所有分类 + 父级分类商品总数
     */
    public static function getProductCountByCategory($type, $shop_supplier_id, $is_special = 0, $is_sort = 1, $button_filter = false)
    {
        $child_order_conditions = $is_sort ? ['sort' => 'asc', 'create_time' => 'desc'] : ['create_time' => 'desc'];
        $model = new static;
        $supplier = SupplierModel::detail($shop_supplier_id);
        if ($supplier && $supplier['is_main'] == 0 && $supplier['category_set'] == 10) {
            $detail = SupplierModel::where('is_main', '=', 1)->find();
            $shop_supplier_id = $detail['shop_supplier_id'];
        }
        //
        $prefix = env('DB_PREFIX');
        $list = $model->alias('a')
            ->leftJoin("
                    (
                        SELECT if(c.parent_id, c.parent_id, c.category_id) as category_id, count(*) AS product_count
                        FROM {$prefix}product_package product
                        left join {$prefix}category c on `product`.`category_id` = c.category_id or c.category_id = product.special_id
                        where product.type = 10 and product.is_delete = 0 and product.product_status = 10 and c.status = 1 and c.is_button = 0
                        GROUP BY if(c.parent_id, c.parent_id, c.category_id)
                    ) product
                ", 'a.category_id = product.category_id or a.parent_id = product.category_id')
            ->with(['images', 'child' => function ($query) use ($child_order_conditions) {
                $query->field('category_id, is_special, name, parent_id')->where('status', '=', 1);
                $query->order($child_order_conditions);
            }])
            ->field('a.category_id, a.is_special, a.name, a.parent_id, COALESCE(product.product_count, 0) as product_count, a.is_button')
            ->where('a.parent_id', '=', 0)
            ->where('a.type', '=', $type)
            ->where('a.status', '=', 1)
            ->order(['a.is_special' => 'desc', 'a.sort' => 'asc', 'a.category_id' => 'asc'])
            ->where('a.shop_supplier_id', '=', $shop_supplier_id)
            ->select();
        //
        $list = $list->toArray();
        foreach ($list as &$item) {
            if ($item['category_id'] == 0) {
                unset($item['child']);
            }
        }
        return self::handleButtonList($list, $button_filter);
    }

    /**
     * 所有分类 + 父级分类商品总数
     */
    public static function getProductCountByCategoryOptimize($type, $shop_supplier_id, $is_special = 0, $is_sort = 1, $button_filter = false)
    {
        $child_order_conditions = $is_sort ? ['sort' => 'asc', 'create_time' => 'desc'] : ['create_time' => 'desc'];
        $model = new static;
        $supplier = SupplierModel::detail($shop_supplier_id);
        if ($supplier && $supplier['is_main'] == 0 && $supplier['category_set'] == 10) {
            $detail = SupplierModel::where('is_main', '=', 1)->find();
            $shop_supplier_id = $detail['shop_supplier_id'];
        }
        //
        $prefix = env('DB_PREFIX');
        $list = $model->alias('a')
            ->leftJoin("
                    (
                        SELECT if(c.parent_id, c.parent_id, c.category_id) as category_id, count(*) AS product_count
                        FROM {$prefix}product_package product
                        left join {$prefix}category c on `product`.`category_id` = c.category_id or c.category_id = product.special_id
                        where product.type = 10 and product.is_delete = 0 and product.product_status = 10 and c.status = 1 and c.is_button = 0
                        GROUP BY if(c.parent_id, c.parent_id, c.category_id)
                    ) product
                ", 'a.category_id = product.category_id or a.parent_id = product.category_id')
            ->with(['images', 'child' => function ($query) use ($child_order_conditions) {
                $query->field('category_id, is_special, name, parent_id')->where('status', '=', 1);
                $query->order($child_order_conditions);
            }])
            ->field('a.category_id, a.is_special, a.name, a.parent_id, COALESCE(product.product_count, 0) as product_count, a.is_button')
            ->where('a.parent_id', '=', 0)
            ->where('a.type', '=', $type)
            ->where('a.status', '=', 1)
            ->order(['a.is_special' => 'desc', 'a.sort' => 'asc', 'a.category_id' => 'asc'])
            ->where('a.shop_supplier_id', '=', $shop_supplier_id)
            ->select();
        //
        $list = $list->toArray();
        foreach ($list as &$item) {
            //
            $item['name'] = self::parseJsonValue($item['name']);
            if (!empty($item['child'])) {
                array_walk($item['child'], function (&$child) {
                    $child['name'] = self::parseJsonValue($child['name']);
                });
            }
            //
            if ($item['category_id'] == 0) {
                unset($item['child']);
                continue;
            }
        }
        return self::handleButtonList($list, $button_filter);
    }

    /**
     * 将JSON字符串转换为数组格式
     * @param string $value 待转换的JSON字符串
     * @return array|string 转换后的数组或原始字符串
     */
    public static function parseJsonValue($value)
    {
        if (empty($value)) {
            return [];
        }
        $decoded = json_decode($value, true);
        if (json_last_error() === JSON_ERROR_NONE && is_array($decoded)) {
            return $decoded;
        }
        return $value;
    }

    /**
     * 所有一级分类及分类在订单已添加商品数量
     */
    public static function getOrderProductCountByCategory($type, $order_id, $product_source = Order::CASHIER_PRODUCT_SOURCE)
    {
        //
        $prefix = env('DB_PREFIX');
        return (new self)->alias('c')
            ->leftJoin("
                    (
                        SELECT if(cc.parent_id, cc.parent_id, cc.category_id) as category_id, count(*) AS product_num
                        FROM {$prefix}product_package product
                        left join {$prefix}product_category cc on `product`.`category_id` = cc.category_id or cc.category_id = product.special_id and cc.is_button = 0
                        where product.type = 10 and product.is_delete = 0 and product.product_status = 10 and cc.status = 1
                        GROUP BY if(cc.parent_id, cc.parent_id, cc.category_id)
                    ) product_num
                ", 'c.category_id = product_num.category_id or c.parent_id = product_num.category_id')
            ->leftJoin("(
                SELECT
                    if(c.parent_id, c.parent_id, c.category_id) as category_id,
                    SUM(op.total_num) AS product_count
                FROM {$prefix}order_product op
                INNER JOIN {$prefix}product_package p ON op.product_id = p.product_id
                LEFT JOIN {$prefix}product_category c ON p.category_id = c.category_id OR c.category_id = p.special_id and c.is_button = 0
                WHERE op.order_id = {$order_id}
                    AND op.is_send_kitchen = 0
                    AND op.add_source = {$product_source}
                    AND p.type = 10
                    AND p.is_delete = 0
                    AND p.product_status = 10
                    AND c.status = 1
                GROUP BY if(c.parent_id, c.parent_id, c.category_id)
            ) product", 'c.category_id = product.category_id OR c.parent_id = product.category_id')
            ->field('c.category_id, c.is_special, c.name, c.parent_id, IF(c.is_special = 0, COALESCE(product.product_count, 0), 0) as product_count, COALESCE(product_num.product_num, 0) AS product_num')
            ->where('c.parent_id', '=', 0)
            ->where('c.category_id', '>', 0)
            ->where('c.type', '=', $type)
            ->where('c.status', '=', 1)
            ->where('product_num.product_num', '>', 0)  // 分类下有商品才显示
            ->order(['c.is_special' => 'desc', 'c.sort' => 'asc', 'c.category_id' => 'asc'])
            ->select();
    }

    /**
     * 获取所有分类
     */
    public static function getCacheAll($type, $is_special, $store = '', $name = '', $is_sort = true)
    {
        return self::getAll($type, $is_special, $store, $name, $is_sort);
    }

    /**
     * 获取所有分类(树状结构)
     */
    public static function getCacheTree($type, $is_special, $store = [], $name = '', $button_filter = true)
    {
        request()->is_paginate = false;
        $list = self::getAll($type, $is_special, $store, $name);
        return self::handleButtonList($list, $button_filter);
    }

    /**
     * 获取指定分类下的所有子分类id
     */
    public static function getSubCategoryId($parent_id, $all = [])
    {
        $arrIds = [$parent_id];
        empty($all) && $all = self::getCacheAll(1, 0);
        foreach ($all['data'] as $key => $item) {
            if ($item['parent_id'] == $parent_id) {
                unset($all[$key]);
                $subIds = self::getSubCategoryId($item['category_id'], $all);
                !empty($subIds) && $arrIds = array_merge($arrIds, $subIds);
            }
        }
        return $arrIds;
    }

    /**
     * 指定的分类下是否存在子分类
     */
    protected static function hasSubCategory($parentId)
    {
        $all = self::getCacheAll(1, 0);
        foreach ($all as $item) {
            if ($item['parent_id'] == $parentId) {
                return true;
            }
        }
        return false;
    }


    /**
     * 关联图片
     */
    public function image()
    {
        return $this->belongsTo('app\common\model\file\UploadFile', 'image_id', 'file_id');
    }

    /**
     * 获取所有一级分类
     */
    public static function getFirstCategory()
    {
        return self::where('parent_id', '=', 0)
            ->order(['create_time' => 'desc'])
            ->select();
    }

    //新增特殊分类
    public function addSpecial($app_id, $shop_supplier_id)
    {
        $data = [
            ['name' => '新品', 'is_special' => 1, 'type' => 0],
            ['name' => '热卖', 'is_special' => 1, 'type' => 0],
            ['name' => '套餐', 'is_special' => 1, 'type' => 0],
            ['name' => '新品', 'is_special' => 1, 'type' => 1],
            ['name' => '热卖', 'is_special' => 1, 'type' => 1],
            ['name' => '套餐', 'is_special' => 1, 'type' => 1],
        ];
        $this->saveAll($data);
    }

    //获取所有分类
    public function getAllCategory($type, $shop_supplier_id, $isSpecial = '', $parentId = '', $button_filter = false)
    {
        $supplier = SupplierModel::detail($shop_supplier_id);
        if ($supplier['is_main'] == 0 && $supplier['category_set'] == 10) {
            $detail = SupplierModel::where('is_main', '=', 1)->find();
            $shop_supplier_id = $detail['shop_supplier_id'];
        }
        $list = $this->where('type', '=', $type)
            ->with(['child'])
            ->when($isSpecial !== '', function ($q) use ($isSpecial) {
                $q->where('is_special', '=', $isSpecial);
            })
            ->when($parentId !== '', function ($q) use ($parentId) {
                $q->where('parent_id', '=', $parentId);
            })
            ->order('is_special desc, sort asc')
            ->select();
        return self::handleButtonList($list->toArray(), $button_filter);
    }

    /**
     * 获取活跃的所有分类
     */
    public static function getActiveALL($type, $is_special, $store = '', $name = '', $button_filter = false)
    {
        return self::handleButtonList(self::getAll($type, $is_special, $store, $name, 0), $button_filter);;
    }

    /**
     * 检查是否被关联
     */
    public function isUseWithProduct($category_id)
    {
        return $this->getProductNum($category_id) > 0;
    }

    /**
     * 获取分类关联的商品数量
     *
     * @param int $category_id
     * @return int
     */
    /**
     * 获取分类关联的商品数量（成品 + 材料）
     *
     * @param int $category_id
     * @return int
     */
    public function getProductNum($category_id): mixed
    {
        $model = new self;
        $prefix = env('DB_PREFIX');

        $level = $model->where('uuid', $category_id)->value('parent_uuid') != 0;
        if ($level) {
            return Product::where('uuid', $category_id)->count();
        } else {
            return $model->alias('a')
                ->leftJoin("
                    (
                        SELECT IF(c.parent_uuid, c.parent_uuid, c.uuid) as category_uuid, COUNT(*) AS product_count
                        FROM {$prefix}product_package product
                        LEFT JOIN {$prefix}product_category c ON product.category_uuid = c.uuid OR c.uuid = product.special_category_uuid
                        WHERE c.uuid > 0
                        GROUP BY IF(c.parent_uuid, c.parent_uuid, c.uuid)
                    ) product
                ", 'a.uuid = product.category_uuid OR a.parent_uuid = product.category_uuid')
                ->field('a.uuid, a.is_special, a.name, a.parent_uuid, COALESCE(product.product_count, 0) as product_count')
                ->where('a.parent_uuid', 0)
                ->where('a.uuid', $category_id)
                ->value('product_count') ?: 0;
        }
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
     * 是否显示"全部"按钮分类
     */
    public static function handleButtonList($list, $button_filter = false)
    {
        $version = request()->header('Version-Name') ?: '0.0.0';
        if (($version && version_compare($version, '1.0.8', '<')) || $button_filter) {
            $list = array_filter($list, function ($category) {
                $is_button = $category['is_button'] ?? 0;
                return $is_button == 0;
            });
        }
        return array_values($list);
    }
}
