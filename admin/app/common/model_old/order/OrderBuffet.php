<?php

namespace app\common\model_old\order;

use app\common\library\helper;
use app\common\model_old\BaseModel;
use app\common\enum\order\OrderStatusEnum;
use app\common\model_old\order\Order as OrderModel;

/**
 * 订单自助餐模型
 */
class OrderBuffet extends BaseModel
{
    protected $name = 'order_buffet';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'name_text',
        'total_product_price',
    ];

    public function getNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['name']);
    }

    /**
     * 兼容商品列表时的价格字段
     */
    public function getTotalProductPriceAttr($value, $data = [])
    {
        return $data['total_price'];
    }

    /**
     * 关联自助餐表
     */
    public function buffet()
    {
        return $this->belongsTo('app\\common\\model\\buffet\\Buffet', 'buffet_id', 'id');
    }

    /**
     * 关联自助餐优惠订单数据
     */
    public function delOrderBuffetDiscount()
    {
        return $this->hasMany('app\\common\\model\\order\\OrderBuffetDiscount', 'order_id', 'order_id')->where('buffet_id', '=', $this['buffet_id']);
    }

    /**
     * 订单商品列表
     */
    public function buffetProduct()
    {
        return $this->hasMany('app\\common\\model\\buffet\\BuffetProduct', 'buffet_id', 'buffet_id');
    }

    /**
     * 订单自助餐顾客类型价格列表
     */
    public function buffetCustomerType()
    {
        return $this->hasMany('app\\common\\model\\order\\OrderBuffetCustomer', 'order_id', 'order_id');
    }

    // 删除订单自助餐
    public function del($order_buffet_id)
    {
        $this->startTrans();
        try {
            $model = $this->where('id', '=', $order_buffet_id)->find();
            if (!$model) {
                $this->error = '记录不存在';
                return false;
            }
            // 检查订单状态
            $order = OrderModel::detail([
                ['order_id', '=', $model['order_id']],
                ['order_status', '=', OrderStatusEnum::NORMAL]
            ]);
            if (!$order) {
                $this->rollback();
                $this->error = '当前订单不可修改';
                return false;
            }
            if ($order['is_lock'] == 1) {
                $this->rollback();
                $this->error = '订单已被锁定，请解锁后重新操作';
                return false;
            }
            $orderId = $order['order_id'];
            $model->force()->delete();
            $model->delOrderBuffetDiscount()->delete();
            // 就餐剩余时间变化
            $time_limit = (new OrderBuffet())->where('order_id', '=', $orderId)->min('time_limit');
            //
            if (!(new OrderBuffet())->where('order_id', '=', $orderId)->find()) {
                $order->save(['buffet_expired_time' => 0]);
            } else if ($time_limit > 0) {
                $time_limit = floatval($time_limit);
                $buffet_expired_time = helper::bcadd(strtotime($order['create_time']), helper::bcmul($time_limit, 60));
                $order->save(['buffet_expired_time' => $buffet_expired_time]);
            }
            (new OrderModel())->reloadPrice($orderId);
            $this->commit();
            return $orderId;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 转移自助餐到子订单
     * @param $mainOrderId
     * @param $fromOrderId
     * @param $toOrderId
     * @param $price
     * @param $num
     * @return bool
     */
    public static function addToSubOrder($mainOrderId, $fromOrderId, $toOrderId, $price, $num)
    {
        // 获取自助餐顾客
        $fromOrderBuffet = self::where([
            'order_id' => $mainOrderId,
            'sub_order_id' => $fromOrderId,
        ])->find();

        if ($fromOrderBuffet->num == $num) {
            // 转移全部数量
            $fromOrderBuffet->sub_order_id = $toOrderId;
            self::addOrIncNum($fromOrderBuffet, $price, $num);
            $fromOrderBuffet->force()->delete();
        } else {
            // 转移部分数量
            // 从原子单中扣除部分数量
            $fromTotalNum = helper::bcsub((string)$fromOrderBuffet->num, (string)$num);
            $fromTotalPrice = helper::bcmul((string)$price, $fromTotalNum);
            $fromOrderBuffet->num = (int)$fromTotalNum;
            $fromOrderBuffet->total_price = (float)$fromTotalPrice;
            $fromOrderBuffet->save();
            // 添加到目标子单
            $fromOrderBuffet->sub_order_id = $toOrderId;
            self::addOrIncNum($fromOrderBuffet, $price, $num);
        }
    }

    /**
     * 添加或增加自助餐数量
     * @param $orderBuffet
     * @param $price
     * @param $num
     * @return bool
     */
    public static function addOrIncNum($orderBuffet, $price, $num)
    {
        $existsOrderBuffet = self::where([
            'order_id' => $orderBuffet->order_id,
            'sub_order_id' => $orderBuffet->sub_order_id,
            'buffet_id' => $orderBuffet->buffet_id,
            'buy_limit_status' => $orderBuffet->buy_limit_status,
            'is_comb' => $orderBuffet->is_comb,
        ])->find();
        if ($existsOrderBuffet) {
            $totalNum = helper::bcadd((string)$existsOrderBuffet->num, (string)$num);
            $totalPrice = helper::bcmul((string)$price, $totalNum);
            $existsOrderBuffet->save([
                'num' => (int)$totalNum,
                'total_price' => (float)$totalPrice,
            ]);
        } else {
            $totalPrice = (float)helper::bcmul((string)$price, (string)$num);
            $newOrderBuffet = new self();
            $newOrderBuffet->create([
                'order_id' => $orderBuffet->order_id,
                'sub_order_id' => $orderBuffet->sub_order_id,
                'buffet_id' => $orderBuffet->buffet_id,
                'name' => $orderBuffet->getData('name'),
                'price' => $orderBuffet->price,
                'num' => (int)$num,
                'total_price' => (float)$totalPrice,
                'buy_limit_status' => $orderBuffet->buy_limit_status,
                'is_comb' => $orderBuffet->is_comb,
                'time_limit' => $orderBuffet->time_limit,
                'app_id' => $orderBuffet->app_id,
            ]);
        }
    }
}
