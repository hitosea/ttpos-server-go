<?php

namespace app\common\model\bill;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

class SaleOrderConpon extends BaseModel {
    use SoftDelete;

    protected $name = 'sale_order_coupon';
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 关联销售订单
     */
    public function saleOrder()
    {
        return $this->belongsTo(SaleOrder::class, 'sale_order_uuid', 'uuid');
    }
}