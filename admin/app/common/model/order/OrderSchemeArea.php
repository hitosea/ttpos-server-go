<?php

namespace app\common\model\order;

use think\facade\Db;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\product\Product;
use app\common\model\store\TableArea;

/**
 * 订单方案模型
 */
class OrderSchemeArea extends BaseModel
{
    use SoftDelete;
    protected $pk = 'id';
    protected $name = 'product_must_plan_region';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
}
