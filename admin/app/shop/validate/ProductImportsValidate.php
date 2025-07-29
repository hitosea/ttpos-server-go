<?php

namespace app\shop\validate;

use think\facade\Db;
use app\common\model\product\Spec;
use app\common\model\product\Unit;
use app\common\model\tax\TaxCategory;
use app\common\validate\BaseValidate;
use app\common\model\product\Category;
use app\common\enum\settings\LanguageEnum;

/**
 * 商品导入验证类
 */
class ProductImportsValidate extends  BaseValidate
{
    //定义验证规则
    protected $rule = [
        'product_name' => 'require',
        'category_name' => 'require',
        'deduct_stock_type' => 'require|in:1,2,10,20',
        'num_type' => 'require|in:1,2',
        'product_unit' => 'require',
        'spec_name' => 'require',
        'img_name' => 'string',
        'product_stock' => 'require|integer|between:0,99999999',
        'barcode' => 'string',
        'product_price' => 'float',
        'product_status' => 'require|in:1,0,10,20',
        'product_ratin_tax_type' => 'require',
        'product_takeout_tax_type' => 'require',
        'shows' => 'require|integer',
        'product_sort' => 'integer',
        'limit_num' => 'integer',
        'is_enable_grade' => 'require|in:0,1',
        'category_id' => 'require',
        'unit_id' => 'require',
        'spec_id' => 'require',
        'ratin_tax_id' => 'require',
        'takeout_tax_id' => 'require',
        'open_overall_discount' => 'require|in:0,1',
    ];

    protected $message = [
        'product_name.require' => '商品名称不能为空',
        'category_name.require' => '所属分类不能为空',
        'category_id.require' => '所属分类不能为空',
        'category_name.checkCategoryNameExist' => '所属分类不存在',
        'deduct_stock_type.require' => '库存计算方式不能为空',
        'deduct_stock_type.in' => '库存计算方式必须是1和2',
        'num_type.require' => '计价方式必须是1和2',
        'num_type.in' => '计价方式必须是1和2',
        'product_unit.require' => '商品单位不能为空',
        'unit_id.require' => '商品单位不能为空',
        'spec_id.require' => '规格不能为空',
        'spec_name.require' => '规格名称不能为空',
        'img_name.require' => '图片名称不能为空',
        'product_stock.require' => '库存数量不能为空',
        'product_stock.integer' => '库存数量必须是0到99999999之间的整数',
        'product_stock.between' => '库存数量必须是0到99999999之间的整数',
        'barcode.require' => '商品条码不能为空',
        // 'product_price.require' => '商品价格不能为空', // 价格可以为0，先注释掉
        'product_price.float' => '商品价格必须是浮点数',
        'product_status.require' => '商品状态必须是0和1',
        'product_status.in' => '商品状态必须是0和1',
        'product_ratin_tax_type.require' => '堂食税类不能为空',
        'ratin_tax_id.require' => '堂食税类不能为空',
        'product_takeout_tax_type.require' => '外带税类不能为空',
        'takeout_tax_id.require' => '外带税类不能为空',
        'shows.require' => '显示不能为空',
        'shows.integer' => '显示必须是12345',
        'product_sort.require' => '商品排序不能为空',
        'limit_num.require' => '限购数量不能为空',
        'limit_num.integer' => '限购数量必须是整数',
        'is_enable_grade.require' => '是否开启会员折扣必须是0和1',
        'is_enable_grade.in' => '是否开启会员折扣必须是0和1',
        'open_overall_discount.require' => '是否开启整单折扣必须是0和1',
        'open_overall_discount.in' => '是否开启整单折扣必须是0和1',
    ];

    protected $scene = [
        'get' => [
            'product_name',
            'category_name',
            'deduct_stock_type',
            'num_type',
            'product_unit',
            'spec_name',
            'img_name',
            'product_stock',
            'barcode',
            'product_price',
            'product_status',
            'product_ratin_tax_type',
            'product_takeout_tax_type',
            'shows',
            'product_sort',
            'limit_num',
            'is_enable_grade',
            'open_overall_discount',
        ],
        'save' => [
            'product_name',
            'deduct_stock_type',
            'num_type',
            'img_name',
            'product_stock',
            'barcode',
            'product_price',
            'product_status',
            'ratin_tax_id',
            'takeout_tax_id',
            'product_sort',
            'limit_num',
            'is_enable_grade',
            'category_id',
            'unit_id',
            'spec_id',
            'ratin_tax_id',
            'takeout_tax_id',
        ],
    ];

    /**
     * 验证分类是否存在
     */
    public function checkCategoryNameExist($value, $rule = [], $data = [])
    {
        $values = explode('/', $value);
        $category_id = '';
        //
        $res = Category::where('parent_uuid', 0)->where(function ($q) use ($values) {
            foreach (LanguageEnum::data() as $lang) {
                $key = $lang['name'];
                $q->whereOr([[Db::raw("JSON_UNQUOTE(JSON_EXTRACT(name, '$.$key'))"), '=', "{$values[0]}"]]);
            }
        })->find();
        if (!$res) {
            // $msg = __("所属分类不存在") . ' [' . $values[0] . ']';
            // if ($rule) {
            //     throw new \think\exception\HttpException(0, $msg);
            // }
            return $category_id;
        }
        $category_id = $res->uuid;
        //
        if ($values[1] ?? '') {
            $res = Category::where('parent_uuid', $res->uuid)->where(function ($q) use ($values) {
                foreach (LanguageEnum::data() as $lang) {
                    $key = $lang['name'];
                    $q->whereOr([[Db::raw("JSON_UNQUOTE(JSON_EXTRACT(name, '$.$key'))"), '=', "{$values[1]}"]]);
                }
            })->find();
            if (!$res) {
                // $msg = __("所属分类不存在") . ' [' . $values[1] . ']';
                // if ($rule) {
                //     throw new \think\exception\HttpException(0, $msg);
                // }
                return $category_id;
            }
            $category_id = $res->uuid;
        }
        //
        if ($rule) {
            return $category_id;
        }
        //
        return true;
    }

    /**
     * 验证商品单位是否存在
     */
    public function checkProductUnitExist($value, $rule = [], $data = [])
    {
        $res = Unit::where(function ($q) use ($value) {
            foreach (LanguageEnum::data() as $lang) {
                $key = $lang['name'];
                $q->whereOr([[Db::raw("JSON_UNQUOTE(JSON_EXTRACT(name, '$.$key'))"), '=', "{$value}"]]);
            }
        })->find();
        if (!$res) {
            // $msg = __("商品单位不存在") . ' [' . $value . ']';
            // if ($rule) {
            //     throw new \think\exception\HttpException(0, $msg);
            // }
            // return $msg;
            return '';
        }
        //
        if ($rule) {
            return $res->uuid;
        }
        //
        return true;
    }

    /**
     * 验证规格名称是否存在
     */
    public function checkSpecNameExist($value, $rule = [], $data = [])
    {
        $res = Spec::where(function ($q) use ($value) {
            foreach (LanguageEnum::data() as $lang) {
                $key = $lang['name'];
                $q->whereOr([[Db::raw("JSON_UNQUOTE(JSON_EXTRACT(name, '$.$key'))"), '=', "{$value}"]]);
            }
        })->find();
        if (!$res) {
            // $msg = __("规格不存在") . ' [' . $value . ']';
            // if ($rule) {
            //     throw new \think\exception\HttpException(0, $msg);
            // }
            // return $msg;
            return '';
        }
        //
        if ($rule) {
            return $res->uuid;
        }
        //
        return true;
    }

    /**
     * 验证tax是否存在
     */
    public function checkTaxExist($value, $rule = [], $data = [])
    {
        $res = TaxCategory::where('name', $value)->find();
        if (!$res) {
            // $msg = __("规格不存在") . ' [' . $value . ']';
            // if ($rule) {
            //     throw new \think\exception\HttpException(0, $msg);
            // }
            // return $msg;
            return '';
        }
        //
        if ($rule) {
            return $res->uuid;
        }
        //
        return true;
    }
}
