<?php

namespace app\job\event;

use think\facade\Cache;
use app\common\library\helper;
use app\job\service\OrderService;
use app\common\enum\order\OrderTypeEnum;
use app\job\model\order\Order as OrderModel;
use app\common\service\order\OrderCompleteService;
use app\common\model\settings\Setting as SettingModel;
use app\common\model\order\OrderDeliver as OrderDeliverModel;

/**
 * 订单事件管理
 */
class Order
{
    // 模型
    private $model;

    // 应用id
    private $appId;

    /**
     * 执行函数
     */
    public function handle($app_id)
    {
        try {
            $this->appId = $app_id;
            $this->model = new OrderModel();
            // 普通订单行为管理
            $this->master();
        } catch (\Throwable $e) {
            echo 'ERROR ORDER: ' . $e->getMessage() . PHP_EOL;
            log_write('ORDER TASK : ' . $app_id . '__ ' . $e->getMessage(), 'task');
        }
        return true;
    }

    /**
     * 普通订单行为管理
     * 1分钟执行一次
     */
    private function master()
    {
        $key = "task_space__order__{$this->appId}";
        if (Cache::has($key)) return true;
        // 获取商城交易设置
        $config = SettingModel::getItem('trade', $this->appId);
        $this->model->transaction(function () use ($config) {
            // 已支付订单自动核销(桌台订单除外)
            $this->receive($config['order']['receive_days']);
            // 已支付桌台订单自动核销
            $this->receiveTable();
        });
        Cache::set($key, time(), 60);
        return true;
    }


    /**
     * 已支付桌台订单自动核销
     */
    private function receiveTable()
    {
        // 订单id集
        $orderId_arr = $this->model->where('pay_status', '=', 20)
            ->where('order_status', '=', 10)
            ->where('eat_type', '=', 10)
            ->where('auto_close', '=', 0)
            ->where('close_time', '<=', time())
            ->column('order_id');
        $orderIds = helper::getArrayColumnIds($orderId_arr);
        if (!empty($orderIds)) {
            // 更新订单状态
            $this->model->onBatchUpdate($orderIds, [
                'receipt_status' => 20,
                'receipt_time' => time(),
                'order_status' => 30,
                'delivery_status' => 20,
                'delivery_time' => time(),
            ]);
            // 批量处理已完成的订单
            $this->onReceiveCompleted($orderIds);
            // 记录日志
            $this->dologs('receiveTable', [
                'orderIds' => json_encode($orderIds),
            ]);
        }
        return true;
    }

    /**
     * 已支付订单自动核销
     */
    private function receive($receiveDays)
    {
        // 截止时间
        if ($receiveDays <= 0) return false;
        $deadlineTime = time() - ($receiveDays * 60);
        // 条件
        // 订单id集
        $orderId_arr = $this->model->where('pay_status', '=', 20)
            ->where('order_status', '=', 10)
            ->where('eat_type', '<>', 10)
            ->where('pay_time', '<=', $deadlineTime)
            ->column('order_id');
        $orderIds = helper::getArrayColumnIds($orderId_arr);
        if (!empty($orderIds)) {
            // 更新订单状态
            $this->model->onBatchUpdate($orderIds, [
                'receipt_status' => 20,
                'receipt_time' => time(),
                'order_status' => 30
            ]);
            $this->model->onBatchUpdateStatus($orderIds, [
                'delivery_status' => 20,
                'delivery_time' => time(),
            ]);
            // 批量处理已完成的订单
            $this->onReceiveCompleted($orderIds);
            // 记录日志
            $this->dologs('receive', [
                'receive_days' => $receiveDays,
                'deadline_time' => $deadlineTime,
                'orderIds' => json_encode($orderIds),
            ]);
        }
        return true;
    }

    /**
     * 批量处理已完成的订单
     */
    private function onReceiveCompleted($orderIds)
    {
        // 获取已完成的订单列表
        $list = $this->model->getReceiveList($orderIds, ['user']);
        if ($list->isEmpty()) return false;
        // 执行订单完成后的操作
        $OrderCompleteService = new OrderCompleteService(OrderTypeEnum::MASTER);
        $OrderCompleteService->complete($list, $this->appId);
        return true;
    }


    /**
     * 记录日志
     */
    private function dologs($method, $params = [])
    {
        $value = 'behavior Order --' . $method;
        foreach ($params as $key => $val)
            $value .= ' --' . $key . ' ' . $val;
        return log_write($value, 'task');
    }
}
