<?php

namespace app\common\model_old\order;


use app\common\model_old\BaseModel;

/**
 * 订单结算记录
 */
class OrderSettled extends BaseModel
{
    protected $name = 'order_settled';
    protected $pk = 'settled_id';

    /**
     * 关联订单主表
     */
    public function orderMaster()
    {
        return $this->belongsTo('app\\common\\model_old\\order\\Order');
    }

    /**
     * 关联供应商表
     */
    public function supplier()
    {
        return $this->belongsTo('app\\common\\model_old\\supplier\\Supplier', 'shop_supplier_id', 'shop_supplier_id')->field(['shop_supplier_id', 'name']);
    }
    /**
     * 详情
     */
    public static function detail($settled_id)
    {
        return self::with(['orderMaster'])->find($settled_id);
    }
}
