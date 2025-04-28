<?php

namespace app\cashier\model\order;

use app\common\enum\order\OrderStatusEnum;
use app\common\model\order\Order as OrderModel;

/**
 * 收银订单模型
 */
class OrderCashier extends OrderModel
{
    protected $append = [];

    /**
     * 隐藏字段
     * @var array
     */
    protected $hidden = [];

    public function getPayStatusAttr($value)
    {
        return $value;
    }

    public function getOrderStatusAttr($value)
    {
        return $value;
    }

    public function getUpdatePriceAttr($value)
    {
        return $value;
    }

    public function getDeliveryTypeAttr($value)
    {
        return $value;
    }

    public function getDeliveryStatusAttr($value)
    {
        return $value;
    }

    public function getReceiptStatusAttr($value)
    {
        return $value;
    }

    /**
     * 获取收银进行中订单实例
     * @param $where
     * @param $field
     * @return array|false|mixed
     */
    public function underwayCashierDetail($where, $field = ['*'])
    {
        is_array($where) ? $filter = $where : $filter['order_id'] = (int) $where;
        $order = self::where($filter)->where('order_status', OrderStatusEnum::NORMAL)->order('order_id', 'desc')->field($field)->hidden(['create_time', 'update_time'])->find();
        if (!$order) {
            $this->error = '订单不存在';
            return false;
        }
        if ($order->is_lock == 1) {
            $this->error = '订单已被锁定，请解锁后重新操作';
            return false;
        }
        return $order;
    }
}
