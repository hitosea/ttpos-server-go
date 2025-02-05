<?php

namespace app\common\model\order;

use app\common\model\BaseModel;
use app\common\exception\BaseException;
use app\common\enum\order\OrderStatusEnum;
use app\common\library\helper;
use app\common\model\order\Order as OrderModel;

/**
 * 订单自助餐模型
 */
class OrderDelay extends BaseModel
{
    protected $name = 'order_delay';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'name_text',
        'total_product_price',
    ];

    /**
     * 获取名称
     */
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

    // 删除订单加钟
    public function del($order_delay_id)
    {
        $this->startTrans();
        try {
            $model = $this->where('id', '=', $order_delay_id)->find();
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
            $del_model_delay_time = $model['delay_time'];
            $expired_time = $model['expired_time'];
            $model->force()->delete();
            // 订单剩余时间扣除
            if ($expired_time == 0 && $order['buffet_expired_time'] != 1) {
                $buffet_expired_time = $order['buffet_expired_time'] - $del_model_delay_time * 60;
                $order->save(['buffet_expired_time' => $buffet_expired_time]);
            }
            (new OrderModel())->reloadPrice($order['order_id']);
            $this->commit();
            return $orderId;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 转移加钟到子单
     * @param $id 加钟ID
     * @param $fromOrderId 原订单ID
     * @param $toOrderId 目标订单ID
     * @param $num 数量
     * @return bool
     */
    public static function addToSubOrder($id, $fromOrderId, $toOrderId, $num)
    {
        $fromOrderDelay = self::where([
            'id' => $id,
            'sub_order_id' => $fromOrderId,
        ])->find();
        if ($fromOrderDelay->num == $num) {
            // 转移全部数量
            $fromOrderDelay->sub_order_id = $toOrderId;
            self::addOrIncNum($fromOrderDelay, $num);
            $fromOrderDelay->force()->delete();
        } else {
            // 转移部分数量
            $fromTotalNum = helper::bcsub((string)$fromOrderDelay->num, (string)$num);
            $fromTotalPrice = helper::bcmul((string)$fromOrderDelay->price, $fromTotalNum);
            $fromOrderDelay->num = (int)$fromTotalNum;
            $fromOrderDelay->total_price = (float)$fromTotalPrice;
            $fromOrderDelay->save();

            // 添加到目标子单
            $fromOrderDelay->sub_order_id = $toOrderId;
            self::addOrIncNum($fromOrderDelay, $num);
        }
    }

    /**
     * 添加或增加加钟数量
     * @param $orderDelay
     * @param $num
     * @return bool
     */
    public static function addOrIncNum($orderDelay, $num)
    {
        $existsOrderDelay = self::where([
            'order_id' => $orderDelay->order_id,
            'sub_order_id' => $orderDelay->sub_order_id,
            'delay_id' => $orderDelay->delay_id,
        ])->find();

        if ($existsOrderDelay) {
            $totalNum = helper::bcadd((string)$existsOrderDelay->num, (string)$num);
            $totalPrice = helper::bcmul((string)$existsOrderDelay->price, $totalNum);
            $existsOrderDelay->save([
                'num' => (int)$totalNum,
                'total_price' => (float)$totalPrice,
            ]);
        } else {
            unset($orderDelay->id);
            unset($orderDelay->create_time);
            unset($orderDelay->update_time);
            $totalPrice = helper::bcmul((string)$orderDelay->price, (string)$num);
            $orderDelay->num = (int)$num;
            $orderDelay->total_price = (float)$totalPrice;
            $newOrderDelay = new self();
            $newOrderDelay->save($orderDelay);
        }
    }
}
