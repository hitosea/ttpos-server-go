<?php

namespace app\shop\controller\store;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\order\OrderOperationLog;
use app\shop\model\order\Order as OrderModel;

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
        $model = new OrderModel();
        $data = $this->postData();
        $data['shop_supplier_id'] = $this->store['user']['shop_supplier_id'];
        $data['order_type'] = 1;
        // 时间模式
        if (!isset($data['time_mode']) || !is_array($data['time_mode'])) {
            $data['time_mode'] = [0]; // 默认开台时间
        }
        if ($exportList = $model->exportList($dataType, $data)) {
            if (($data['request_type'] ?? '') == 1) {
                return $this->renderSuccess('操作成功');
            }
            return $exportList;
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("取消订单")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.operate/orderCancel")
     * @Apidoc\Param("order_id", type="int", require=true, default="", desc="订单id")
     * @Apidoc\Param("cancel_remark", type="string", require=false, default="", desc="取消原因")
     * @Apidoc\Returned()
     */
    public function orderCancel($order_id)
    {
        $data = $this->postData();
        if (mb_strlen($data['cancel_remark'] ?? '') > 50) {
            return $this->renderError('备注最长50个字符');
        }
        $model = OrderModel::detail($order_id);
        if ($model->delStay($order_id, $data['cancel_remark'])) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
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
    public function delete($order_id)
    {
        $model = OrderModel::detail($order_id);
        if (!$model) {
            return $this->renderError('订单不存在');
        }
        if ($model->orderDelete()) {
            return $this->renderSuccess('删除成功');
        }
        return $this->renderError($model->getError() ?: '删除失败');
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
        $params = $this->postData();
        $model = OrderModel::detail($order_id, ['payType', 'refundType', 'mergeList']);
        if ($this->request->isGet()) {
            $pay_list = $model->payTypeCellRefundMoneys();
            $product_list = $model->getOrderProductList();
            return $this->renderSuccess('', compact('pay_list', 'product_list'));
        }
        if ($model?->orderRefund($params)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model?->getError() ?: '操作失败', $model->getErrorData(), $model->getErrorCode());
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
     * @Apidoc\Param("payment_order_id", type="int",require=true, default=0, desc="订单支付id")
     * @Apidoc\Param("refund_money", type="int",require=true, default=0, desc="退款金额")
     * @Apidoc\Param("refund_destination_id", type="int",require=true, default=0, desc="退款ID")
     * @Apidoc\Param("bank_code", type="int",require=true, default=004, desc="退款bank_code")
     * @Apidoc\Param("account_no", type="string",require=true, default="1941288621", desc="退款账号")
     * @Apidoc\Param("account_name", type="string",require=true, default="MR.TAO ZHANG", desc="退款名称")
     */
    public function orderRefundAgain()
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
        $model = new OrderOperationLog;
        if ($model->refundAgain($payment_order_id, $refund_money, $refund_destination_id, $bank_code, $account_no, $account_name)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }
}
