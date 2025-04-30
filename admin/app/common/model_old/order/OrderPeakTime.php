<?php

namespace app\common\model_old\order;

use app\common\model_old\BaseModel;

/**
 * 订单自助餐模型
 */
class OrderPeakTime extends BaseModel
{
    protected $name = 'order_peak_time';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [];

    /**
     * 时间段
     */
    public function getTimePeriodAttr($value, $data)
    {
        $startHour = str_pad($data['hour'], 2, '0', STR_PAD_LEFT);      // 将整点数补齐为两位数作为起始小时
        $endHour = str_pad($data['hour'] + 1, 2, '0', STR_PAD_LEFT);    // 结束小时为起始小时加1
        $time = $startHour . ':00-' . $endHour . ':00';
        return date('m/d', $data['date']) . ' ' . $time;
    }

    /**
     * 高峰时间段记录
     * @param string $type inc-加峰值 desc-减峰值
     * @param int $orderId 订单id
     * @return bool
     */
    public function record($type, $orderId, $refundMoney = 0)
    {
        // 类型判断
        if (!in_array($type, ['inc', 'desc'])) {
            return false;
        }
        //
        $order = Order::where('order_id', $orderId)->field('pay_time, pay_price, refund_money, cashier_id, is_free, shop_supplier_id, app_id')->find();
        if (!$order) {
            return false;
        }
        $time = $order['pay_time'];
        $isFree = $order['is_free'] ?? 0;
        $amount = $isFree ? 0 : max(0, $order['pay_price'] - $order['refund_money']);
        $cashierId = $order['cashier_id'] ?? 0;
        $shopSupplierId = $order['shop_supplier_id'];
        $appId = $order['app_id'];

        // 日期-天 和 时间-小时
        $date = strtotime(date('Y-m-d', $time));
        $hour = date('H', $time);

        // 查询条件
        $condition = [
            'date' => $date,
            'hour' => $hour,
            'cashier_id' => $cashierId,
            'shop_supplier_id' => $shopSupplierId,
            'app_id' => $appId
        ];

        // 现有记录
        $existingRecord = $this->where($condition)->find();
        if ($existingRecord) {
            // 产品需求，当遇到订单退款时，订单数不改变，只改变金额（v1.1.1）
            $descNum = $refundMoney > 0 ? $existingRecord['num'] : max(0, $existingRecord['num'] - 1);
            $descAmount = $refundMoney > 0 ? max(0, $existingRecord['amount'] - $refundMoney) : max(0, $existingRecord['amount'] - $amount);
            $updateData = [
                'num' => $type === 'inc' ? $existingRecord['num'] + 1 : $descNum,
                'amount' => $type === 'inc' ? $existingRecord['amount'] + $amount : $descAmount,
            ];
            $this->where($condition)->update($updateData);
        } else if ($type === 'inc') {
            $insertData = array_merge($condition, [
                'num' => 1,
                'amount' => $amount,
            ]);
            $this->save($insertData);
        }

        return false;
    }

    /**
     * 获取最高峰时间段记录
     * @param int $startTime 开始时间
     * @param int $endTime 结束时间
     * @param int $shopCashierId 收银员id，默认0
     * @return array
     */
    public static function getMaxRecord($startTime, $endTime, $shopCashierId = 0)
    {
        $startTimeH = strtotime(date('Y-m-d H', $startTime) . ':00:00');
        $endTimeH = strtotime(date('Y-m-d H', $endTime) . ':00:00');
        //
        $maxIds = OrderPeakTime::whereRaw("(date + hour * 60 * 60) between $startTimeH and $endTimeH")
            ->when($shopCashierId, function ($q) use ($shopCashierId) {
                $q->where('cashier_id', '=', $shopCashierId);
            })
            ->field('sum(num) as sum_num, group_concat(id) as ids')
            ->order('sum_num', 'desc')
            ->group("CONCAT(date,hour)")
            ->find();
        //
        return $maxIds ? OrderPeakTime::where('id', 'in', $maxIds->ids)
            ->field('date, hour, sum(num) as num, sum(amount) amount')
            ->group("CONCAT(date,hour)")
            ->select()
            ->append(['time_period']) : [];
    }
}
