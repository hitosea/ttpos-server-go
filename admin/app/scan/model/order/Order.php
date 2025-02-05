<?php

namespace app\scan\model\order;

use think\facade\Log;
use think\facade\Cache;
use app\common\library\helper;
use app\common\model\order\OrderDelay;
use app\common\model\order\OrderBuffet;
use app\common\model\supplier\Supplier;
use app\common\enum\order\OrderTypeEnum;
use app\common\model\order\OrderProduct;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderSourceEnum;
use app\common\enum\order\OrderStatusEnum;
use app\scan\model\store\Table as TableModel;
use app\common\model\order\OrderProductReturn;
use app\common\model\order\Order as OrderModel;
use app\common\service\order\OrderRefundService;
use app\common\service\order\OrderCompleteService;
use app\common\model\settings\Setting as SettingModel;
use app\common\service\product\factory\ProductFactory;
use app\cashier\service\order\paysuccess\type\MasterPaySuccessService;

/**
 * 普通订单模型
 */
class Order extends OrderModel
{
    /**
     * 隐藏字段
     * @var array
     */
    protected $hidden = [
        'update_time',
    ];

    /**
     * 用户中心订单列表
     */
    public function getList($params)
    {
        $model = $this;
        if (isset($params['shop_supplier_id']) && $params['shop_supplier_id']) {
            $model = $model->where('shop_supplier_id', '=', $params['shop_supplier_id']);
        }
        if (isset($params['eat_type']) && $params['eat_type']) {
            $model = $model->where('eat_type', '=', $params['eat_type']);
        }
        if (isset($params['search']) && $params['search']) {
            $model = $model->like('order_no', $params['search']);
        }
        if (isset($params['order_type']) && $params['order_type']) {
            $model = $model->where('order_type', '=', $params['eat_type']);
        }


        $startTime = 0;
        $endTime = 0;
        //查询时间
        switch ($params['time_type'] ?? 1) {
            case '1': //今天
                $startTime = strtotime(date('Y-m-d'));
                $endTime = $startTime + 86399;
                break;
            case '2': //昨天
                $startTime = strtotime("-1 days", strtotime(date('Y-m-d')));
                $endTime = $startTime + 86399;
                break;
            case '3': //一周
                $startTime = strtotime("-7 days", strtotime(date('Y-m-d')));
                $endTime = time();
                break;
        }
        if (isset($params['time']) && $params['time']) {
            $startTime = strtotime($params['time'][0]);
            $endTime = strtotime($params['time'][0]) + 86399;
        }
        if ($startTime && $endTime) {
            $model = $model->where('create_time', 'between', [$startTime, $endTime]);
        }

        switch ($params['dataType'] ?? 1) {
            case '1': //进行中
                $model = $model->where('order_status', '=', 10);
                break;
            case '2': //已完成
                $model = $model->where('pay_status', '=', 20)->where('order_status', '=', 30);
                break;
            case '3': //已取消
                $model = $model->where('order_status', '=', 20);
                break;
        }
        return $model->with(['product.image', 'supplier'])
            ->where('is_delete', '=', 0)
            ->where('delivery_type', 'in', [30, 40])
            ->where('eat_type', '<>', 0)
            ->order(['create_time' => 'desc'])
            ->field("*,FROM_UNIXTIME(pay_time,'%Y-%m-%d %H:%i:%s') as pay_time_text ")
            ->paginate($params);
    }

    /**
     * 取消订单
     */
    public function cancel($table_id)
    {
        $detail = $this->where('table_id', '=', $table_id)
            ->where('order_status', '=', 10)
            ->where('eat_type', '=', 10)
            ->find();

        if (!$detail) {
            TableModel::close($table_id);
            $this->error = "订单不存在";
            return false;
        }

        if ($detail['pay_status']['value'] == 20) {
            TableModel::close($table_id);
            $this->error = "订单已付款，不允许取消";
            return false;
        }
        if ($detail['order_status']['value'] != 10) {
            $this->error = "订单状态错误，不允许取消";
            return false;
        }
        return $detail->save(['order_status' => 20]);
    }


    /**
     * 取消订单
     */
    public function cancels()
    {
        if ($this->pay_status['value'] == 20) {
            $this->error = "订单已付款，不允许取消";
            return false;
        }
        if ($this->order_status['value'] != 10) {
            $this->error = "订单状态错误，不允许取消";
            return false;
        }
        return $this->save(['order_status' => 20]);
    }

    /**
     * 待支付订单详情
     */
    public static function getPayDetail($orderNo)
    {
        $model = new static();
        return $model->where(['order_no' => $orderNo, 'pay_status' => 10, 'is_delete' => 0])->with(['product', 'user', 'supplier'])->find();
    }

    /**
     * 设置错误信息
     */
    protected function setError($error)
    {
        empty($this->error) && $this->error = $error;
    }

    /**
     * 是否存在错误
     */
    public function hasError()
    {
        return !empty($this->error);
    }

    /**
     * 主订单购买的数量
     * 未取消的订单
     */
    public static function getHasBuyOrderNum($user_id, $product_id)
    {
        $model = new static();
        return $model->alias('order')->where('order.user_id', '=', $user_id)
            ->join('order_product', 'order_product.order_id = order.order_id', 'left')
            ->where('order_product.product_id', '=', $product_id)
            ->where('order.order_source', '=', OrderSourceEnum::MASTER)
            ->where('order.order_status', '<>', 21)
            ->sum('total_num');
    }

    //查询桌号信息
    public static function getTableInfo($table_id)
    {
        return (new static())->where('table_id', '=', $table_id)
            ->where('is_delete', '=', 0)
            ->order('order_id desc')
            ->find();
    }

    //查询桌号订单信息
    public static function getOrderInfo($table_id)
    {
        return (new static())->with('product')
            ->where('table_id', '=', $table_id)
            ->where('order_status', '=', 10)
            ->where('is_delete', '=', 0)
            ->find();
    }

    //退菜
    public function moveProduct($order_product_id, $num, $return_reason = '')
    {
        if ($this['order_status']['value'] != 10) {
            $this->error = "订单已完成,不允许退菜";
            return false;
        }
        if (count($this['product']) <= 1) {
            $this->error = "仅剩一个商品，不允许退菜，请选择退单";
            return false;
        }
        $orderProduct = OrderProduct::detail($order_product_id);
        if ($orderProduct['total_num'] < $num) {
            $this->error = "退菜数量不能大于当前商品数量";
            return false;
        }
        $this->startTrans();
        try {
            $isPay = $this['pay_status']['value'] == 20 ? 1 : 0;
            $orderProductNum = $orderProduct['total_num'];
            $orderProduct['total_num'] = $num;
            // 退回商品库存
            ProductFactory::getFactory($this['order_source'])->backProductStock([$orderProduct], $isPay);
            if ($orderProductNum == $num) {
                $orderProduct->delete();
            } else {
                $total_num = $orderProductNum - $num;
                $orderProduct->save([
                    'total_num' => $total_num,
                ]);
            }
            // 退菜记录
            if ($num > 0) {
                OrderProductReturn::add([
                    'order_id' => $this['order_id'],
                    'order_product_id' => $order_product_id,
                    'product_id' => $orderProduct['product_id'],
                    'num' => $num,
                    'reason' => $return_reason,
                ]);
            }
            //
            $this->reloadPrice($this['order_id']);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    //结账完成
    public function settle()
    {
        if ($this['pay_status']['value'] != 20) {
            $this->error = "订单未付款，不允许操作";
            return false;
        }
        if ($this['order_status']['value'] != 10) {
            $this->error = "订单状态错误，不允许操作";
            return false;
        }
        return $this->transaction(function () {
            // 更新订单状态：已发货、已收货
            $status = $this->save([
                'delivery_status' => 20,
                'delivery_time' => time(),
                'receipt_status' => 20,
                'receipt_time' => time(),
                'order_status' => OrderStatusEnum::COMPLETED
            ]);
            // 执行订单完成后的操作
            $OrderCompleteService = new OrderCompleteService(OrderTypeEnum::MASTER);
            $OrderCompleteService->complete([$this], static::$app_id);
            return $status;
        });
    }

    // 自助餐到期提醒
    public static function buffetTimeRemind($tid, $buffet_expired_time, $tablet_end_time_minute, $app_id = 0)
    {
        $second = $tablet_end_time_minute * 60;
        $now_timestamp = time();
        $buffet_remaining_time = $buffet_expired_time - $now_timestamp;
        $lock = Cache::get("{$app_id}_remind::{$tid}");
        if (($buffet_remaining_time < $second) && !$lock) {
            Cache::set("{$app_id}_remind::{$tid}", 1, $second);
            return 1;
        }
        return 0;
    }

    /**
     * 扫码端基础信息
     * @param $need_unsend_list     // 是否需有返回未送厨商品列表
     * @param $need_send_list       // 是否需有返回已送厨商品列表
     * @param $product_source       // 商品终端来源
     * @param $setting_data         // 后台配置(已查传入提升效率)
     * @return array
     */
    public function getScanOrderInfo($need_unsend_list = 0, $need_send_list = 0, $product_source = self::SCAN_PRODUCT_SOURCE, $setting_data = [])
    {
        /** @var OrderModel $detail */
        $detail = $this;
        $table_id = $detail['table_id'] ?? 0;
        // 自助餐设置
        $buffetSetting = $setting_data ? $setting_data[SettingEnum::BUFFET]['values'] : SettingModel::getSupplierItem(SettingEnum::BUFFET, $detail['shop_supplier_id'] ?? 0, $detail['app_id'] ?? 0);
        $is_lock = $detail->is_lock;
        $buffet_remaining_time = Order::getBuffetRemainingTime($detail['buffet_expired_time']);
        [$is_remain_continue, $remain_continue_notice_time, $remain_continue_time] = OrderModel::getBuffetRemain($detail['order_id']);
        $buffet_order_remaining_time = $buffet_remaining_time - $remain_continue_time * 60;
        $buffet_order_remaining_time = max($buffet_order_remaining_time, 0);
        $buffet = [
            'remind' => OrderModel::buffetTimeRemind($table_id, $detail['buffet_expired_time'], $buffetSetting['tablet_end_time'], $detail['app_id'] ?? 0),
            'minute' => (int)($buffetSetting['tablet_end_time'] ?? 5),  // 平板/扫码H5结束时间提醒
            'elapsed_time' => $detail['elapsed_time'],
            'buffet_remaining_time' => $buffet_remaining_time,  // 自助餐用餐剩余时间
            'buffet_order_remaining_time' => $buffet_order_remaining_time, // 自助餐点餐剩余时间
            'buffet_expired_time' => $detail['buffet_expired_time'],
            'is_buy_continue' => (int)($buffetSetting['is_buy_continue'] ?? 1),
            'is_remain_continue' => $is_remain_continue,       //  平板是否可继续点餐开关
            'remain_continue_notice_time' => $remain_continue_notice_time,     // 剩余xx分提醒不可继续点餐
            'remain_continue_time' => $remain_continue_time,    // 剩余xx分不可继续点餐
        ];
        $order = [
            'order_id' => $detail['order_id'],
            'is_must_notice' => $detail['is_must_notice'],
            'is_buffet' => $detail['is_buffet'],
            'meal_num' => $detail['meal_num'],
            'unsend_total_num' => OrderProduct::where('order_id', $detail['order_id'])
                ->where('add_source', OrderModel::SCAN_PRODUCT_SOURCE)
                ->where('is_send_kitchen', 0)
                ->where('batch_time', '=', 0)
                ->sum('total_num'),           // 未送厨商品数量
            'unsend_total_price' => OrderProduct::where('order_id', $detail['order_id'])
                ->where('add_source', OrderModel::SCAN_PRODUCT_SOURCE)
                ->where('is_send_kitchen', 0)
                ->where('batch_time', '=', 0)
                ->sum('total_price'),       // 未送厨商品总价
        ];
        $is_change_buffet = (new OrderBuffet)->field('buffet_id')->where('order_id', $detail['order_id'])->order('buffet_id', 'asc')->column('buffet_id');
        $is_change_delay = (new OrderDelay)->where('order_id', $detail['order_id'])->column('id');
        $is_change_product = (new OrderProduct)->where('order_id', $detail['order_id'])->where('is_send_kitchen', 1)->column(['product_id', 'total_num']);

        // 未送厨商品
        $unSendProductList = OrderProduct::field(['sum(total_num) as total_num', 'product_id', 'is_send_kitchen'])
            ->where('order_id', $detail['order_id'])
            ->where('add_source', OrderModel::SCAN_PRODUCT_SOURCE)
            ->where('is_send_kitchen', 0)
            ->where('batch_time', 0)    // 未下单的 1.0.9
            ->group("product_id")
            ->select()->toArray();
        $cache_id = md5(json_encode($is_change_buffet) . json_encode($is_change_delay) . json_encode($is_change_product) . json_encode($unSendProductList));

        // 操作返回数据
        $scanUnSendProductList = [];
        if ($need_unsend_list) {
            $scanUnSendProductList = OrderProduct::where('order_id', $detail['order_id'])
                ->where('add_source', $product_source)
                ->where('is_send_kitchen', 0)
                ->where('batch_time', 0)    // 未下单的 1.0.9
                ->select();
        }
        $scanSendProductList = [];
        if ($need_send_list) {
            $model = new OrderModel();
            $n_order = $model->getSendKitchen($table_id);
            if ($n_order) {
                $delay = $n_order['delay'];
                $buffetCustomerType = $n_order['buffetCustomerType'];
                $scanSendProductList = OrderProduct::getGroupByTime($n_order['order_id'], $buffetCustomerType, $delay, [], 1);
                array_multisort(array_column($scanSendProductList, 'timestamp'), SORT_DESC, $scanSendProductList); //SORT_DESC降序，SORT_ASC升序
            }
        }
        // 订单必点商品方案\商品ID
        $res = $detail->getSchemeProductList(true);
        $schemeProductList = $res['orderSchemeProductList'];
        $schemeProductIds = $res['scheme_product_ids'];
        //
        return compact('is_lock', 'buffet', 'order', 'cache_id', 'unSendProductList', 'scanUnSendProductList', 'scanSendProductList', 'schemeProductList', 'schemeProductIds');
    }
}
