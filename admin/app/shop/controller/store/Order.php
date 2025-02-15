<?php

namespace app\shop\controller\store;

use help\HttpHelp;
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
        $data = $this->postData();
        // 订单来源 0-全都 10-桌台 20-收银
        $bill_type = (trim($data['order_source'] ?? -1) ?: -1);
        if ($bill_type == 10) {
            $bill_type = 0;
        }
        if ($bill_type == 20) {
            $bill_type = 1;
        }
        // 点餐方式
        $dining_method = trim($data['style_id'] ?? -1) ?: -1;
        if ($dining_method == 30) {
            $dining_method = 1;
        }
        if ($dining_method == 40) {
            $dining_method = 0;
        }
        // 账单状态, -1=全都、 0=待付款、1=已完成、2=已取消
        $status = -1;
        switch ($dataType) {
            case 'payment':
                $status = 0;
                break;
            case 'process':
                $status = 1;
                break;
            case 'complete':
                $status = 2;
                break;
            case 'cancel':
                $status = 3;
                break;
        }
        // 
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/order/list', [
            'order_no' => $data['order_no'] ?? '',
            'bill_type' =>$bill_type,
            'date_type' => (trim($data['time_type'] ?? '') ?: 0) - 1,
            'dining_method' => $dining_method,
            'status' => $status,
            'page_no' => $data['page'] ?? 1,
            'page_size' => $data['list_rows'] ?? 10,
            'enable_create_time' => in_array(0, $data['time_mode'] ?? []),
            'enable_pay_time' =>in_array(1, $data['time_mode'] ?? []),
            'query_start_time' =>($data['time'][0] ?? '') ? strtotime($data['time'][0]) : 0,
            'query_end_time' =>($data['time'][1] ?? '') ? strtotime($data['time'][1]) : 0,
        ], [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
        ]);
        if (!$res) {
            return $this->renderError('请求失败');
        } 
        $result = json_decode($res, true);
        if (($result['code'] ?? 0) != 1) {
            return $this->renderError($result['message'] ?? '请求失败');
        }
        // 
        foreach ($result['data']['list'] as $key => &$item) {
            $item['finish_time'] = $item['finish_time'] ? date('Y-m-d H:i:s', $item['finish_time']) : '';
            if ($item['sale_orders']) {
                foreach ($item['sale_orders'] as $subKey => &$subItem) {
                    $subItem['finish_time'] = $subItem['finish_time'] ? date('Y-m-d H:i:s', $subItem['finish_time']) : '';
                }
            }
        }
        //
        $result['data']['ex_style']  = DeliveryTypeEnum::store();
        // 
        return $this->renderSuccess('', $result['data']);
    }

    /**
     * @Apidoc\Title("订单详情")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.order/detail")
     * @Apidoc\Param("sale_bill_uuid", type="int", require=true, default="", desc="销售账单UUID")
     * @Apidoc\Param("sale_order_uuid", type="int", require=true, default="", desc="销售订单UUID 当查看子订单信息的时候才需要传")
     * @Apidoc\Returned("detail", type="array", ref="app\shop\model\order\Order\detail", desc="订单详情")
     */
    public function detail($sale_bill_uuid, $sale_order_uuid = 0)
    {
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/order/info', [
            'sale_bill_uuid' => $sale_bill_uuid,
            'sale_order_uuid' => $sale_order_uuid,
        ], [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
        ]);
        if (!$res) {
            return $this->renderError('请求失败');
        } 
        $result = json_decode($res, true);
        if (($result['code'] ?? 0) != 1) {
            return $this->renderError($result['message'] ?? '请求失败');
        }
        return $this->renderSuccess('', $result['data']);

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
