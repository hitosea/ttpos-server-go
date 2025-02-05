<?php

namespace app\shop\controller\store;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\order\OrderProductFree;
use app\common\model\order\OrderOperationLog;
use app\shop\model\order\Order as OrderModel;
use app\common\enum\settings\DeliveryTypeEnum;

/**
 * 订单管理
 * @Apidoc\Group("order")
 * @Apidoc\Sort(4)
 */
class Order extends Controller
{
    /**
     * @Apidoc\Title("订单列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.order/index")
     * @Apidoc\Param("time_type", type="int", require=false, default="", desc="时间类型 0-全都 1-今天 2-昨天 3-周")
     * @Apidoc\Param("order_source", type="int", require=false, default="", desc="订单来源 0-全都 10-桌台 20-收银")
     * @Apidoc\Param("order_no", type="string", require=false, default="", desc="订单号")
     * @Apidoc\Param("style_id", type="int", require=false, default="", desc="配送方式")
     * @Apidoc\Param("time_mode", type="array",require=true, default="0", desc="时间模式 0开台时间 1支付时间，可多选（v1.0.9）")
     * @Apidoc\Param("time", type="array", require=false, default="", desc="时间范围 [2024-01-01, 2024-01-11]")
     * @Apidoc\Param("dataType", type="string", require=false, default="all", desc="订单类型 all-全部 payment-待付款 process-进行中 complete-已完成 cancel-已取消")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\order\Order\getList", desc="列表")
     */
    public function index($dataType = 'all')
    {
        // 订单列表
        $model = new OrderModel();
        $data = $this->postData();
        // 时间模式
        if (!isset($data['time_mode']) || !is_array($data['time_mode'])) {
            $data['time_mode'] = [0]; // 默认开台时间
        }
        //
        $data['order_type'] = 1;
        $data['parent_id'] = 0;
        $data['shop_supplier_id'] = $this->store['user']['shop_supplier_id'];
        $list = $model->getList($dataType, $data);
        foreach ($list as $key => $item) {
            // 是否显示退款按钮 1-显示 0-隐藏
            /** @var OrderModel $item */
            [$list[$key]['is_refund_button'], $list[$key]['is_cancel_button']] = $item->getButtonStatus($item);
            if ($item['subOrder']) {
                foreach ($item['subOrder'] as $subKey => $subItem) {
                    /** @var OrderModel $subItem */
                    [$list[$key]['subOrder'][$subKey]['is_refund_button'], $list[$key]['subOrder'][$subKey]['is_cancel_button']] = $subItem->getButtonStatus($subItem);
                }
            }
            // 拆单主单支付方式去重
            if ($item['parent_id'] == 0 && count($item['subOrder']) > 0) {
                $payTypes = $item['payType']->toArray();
                $uniquePayTypes = [];
                foreach ($payTypes as $payType) {
                    $uniquePayTypes[$payType['value']] = $payType;
                }
                $item['payType'] = new \think\Collection(array_values($uniquePayTypes));
            }
        }
        $order_count = [
            'order_count' => [
                'all' => $model->getCount('all', $data),
                'payment' => $model->getCount('payment', $data),
                'process' => $model->getCount('process', $data),
                'complete' => $model->getCount('complete', $data),
                'cancel' => $model->getCount('cancel', $data),
            ],
        ];
        $ex_style = DeliveryTypeEnum::store();
        return $this->renderSuccess('', compact('list', 'ex_style', 'order_count'));
    }

    /**
     * @Apidoc\Title("订单详情")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.order/detail")
     * @Apidoc\Param("order_id", type="int", require=true, default="", desc="订单id")
     * @Apidoc\Returned("detail", type="array", ref="app\shop\model\order\Order\detail", desc="订单详情")
     */
    public function detail($order_id)
    {
        // 订单详情
        /** @var OrderModel $detail */
        $detail = OrderModel::detailWithTrashed($order_id, null, ["'' as free_tag_text"]);
        if (isset($detail['pay_time']) && $detail['pay_time'] > 0) {
            $detail['pay_time'] = date('Y-m-d H:i:s', $detail['pay_time']);
        }
        if (isset($detail['delivery_time']) && $detail['delivery_time'] != '') {
            $detail['delivery_time'] = date('Y-m-d H:i:s', $detail['delivery_time']);
        }
        //
        if ($detail) {
            $orderProductIds = array_column($detail['product']->toArray(), 'order_product_id');
            $orderProductFrees = OrderProductFree::where('order_product_id', 'in', $orderProductIds)->select()->toArray();
            foreach ($detail['product'] as &$orderProduct) {
                $orderProduct->free_tag_text = $orderProduct->getFreeTagText($orderProductFrees);
            }
            // 是否显示退款按钮 1-显示 0-隐藏
            [$detail['is_refund_button'], $detail['is_cancel_button']] = $detail->getButtonStatus($detail);
            // 拆单主单支付方式去重
            if ($detail['parent_id'] == 0 && count($detail['subOrder']) > 0) {
                $payTypes = $detail['payType']->toArray();
                $uniquePayTypes = [];
                foreach ($payTypes as $payType) {
                    $uniquePayTypes[$payType['value']] = $payType;
                }
                $detail['payType'] = new \think\Collection(array_values($uniquePayTypes));
            }
        }
        // 操作日志
        $operationOrderId = $detail['parent_id'] > 0 ? $detail['parent_id'] : $order_id;
        $operationLog = OrderOperationLog::getLogList($operationOrderId);
        //
        return $this->renderSuccess('', compact('detail', 'operationLog'));
    }
}
