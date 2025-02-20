<?php

namespace app\common\model\product;

use think\facade\Db;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 * 产品属性组
 */
class ProductAttributeGroup extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_package_attribute_group';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $append = [];

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->belongsTo('app\common\model\product\Product', 'product_package_uuid', 'uuid');
    }

    /**
     * 关联属性
     */
    public function attribute()
    {
        return $this->belongsTo('app\common\model\product\AttributeGroup', 'product_attribute_group_uuid', 'uuid');
    }
}
