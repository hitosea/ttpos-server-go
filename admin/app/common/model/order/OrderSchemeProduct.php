<?php

namespace app\common\model\order;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 * 订单方案模型
 */
class OrderSchemeProduct extends BaseModel
{
    use SoftDelete;
    protected $pk = 'id';
    protected $name = 'product_must_plan_item';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    protected $append = ['product_id'];

    public function getProductIdAttr($value, $data)
    {
        return $data['product_package_uuid'] ?: 0;
    }
}
