<?php

namespace app\cashier\controller\order;

use hg\apidoc\annotation as Apidoc;
use app\cashier\controller\Controller;
use app\shop\model_old\order\Order as OldOrderModel;
use app\common\model\app\App as AppModel;

/**
 * 订单
 * @Apidoc\Group("order")
 * @Apidoc\Sort(1)
 */
class Order extends Controller
{
    /**
     * @Apidoc\Title("订单列表")
     * @Apidoc\Tag("订单列表")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/cashier/order.order/index")
     * @Apidoc\Param("eat_type", type="int",require=true, default="10", desc="订单类型 0-全部,10-收银订单，20-桌台订单")
     * @Apidoc\Param("time_type", type="int",require=true, default="1", desc="时间类型 0-全都,1-今天,2-昨天,3-周")
     * @Apidoc\Param("time_mode", type="array",require=true, default="0", desc="时间模式 0开台时间 1支付时间，可多选（v1.0.9）")
     * @Apidoc\Param("time", type="array",require=true, default="", desc="时间范围 [2024-01-01, 2024-01-11]")
     * @Apidoc\Param("search", type="string",require=true, default="", desc="订单号")
     * @Apidoc\Param("order_type", type="string",require=true, default="", desc="用餐方式 0-外卖,1-店内")
     * @Apidoc\Param("dataType", type="string", require=false, default="all", desc="订单类型 all-全部 payment-待付款 process-进行中 complete-已完成 cancel-已取消（v1.1.0以上版本）")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list",type="array",ref="app\shop\model\order\Order\getList")
     */
    public function index()
    {
        $data = $this->postData();
        $app = AppModel::where('uuid', request()->appId)->find();
        if (!$app->old_company_id) {
            return $this->renderError('未存在迁移库');
        }
        request()->appId = $app->old_company_id;
        // 
        $result = OldOrderModel::getLists($data);
        foreach ($result['list'] as $key => $value) {
            $result['list'][$key]['finish_time'] = $value['finish_time'] == '-' ? 0 : strtotime($value['finish_time']);
            foreach ($value['sale_orders'] as $subKey => $subValue) {
                $result['list'][$key]['sale_orders'][$subKey]['finish_time'] = $subValue['finish_time'] == '-' ? 0 : strtotime($subValue['finish_time']);
            }
        }
        // 
        return $this->renderSuccess('', $result);
    }

    /**
     * @Apidoc\Title("订单详情")
     * @Apidoc\Tag("订单详情")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/cashier/order.order/detail")
     */
    public function detail($sale_bill_uuid)
    {
        $app = AppModel::where('uuid', request()->appId)->find();
        if (!$app->old_company_id) {
            return $this->renderError('未存在迁移库');
        }
        request()->appId = $app->old_company_id;
        // 
        $result = OldOrderModel::details($sale_bill_uuid);
        foreach ($result['detail']['pay_types'] as $subKey => $subValue) {
            $result['detail']['pay_types'][$subKey]['code'] = $subValue['code'] . '';
        }
        $result['detail']['order_amount'] = floatval($result['detail']['order_amount']);
        $result['detail']['payment_amount'] = floatval($result['detail']['payment_amount']);
        $result['detail']['refund_amount'] = floatval($result['detail']['refund_amount']);
        $result['detail']['create_time'] = !$result['detail']['create_time'] ? 0 : strtotime($result['detail']['create_time']);
        $result['detail']['finish_time'] = ($result['detail']['finish_time'] == '-' ? 0 : strtotime($result['detail']['finish_time'])) ?: 0;
        //
        foreach ($result['detail']['sale_orders'] ?? [] as $subKey => $subValue) {
            $result['detail']['sale_orders'][$subKey]['finish_time'] = ($subValue['finish_time'] == '-' ? 0 : strtotime($subValue['finish_time'])) ?: 0;
            $result['detail']['sale_orders'][$subKey]['order_amount'] = floatval($subValue['order_amount']);
            $result['detail']['sale_orders'][$subKey]['payment_amount'] = floatval($subValue['payment_amount']);
            $result['detail']['sale_orders'][$subKey]['refund_amount'] = floatval($subValue['refund_amount']);
            // 
            foreach ($subValue['products'] ?? [] as $key => $subValue) {
                $result['detail']['sale_orders'][$subKey]['products'][$key]['refund_amount'] = floatval($subValue['refund_amount']);
            }
        }
        // 
        foreach ($result['operation_log']['list'] as $subKey => $subValue) {
            $result['operation_log']['list'][$subKey]['create_time'] = strtotime($subValue['create_time']) ?: 0;
        }
        // 
        return $this->renderSuccess('', $result);
    }
}
