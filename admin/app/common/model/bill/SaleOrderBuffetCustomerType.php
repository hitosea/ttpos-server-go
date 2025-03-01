<?php

namespace app\common\model\bill;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 * 销售单自助餐顾客类型模型
 */
class SaleOrderBuffetCustomerType extends BaseModel
{
    use SoftDelete;

    protected $name = 'sale_order_buffet_customer_type';
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 关联销售单
     */
    public function saleOrder()
    {
        return $this->belongsTo(SaleOrder::class, 'sale_order_uuid', 'uuid');
    }
}
