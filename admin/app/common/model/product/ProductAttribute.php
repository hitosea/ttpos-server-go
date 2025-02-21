<?php

namespace app\common\model\product;

use think\facade\Db;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 * 产品属性
 */
class ProductAttribute extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_package_attribute';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $append = [];

    /**
     * 关联属性
     */
    public function attribute()
    {
        return $this->belongsTo('app\common\model\product\Attribute', 'attribute_uuid', 'uuid');
    }

    /**
     * 关联属性组
     */
    public function productAttributeGroup()
    {
        return $this->belongsTo('app\common\model\product\ProductAttributeGroup', 'product_package_attribute_group_uuid', 'uuid');
    }
}
