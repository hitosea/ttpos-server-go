<?php

namespace app\shop\validate;

use app\common\model\product\Product;
use app\common\validate\BaseValidate;

/**
 * 商品导入验证类
 */
class ProductPackageRecommendValidate extends  BaseValidate
{
    //定义验证规则
    protected $rule = [
        'title' => 'require|max:30',
        'status' => 'require|in:0,1',
        'product_packages' => 'require|array|min:3|max:15|checkProductPackagesExists|checkProductPackagesSortUnique',
        'product_packages.*.uuid' => 'require|integer',
        'product_packages.*.sort' => 'require|integer',
    ];

    protected $message = [
        'title.require' => '推荐标题不能为空',
        'title.max' => '推荐标题最多30个字符',
        'status.require' => '状态不能为空',
        'status.in' => '状态只能为0或1',
        'product_packages.require' => '商品不能为空',
        'product_packages.array' => '商品必须为数组',
        'product_packages.min' => '商品至少3个',
        'product_packages.max' => '商品最多15个',
        'product_packages.checkProductPackagesSortUnique' => '商品排序不能重复',
        'product_packages.checkProductPackagesExists' => '商品不存在',
        'product_packages.*.uuid.require' => '商品uuid不能为空',
        'product_packages.*.uuid.integer' => '商品uuid必须为整数',
        'product_packages.*.sort.require' => '商品排序不能为空',
        'product_packages.*.sort.integer' => '商品排序必须为整数',
    ];

    protected $scene = [
        'save' => [
            'title',
            'status',
            'product_packages',
        ],
    ];

    /**
     * 检查商品排序是否唯一
     */
    public function checkProductPackagesSortUnique($value, $rule, $data)
    {
        if (!is_array($value)) {
            return false;
        }
        $sorts = array_column($value, 'sort');

        return count($sorts) === count(array_unique($sorts));
    }

    /**
     * 检查商品是否存在
     */
    public function checkProductPackagesExists($value, $rule, $data)
    {
        if (!is_array($value)) {
            return false;
        } 
        $count = Product::where('uuid', 'in', array_column($value, 'uuid'))->where('delete_time', 0)->count();
        return count($value) == $count;
    }
}
