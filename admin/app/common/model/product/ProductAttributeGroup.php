<?php

namespace app\common\model\product;

use think\facade\Db;
use app\common\model\BaseModel;

/**
 * 产品属性组
 */
class ProductAttributeGroup extends BaseModel
{
    protected $name = 'product_attribute_group';
    protected $pk = 'group_attribute_id';
    protected $append = [];

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->belongsTo('app\common\model\product\Product', 'product_id', 'product_id');
    }

    /**
     * 关联属性
     */
    public function attribute()
    {
        return $this->belongsTo('app\common\model\product\Attribute', 'attribute_id', 'attribute_id');
    }
}
