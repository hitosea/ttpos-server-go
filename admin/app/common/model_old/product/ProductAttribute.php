<?php

namespace app\common\model_old\product;

use think\facade\Db;
use app\common\model_old\BaseModel;

/**
 * 产品属性
 */
class ProductAttribute extends BaseModel
{
    protected $name = 'product_attribute';
    protected $pk = 'attribute_id';
    protected $append = [];

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->belongsTo('app\common\model_old\product\Product', 'product_id', 'product_id');
    }

    /**
     * 关联属性
     */
    public function attribute()
    {
        return $this->belongsTo('app\common\model_old\product\Attribute', 'attribute_id', 'attribute_id');
    }

    /**
     * 关联属性组
     */
    public function productAttributeGroup()
    {
        return $this->belongsTo('app\common\model_old\product\ProductAttributeGroup', 'group_attribute_id', 'group_attribute_id');
    }
}
