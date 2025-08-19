<?php

namespace app\shop\controller\store;

use help\HttpHelp;
use app\common\model\order\Order;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\shop\service\order\ExportService;
use app\shop\model\order\Order as OrderModel;
use app\shop\model\shop\User as ShopUser;

/**
 * 订单操作
 * @Apidoc\Group("order")
 * @Apidoc\Sort(4)
 */
class Operate extends Controller
{
    /**
     * @Apidoc\Title("订单导出")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.operate/export")
     * @Apidoc\Param("order_no", type="string", require=false, default="", desc="订单号")
     * @Apidoc\Param("style_id", type="int", require=false, default="", desc="配送方式")
     * @Apidoc\Param("time_mode", type="array",require=true, default="0", desc="时间模式 0开台时间 1支付时间，可多选（v1.0.9）")
     * @Apidoc\Param("time", type="array", require=false, default="", desc="时间范围 [2024-01-01, 2024-01-11]")
     * @Apidoc\Param("dataType", type="string", require=false, default="all", desc="订单类型 all-全部 payment-待付款 process-进行中 complete-已完成 cancel-已取消")
     * @Apidoc\Returned()
     */
    public function export($dataType = 'all')
    {
        $url = 'http://nginx/api/v1/shop/order/export';
        $data = $this->postData();
        // 
        $data['page'] = 1;
        $data['list_rows'] = 1001;
        $res = HttpHelp::getRequest($url, $this->buildListQueryParams($data), [
            'Authorization: Bearer ' . (request()->header('token') ?: request()->param('token')),
            'Accept-Language: ' . (request()->header('language') ?: request()->param('language')),
        ]);
        if (!$res) {
            return $this->renderError('请求失败');
        } 
        $result = json_decode($res, true);
        if (($result['code'] ?? -1) != 0) {
            return $this->renderError($result['message'] ?? '请求失败');
        }
        if ($result['data']['meta']['total'] > 1000) {
            return $this->renderError('请选择具体时间段，最多可导出1000条以下的数据');
        }
        //  
        return (new ExportService())->orderList($result['data']['list']);
    }

    /**
     * 构建列表查询参数
     */
    private function buildListQueryParams($data)
    {
        // 搜索日期类型: '0'-全部 '1'-今天 '2'-昨天 '3'-本周
        $dateType = intval($data['time_type']) - 1;
        // 搜索订单类型: ''-全部 '10'-桌台 '20-点餐
        $billType = intval(trim($data['order_source'] ?? 0) ?: -1);
        if ($billType == 10) {
            $billType = 0;
        }
        if ($billType == 20) {
            $billType = 1;
        }
        // 搜索订单号
        $orderNo = $data['order_no'];
        // 搜索用餐方式: ''-全部 '30'-打包 '40'-堂食
        $diningMethod = intval(trim($data['style_id']) ?: -1);
        if ($diningMethod == 30) {
            $diningMethod = 1;
        }
        if ($diningMethod == 40) {
            $diningMethod = 0;
        }
        // 搜索开台时间、支付时间
        $enableCreateTime = false;
        $enablePayTime = false;
        $queryStartTime = 0;
        $queryEndTime = 0;
        $time = $data['time'] ?: [];
        if (!empty($time)) {
            $queryStartTime = strtotime($time[0]);
            $queryEndTime = strtotime($time[1]) + 86399;
            $timeMode = $data['time_mode'] ?: [];
            if (in_array(0, $timeMode)) {
                $enableCreateTime = true;
            }
            if (in_array(1, $timeMode)) {
                $enablePayTime = true;
            }
        }
        // 搜索订单状态
        $status = [
            'all' => -1,
            'payment' => 0,
            'complete' => 1,
            'cancel' => 2,
        ][$data['dataType']];

        return [
            'order_no' => $orderNo,
            'bill_type' =>$billType,
            'date_type' => $dateType,
            'dining_method' => $diningMethod,
            'status' => $status,
            'page_no' => $data['page'] ?? 1,
            'page_size' => $data['list_rows'] ?? 10,
            'enable_create_time' => $enableCreateTime,
            'enable_pay_time' => $enablePayTime,
            'query_start_time' => $queryStartTime,
            'query_end_time' => $queryEndTime,
        ];
    }

    /**
     * @Apidoc\Title("取消订单")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.operate/orderCancel")
     * @Apidoc\Param("order_id", type="int", require=true, default="", desc="订单id")
     * @Apidoc\Param("cancel_remark", type="string", require=false, default="", desc="取消原因")
     * @Apidoc\Returned()
     */
    public function orderCancel()
    {
        $data = $this->postData();
        if (mb_strlen($data['cancel_remark'] ?? '') > 50) {
            return $this->renderError('备注最长50个字符');
        }
        $order = new Order();
        $res = $order->delStay($data['order_id'], $data['cancel_remark'] ?? '');
        if (!$res) {
            return $this->renderError($order->getError() ?: '操作失败');
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("门店核销")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.operate/extract")
     * @Apidoc\Param("order_id", type="int", require=true, default="", desc="订单id")
     * @Apidoc\Returned()
     * NotParse
     */
    public function extract($order_id)
    {
        $model = OrderModel::detail($order_id);
        if ($model->verificationOrder()) {
            return $this->renderSuccess('核销成功');
        }
        return $this->renderError($model->getError() ?: '核销失败');
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.operate/delete")
     * @Apidoc\Param("order_id", type="int", require=true, default="", desc="订单id")
     * @Apidoc\Returned()
     */
    public function delete()
    {
        $data = $this->postData();
        // 请求获取充值订单列表接口
        $res = HttpHelp::deleteRequest('http://nginx/api/v1/shop/order/delete', json_encode([
            'sale_bill_uuid' => intval($data['sale_bill_uuid']),
            'sale_order_uuid' => intval($data['sale_order_uuid']),
        ]), [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
            'Content-Type: application/json; charset=utf-8',
        ]);
        if (!$res) {
            return $this->renderError('请求失败');
        } 
        $result = json_decode($res, true);
        if (($result['code'] ?? -1) != 0) {
            return $this->renderError($result['message'] ?? '请求失败');
        }
        return $this->renderSuccess('删除成功');
    }

    /**
     * @Apidoc\Title("订单退款（v1.0.5）")
     * @Apidoc\Method ("POST, GET")
     * @Apidoc\Desc ("POST 请求时为提交退款数据，GET 请求时为获取退款数据")
     * @Apidoc\Url("/index.php/shop/store.operate/orderRefund")
     * @Apidoc\Param("order_id", type="int",require=true, default=0, desc="订单id")
     * @Apidoc\Param("refund_type", type="int", require=true, desc="退款类型 1-整单退款 2-部分退款")
     * @Apidoc\Param("refund_method", type="int", require=true, desc="退款方式 1-系统退款 2-线下退款")
     * @Apidoc\Param("refund_product", type="array", require=false, desc="部分退款商品数组[{order_product_id,refund_num}] refund_type为2必填")
     * @Apidoc\Param("refund_buffet", type="array", require=false, desc="部分退款自助餐数组[{id,refund_num}] refund_type为2必填")
     * @Apidoc\Param("refund_delay", type="array", require=false, desc="部分退款加钟数组[{id,refund_num}] refund_type为2必填")
     * @Apidoc\Returned("pay_list", type="array", desc="支付列表，包含剩余可退金额，Get请求时返回", children={
     *      @Apidoc\Property("name",type="string", desc="支付名称"),
     *      @Apidoc\Property("price",type="string", desc="支付价格"),
     *      @Apidoc\Property("cell_refund_money",type="string", desc="剩余可退金额"),
     * })
     * @Apidoc\Returned("product_list", type="array", desc="产品列表，部分退款时使用，Get请求时返回, 之前的 order.order/orderProductList 接口可以不用了", children={
     *      @Apidoc\Returned("buffetList", type="array", desc="自助餐", children={
     *           @Apidoc\Property("...",type="...", desc="..."),
     *      }),
     *      @Apidoc\Returned("delayList", type="array", desc="加钟", children={
     *          @Apidoc\Property("...",type="...", desc="..."),
     *      }),
     *      @Apidoc\Returned("productList", type="array", desc="产品", children={
     *          @Apidoc\Property("...",type="...", desc="..."),
     *      }),
     * })
     */
    public function orderRefund($order_id = 0)
    {
        $data = $this->postData();
        $requeustData = [
            'sale_bill_uuid' => intval($data['sale_bill_uuid']),
            'sale_order_uuid' => intval($data['sale_order_uuid']),
        ];
        $requeustHeader = [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
            'Content-Type: application/json; charset=utf-8',
        ];

        if ($this->request->isGet()) {
            $res = HttpHelp::getRequest('http://nginx/api/v1/shop/order/return', $requeustData, $requeustHeader);
            if (!$res) {
                return $this->renderError('请求失败');
            }
            $result = json_decode($res, true);
            if (($result['code'] ?? -1) != 0) {
                return $this->renderError($result['message'] ?? '请求失败');
            }
            return $this->renderSuccess('', $result['data']);
        }

        // 退部分商品
        $requeustData['products'] = [];

        $refundProduct = $data['refund_product'] ?? [];
        foreach ($refundProduct as $item) {
            $requeustData['products'][] = [
                'sale_order_product_uuid' => intval($item['order_product_id']),
                'num' => floatval($item['refund_num']),
            ];
        }

        // 退款到银行账户
        $requeustData['bank_code'] = $data['bank_code'] ?? '';
        $requeustData['account_no'] = $data['account_no'] ?? '';
        $requeustData['account_name'] = $data['account_name'] ?? '';

        $saleBill = OrderModel::where('uuid', $data['sale_bill_uuid'])->find();
        if ($saleBill) {
            $token = ShopUser::getUserTokenByDutyNo($saleBill['duty_no']);
            if ($token) {
                $requeustHeader[0] = 'Authorization: Bearer ' . $token;
            }
        }
        // 退积分
        $requeustData['points'] = floatval($data['points'] ?? 0);

        $res = HttpHelp::postRequest('http://nginx/api/v1/shop/order/return', json_encode($requeustData), $requeustHeader);
        if (!$res) {
            return $this->renderError('请求失败');
        }
        $result = json_decode($res, true);
        $code = $result['code'] ?? -1;
        if ($code == -401) {
            return $this->renderError($result['message'] ?? '请求失败', ['code' => -901], 0);
        }
        if ($code != 0) {
            return $this->renderError($result['message'] ?? '请求失败');
        }

        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("订单退款商品列表（v1.0.5）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/store.operate/orderProductList")
     * @Apidoc\Param("order_id", type="int",require=true, default=0, desc="订单id")
     */
    public function orderProductList($order_id = 0)
    {
        $detail = OrderModel::detail($order_id, []);
        if (!$detail) {
            return $this->renderError('当前状态不可操作');
        }
        $list = $detail->getOrderProductList();
        return $this->renderSuccess('操作成功', $list);
    }

    /**
     * @Apidoc\Title("操作记录-重新退款（v1.0.9）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/store.operate/orderRefundAgain")
     * @Apidoc\Param("return_order_uuid", type="int",require=true, default=0, desc="订单支付id")
     * @Apidoc\Param("return_amount_uuid", type="int",require=true, default=0, desc="退款金额")
     * @Apidoc\Param("bank_code", type="int",require=true, default=004, desc="退款bank_code")
     * @Apidoc\Param("account_no", type="string",require=true, default="1941288621", desc="退款账号")
     * @Apidoc\Param("account_name", type="string",require=true, default="MR.TAO ZHANG", desc="退款名称")
     */
    public function orderRefundAgain()
    {
        $params = $this->postData();
        //
        $requeustHeader = [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
            'Content-Type: application/json; charset=utf-8',
        ];
        // 退款请求数据
        $requeustData = [
            'return_order_uuid' => intval($params['return_order_uuid']),
            'return_amount_uuid' => intval($params['return_amount_uuid']),
            'bank_code' => $params['bank_code'],
            'account_no' => $params['account_no'],
            'account_name' => $params['account_name'],
        ];
        // 退款请求
        $res = HttpHelp::postRequest('http://nginx/api/v1/shop/order/re_return', json_encode($requeustData), $requeustHeader);
        if (!$res) {
            return $this->renderError('请求失败');
        }
        $result = json_decode($res, true);
        if (($result['code'] ?? -1) != 0) {
            return $this->renderError($result['message'] ?? '请求失败');
        }
        return $this->renderSuccess('', $result['data']);
    }
}
