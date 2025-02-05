<?php

namespace app\common\service\order;

use think\facade\Cache;
use app\common\enum\order\OrderTypeEnum;
use app\common\model\order\Order as OrderModel;

/**
 * 订单服务类
 */
class OrderService
{
    /**
     * 订单模型类
     * @var array
     */
    private static $orderModelClass = [
        OrderTypeEnum::MASTER => 'app\common\model\order\Order'
    ];

    /**
     * 生成订单号
     */
    public static function createOrderNo()
    {
        return date('Ymd') . substr(implode('', array_map('ord', str_split(substr(uniqid(), 7, 13), 1))), 0, 8);
    }

    /**
     * 生成新版订单号
     * 訂單號：18位纯数字（前八位是年月日，第九位1是收銀/2是桌台/3是充值，后九位随机生成）
     * 订单来源 order_source 10-桌台 20-收银台
     */
    public static function createNewOrderNo($order_source)
    {
        if ($order_source == 10) {
            $type_num = 2;
        } else if ($order_source == 20) {
            $type_num = 1;
        }
        // 获取时间
        $micro_time = microtime(true);
        $date_part = date('Ymd', (int)$micro_time);
        $micro_seconds = sprintf("%.6f", $micro_time - floor($micro_time));
        $time_part = substr(str_replace('.', '', $micro_seconds), 0, 6); // 6位微秒
        //
        $n = substr(implode('', array_map('ord', str_split(substr(uniqid(), 7, 13), 1))), 0, 3);
        $orderNo = $date_part . $type_num . $time_part . $n;
        if (Cache::get('__CREATE_NEW_ORDERNO__' . $orderNo)) {
            return self::createNewOrderNo($order_source);
        }
        Cache::set('__CREATE_NEW_ORDERNO__' . $orderNo, 1, 5);
        //
        return $orderNo;
    }

    /**
     * 生成交易号
     */
    public static function createTradeNo()
    {
        $snowflake = new Snowflake(1);
        return $snowflake->next();
    }

    /**
     * 整理订单列表 (根据order_type获取不同类型的订单记录)
     */
    public static function getOrderList($data, $orderIndex = 'order', $with = [])
    {
        // 整理订单id
        $orderIds = [];
        foreach ($data as &$item) {
            $orderIds[$item['order_type']['value']][] = $item['order_id'];
        }
        // 获取订单列表
        $orderList = [];
        foreach ($orderIds as $orderType => $values) {
            $model = new OrderModel();
            $orderList[$orderType] = $model->getListByIds($values, $with);
        }
        // 格式化到数据源
        foreach ($data as $key => &$item) {
            if (!isset($orderList[$item['order_type']['value']][$item['order_id']])) {
                $item->delete();
                unset($data[$key]);
                continue;
            }
            $item[$orderIndex] = $orderList[$item['order_type']['value']][$item['order_id']];
        }
        return $data;
    }
}
