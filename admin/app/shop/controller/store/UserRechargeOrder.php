<?php

namespace app\shop\controller\store;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\order\UserRechargeOrder as UserRechargeOrderModel;
use app\shop\service\order\UserRechargeExportService;
use help\HttpHelp;
use think\facade\Log;

/**
 * 充值订单（v1.1.0）
 * @Apidoc\Group("order")
 * @Apidoc\Sort(20)
 */
class UserRechargeOrder extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.UserRechargeOrder/index")
     * @Apidoc\Param("time_type", type="int", require=false, default="0", desc="时间类型 0-全都 1-今天 2-昨天 3-周")
     * @Apidoc\Param("order_no", type="string", require=false, default="", desc="订单号")
     * @Apidoc\Param("time_mode", type="array",require=true, default={0}, desc="时间模式 0添加时间 1支付时间，可多选")
     * @Apidoc\Param("time", type="array", require=false, default="", desc="时间范围 [2024-01-01, 2024-01-11]")
     * @Apidoc\Param("data_type", type="string", require=false, default="all", desc="订单类型 all-全部 payment-待付款 process-进行中 complete-已完成 cancel-已取消")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\common\model\order\UserRechargeOrder\getList", desc="列表")
     */
    public function index()
    {
        // 获取参数
        $data = $this->postData();
        // 请求获取充值订单列表接口
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/recharge_order/list', $this->buildListQueryParams($data), [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
        ]);

        if (!$res) {
            return $this->renderError('请求失败');
        } 
        $result = json_decode($res, true);
        if (($result['code'] ?? -1) != 0) {
            return $this->renderError($result['message'] ?? '请求失败');
        }

        foreach ($result['data']['list'] as &$item) {
            $item['payment_time'] = $item['payment_time'] ? date('Y-m-d H:i:s', $item['payment_time']) : '';
        }

        return $this->renderSuccess('', $result['data']);
    }

    /**
     * 构建列表查询参数
     */
    private function buildListQueryParams($data)
    {
        // 搜索日期类型: 0全部 1今天 2昨天 3本周
        $dateType = $data['time_type'] - 1;
        // 搜索订单号
        $orderNo = $data['order_no'] ?? '';
        // 搜索日期范围
        $enableCreateTime = false;
        $enablePaymentTime = false;
        $queryStartTime = 0;
        $queryEndTime = 0;
        $times = $data['time'] ?? [];
        if (!empty($times)) {
            $queryStartTime = strtotime($times[0]);
            $queryEndTime = strtotime($times[1]);
            $enableCreateTime = in_array(0, $data['time_mode']);
            $enablePaymentTime = in_array(1, $data['time_mode']);
        }
        // 搜索订单状态: all全部 payment待付款 cancel已取消 complete已完成
        $status = [
            'all' => -1,
            'payment' => 0,
            'complete' => 1,
            'cancel' => 2,
        ][$data['data_type']];

        return [
            'date_type' => $dateType, // 日期类型: -1全部 0今天 1昨天 2本周
            'order_no' => $orderNo, // 订单号
            'enable_create_time' => $enableCreateTime, // true启用添加时间 false禁用添加时间
            'enable_payment_time' => $enablePaymentTime, // true启用支付时间 false禁用支付时间
            'query_start_time' => $queryStartTime, // 起始时间
            'query_end_time' => $queryEndTime, // 结束时间
            'status' => $status, // 订单状态: -1全部 0待付款 1已完成 2已取消
            'page' => intval($data['page'] ?? 1),
            'page_size' => intval($data['list_rows'] ?? 10),
        ];
    }

    /**
     * @Apidoc\Title("订单详情")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.UserRechargeOrder/detail")
     * @Apidoc\Param("id", type="int", require=true, default="", desc="订单id")
     * @Apidoc\Returned("detail", type="array", ref="app\common\model\order\UserRechargeOrder\detail", desc="订单详情")
     */
    public function detail()
    {
        $data = $this->postData();
        $uuid = $data['id'] ?? 0;

        // 请求获取充值订单列表接口
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/recharge_order/info', ['uuid' => $uuid], [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
        ]);

        if (!$res) {
            return $this->renderError('请求失败');
        } 
        $result = json_decode($res, true);
        if (($result['code'] ?? -1) != 0) {
            return $this->renderError($result['message'] ?? '请求失败');
        }

        $result['data']['payment_time'] = $result['data']['payment_time'] ? date('Y-m-d H:i:s', $result['data']['payment_time']) : '';

        foreach ($result['data']['operation_log']['list'] as &$item) {
            $item['create_time'] = $item['create_time'] ? date('Y-m-d H:i:s', $item['create_time']) : '';
        }

        return $this->renderSuccess('', $result['data']);
    }

    /**
     * @Apidoc\Title("订单导出")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.UserRechargeOrder/export")
     * @Apidoc\Param("time_type", type="int", require=false, default="0", desc="时间类型 0-全都 1-今天 2-昨天 3-周")
     * @Apidoc\Param("order_no", type="string", require=false, default="", desc="订单号")
     * @Apidoc\Param("time_mode", type="array",require=true, default={0}, desc="时间模式 0添加时间 1支付时间，可多选")
     * @Apidoc\Param("time", type="array", require=false, default="", desc="时间范围 [2024-01-01, 2024-01-11]")
     * @Apidoc\Param("data_type", type="string", require=false, default="all", desc="订单类型 all-全部 payment-待付款 process-进行中 complete-已完成 cancel-已取消")
     * @Apidoc\Returned()
     */
    public function export()
    {
        $data = $this->postData();
        $data['list_rows'] = 1000;
        Log::debug('data:' . json_encode($data));
        // 请求获取充值订单列表接口
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/recharge_order/list', $this->buildListQueryParams($data), [
            'Authorization: Bearer ' . $data['token'],
            'Accept-Language: ' . $data['language'],
        ]);

        if (!$res) {
            return $this->renderError('请求失败');
        } 
        $result = json_decode($res, true);
        if (($result['code'] ?? -1) != 0) {
            return $this->renderError($result['message'] ?? '请求失败');
        }

        $list = $result['data']['list'];

        // 导出excel文件
        return (new UserRechargeExportService)->orderList($list);
    }

    /**
     * @Apidoc\Title("取消")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.UserRechargeOrder/cancel")
     * @Apidoc\Param("id", type="int", require=true, default="", desc="订单id")
     * @Apidoc\Param("cancel_remark", type="string", require=false, default="", desc="取消备注")
     * @Apidoc\Returned()
     */
    public function cancel()
    {
        $data = $this->postData();
        $uuid = $data['id'] ?? 0;

        // 请求获取充值订单列表接口
        $res = HttpHelp::postRequest('http://nginx/api/v1/shop/recharge_order/cancel', json_encode(['uuid' => intval($uuid)]), [
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
        
        return $this->renderError('操作成功');
    }

    /**
     * @Apidoc\Title("退款")
     * @Apidoc\Method ("POST,GET")
     * @Apidoc\Desc ("POST 请求时为提交退款数据，GET 请求时为获取退款数据")
     * @Apidoc\Url("/index.php/shop/store.UserRechargeOrder/refund")
     * @Apidoc\Query("id", type="int",require=true, default=0, desc="充值订单id，GET时需要")
     * @Apidoc\Param("id", type="int",require=true, default=0, desc="充值订单id")
     * @Apidoc\Param("refund_type", type="int", require=true, desc="退款类型 1-整单退款 2-部分退款")
     * @Apidoc\Param("refund_money", type="int", require=false, desc="退款金额")
     * @Apidoc\Returned("info", type="array", desc="充值/用户信息，Get请求时返回", children={
     *      @Apidoc\Property("recharge_money",type="string", desc="充值金额"),
     *      @Apidoc\Property("gift_money",type="string", desc="赠送金额"),
     *      @Apidoc\Property("gift_point",type="string", desc="赠送积分"),
     *      @Apidoc\Property("cell_refund_money",type="string", desc="剩余可退金额"),
     *      @Apidoc\Returned("user", type="array", desc="用户信息，Get请求时返回", children={
     *       @Apidoc\Property("balance",type="string", desc="用户可用余额"),
     *       @Apidoc\Property("gift_balance",type="string", desc="赠送余额"),
     *       @Apidoc\Property("points",type="string", desc="用户可用积分"),
     *    }),
     * })
     * @Apidoc\Returned("pay_list", type="array", desc="支付列表，包含剩余可退金额，Get请求时返回", children={
     *      @Apidoc\Property("name",type="string", desc="支付名称"),
     *      @Apidoc\Property("price",type="string", desc="支付价格"),
     *      @Apidoc\Property("cell_refund_money",type="string", desc="剩余可退金额"),
     * })
     */
    public function refund()
    {
        $data = $this->request->param(); // get post
        if ($this->request->isGet()) {
            // 请求获取充值订单列表接口
            $res = HttpHelp::getRequest('http://nginx/api/v1/shop/recharge_order/refund', ['uuid' => $data['id'] ?? 0], [
                'Authorization: Bearer ' . request()->header('token'),
                'Accept-Language: ' . request()->header('language'),
            ]);
            if (!$res) {
                return $this->renderError('请求失败');
            } 
            $result = json_decode($res, true);
            if (($result['code'] ?? -1) != 0) {
                return $this->renderError($result['message'] ?? '请求失败');
            }

            $info = $result['data'];
            $pay_list = $result['data']['payment_records'];

            return $this->renderSuccess('', compact('info', 'pay_list'));
        }

        // 请求获取充值订单列表接口
        $res = HttpHelp::postRequest('http://nginx/api/v1/shop/recharge_order/refund', json_encode([
            'uuid' => intval($data['id']),
            'refund_type' => intval($data['refund_type']),
            'refund_money' => floatval($data['refund_money']),
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

        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("重新退款")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/store.UserRechargeOrder/refundAgain")
     * @Apidoc\Param("payment_order_id", type="int",require=true, default=0, desc="订单支付id")
     * @Apidoc\Param("refund_money", type="int",require=true, default=0, desc="退款金额")
     * @Apidoc\Param("refund_destination_id", type="int",require=true, default=0, desc="退款ID")
     * @Apidoc\Param("bank_code", type="int",require=true, default=004, desc="退款bank_code")
     * @Apidoc\Param("account_no", type="string",require=true, default="1941288621", desc="退款账号")
     * @Apidoc\Param("account_name", type="string",require=true, default="MR.TAO ZHANG", desc="退款名称")
     */
    public function refundAgain()
    {
        $params = $this->postData();
        $payment_order_id = $params['payment_order_id'] ?? 0;
        $refund_money = $params['refund_money'] ?? 0;
        $refund_destination_id = $params['refund_destination_id'] ?? 0;
        //
        $bank_code = isset($params['bank_code']) ? $params['bank_code'] : 0;
        $account_no = isset($params['account_no']) ? $params['account_no'] :'';
        $account_name = isset($params['account_name']) ? $params['account_name'] : '';
        //
        $model = new UserRechargeOrderModel;
        if ($model->refundAgain($payment_order_id, $refund_money, $refund_destination_id, $bank_code, $account_no, $account_name)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }
}
