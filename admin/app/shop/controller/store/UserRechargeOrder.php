<?php

namespace app\shop\controller\store;

use app\common\library\helper;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\order\UserRechargeOrderOperationLog;
use app\common\model\order\UserRechargeOrder as UserRechargeOrderModel;

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
        // 订单列表
        $model = new UserRechargeOrderModel();
        $data = $this->postData();
        $dataType = $data['data_type'] ?? 'all';
        // 时间类型
        if (isset($data['time_type']) && !in_array($data['time_type'], ['', 0, 1, 2, 3])) {
            return $this->renderError('时间类型参数错误');
        }
        // 时间模式
        if (isset($data['time_mode']) && !is_array($data['time_mode'])) {
            return $this->renderError('时间模式参数错误');
        }
        // 订单类型
        if (!in_array($dataType, ['all', 'payment', 'process', 'complete', 'cancel'])) {
            return $this->renderError('订单类型参数错误');
        }
        //
        $data['order_type'] = 1;
        $data['shop_supplier_id'] = $this->store['user']['shop_supplier_id'];
        $list = $model->getList($dataType, $data);
        $order_count = [
            'order_count' => [
                'all' => $model->getCount('all', $data),
                'payment' => $model->getCount('payment', $data),
                'process' => $model->getCount('process', $data),
                'complete' => $model->getCount('complete', $data),
                'cancel' => $model->getCount('cancel', $data),
            ],
        ];
        return $this->renderSuccess('', compact('list', 'order_count'));
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
        $id = $data['id'] ?? 0;
        // 订单详情
        $detail = UserRechargeOrderModel::detail($id);
        // 操作日志
        $operationLog = UserRechargeOrderOperationLog::getLogList($id);
        //
        return $this->renderSuccess('', compact('detail', 'operationLog'));
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
        $model = new UserRechargeOrderModel();
        $data = $this->postData();
        $data['list_rows'] = 1000;
        $dataType = $data['data_type'] ?? 'all';
        $data['shop_supplier_id'] = $this->store['user']['shop_supplier_id'];
        // 时间类型
        if (isset($data['time_type']) && !in_array($data['time_type'], [0, 1, 2, 3])) {
            return $this->renderError('时间类型参数错误');
        }
        // 时间模式
        if (isset($data['time_mode']) && !is_array($data['time_mode'])) {
            return $this->renderError('时间模式参数错误');
        }
        // 订单类型
        if (!in_array($dataType, ['all', 'payment', 'process', 'complete', 'cancel'])) {
            return $this->renderError('订单类型参数错误');
        }
        if ($model->exportList($dataType, $data)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
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
        /** @var UserRechargeOrderModel $model */
        $model = UserRechargeOrderModel::detail($data['id'] ?? 0);
        if (!$model) {
            return $this->renderError('数据不存在');
        }
        if ($model->cancel($data)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
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
        /** @var UserRechargeOrderModel $model */
        $model = UserRechargeOrderModel::detail($data['id'] ?? 0);
        if (!$model) {
            return $this->renderError('数据不存在');
        }
        if ($this->request->isGet()) {
            $info = $model->toArray();
            $info['cell_refund_money'] = floatval(helper::bcsub($info['recharge_money'], $info['refund_money']));
            $pay_list = $model->payTypeCellRefundMoneys();
            return $this->renderSuccess('', compact('info', 'pay_list'));
        }
        if ($model?->refund($data)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model?->getError() ?: '操作失败', $model->getErrorData(), $model->getErrorCode());
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
