<?php

namespace app\cashier\controller\order;

use hg\apidoc\annotation as Apidoc;
use app\cashier\controller\Controller;
use app\common\model\app\App as AppModel;
use app\shop\model_old\order\Order as OldOrderModel;
use app\common\service\order\OrderInvoicePrinterService;
use app\cashier\model\order\Order as OldCashierOrderModel;

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

    /**
     * @Apidoc\Title("取消订单")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/cashier/order.order/orderCancel")
     * @Apidoc\Param("order_id", type="int", require=true, default="", desc="订单id")
     * @Apidoc\Param("cancel_remark", type="string", require=false, default="", desc="取消原因")
     * @Apidoc\Returned()
     */
    public function orderCancel()
    {
        $data = $this->postData();
        // 
        $app = AppModel::where('uuid', request()->appId)->find();
        if (!$app->old_company_id) {
            return $this->renderError('未存在迁移库');
        }
        request()->appId = $app->old_company_id;
        // 
        if (mb_strlen($data['cancel_remark'] ?? '') > 50) {
            return $this->renderError('备注最长50个字符');
        }
        $model = OldOrderModel::detail($data['order_id']);
        // 
        if ($model && $model->delStay($data['order_id'], $data['cancel_remark'] ?? '')) {
            return $this->renderSuccess('操作成功');
        }
        // 
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/cashier/order.order/delete")
     * @Apidoc\Param("order_id", type="int", require=true, default="", desc="订单id")
     * @Apidoc\Returned()
     */
    public function delete($order_id)
    {
        $app = AppModel::where('uuid', request()->appId)->find();
        if (!$app->old_company_id) {
            return $this->renderError('未存在迁移库');
        }
        request()->appId = $app->old_company_id;
        // 
        $model = OldOrderModel::detail($order_id);
        if (!$model) {
            return $this->renderError('订单不存在');
        }
        if ($model->orderDelete()) {
            return $this->renderSuccess('删除成功');
        }
        return $this->renderError($model->getError() ?: '删除失败');
    }

    /**
     * @Apidoc\Title("打印小票 - 订单列表")
     * @Apidoc\Tag("打印小票")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/cashier/order.order/print")
     * @Apidoc\Param("order_id", type="int",require=true, default=0, desc="订单id")
     * @Apidoc\Param("print_lang", type="string",require=false, default="", desc="打印语言")
     */
    public function print($order_id = 0)
    {
        $app = AppModel::where('uuid', request()->appId)->find();
        if (!$app->old_company_id) {
            return $this->renderError('未存在迁移库');
        }
        request()->appId = $app->old_company_id;
        /** @var OldCashierOrderModel $order */
        $order = OldCashierOrderModel::detailWithTrashed($order_id, '*');
        if (!$order) {
            return $this->renderError('订单不存在');
        }
        //
        $printLang = $this->postData()['print_lang'] ?? '';
        //
        if (!$result = $order->printSmall($this->cashier['device_id'], $this->cashier['user']['cashier_id'], $printLang, false)) {
            request()->language = '';
            return $this->renderError($order->getError() ?: '打印失败，未连接打印机', ['printer_data' => false]);
        }
        request()->language = '';
        return $this->renderSuccess('发送成功', $result);
    }

    /**
     * @Apidoc\Title("打印发票 - 订单列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/cashier/order.Order/printInvoice")
     * @Apidoc\Param("order_id", type="int", require=true, default="", desc="订单id")
     * @Apidoc\Param("print_lang", type="string",require=false, default="", desc="打印语言")
     * @Apidoc\Param("invoice_info", type="array", desc="发票信息", children={
     *      @Apidoc\Property("company_name",type="string", desc="公司名称"),
     *      @Apidoc\Property("company_addr",type="string", desc="公司地址"),
     *      @Apidoc\Property("company_tax_number",type="string", desc="公司税号"),
     *      @Apidoc\Property("company_phone",type="string", desc="联系电话")
     * })
     * @Apidoc\Returned()
     */
    public function printInvoice()
    {
        $app = AppModel::where('uuid', request()->appId)->find();
        if (!$app->old_company_id) {
            return $this->renderError('未存在迁移库');
        }
        request()->newAppId = request()->appId;
        request()->oldAppId = $app->old_company_id;
        request()->appId = $app->old_company_id;
        // 
        $param = $this->postData();
        $order = OldOrderModel::detailWithTrashed($param['order_id'] ?? 0, '*');
        if (!$order) {
            return $this->renderError('订单不存在');
        }
        if (($order->order_status['value'] ?? '') != 30) {
            return $this->renderError('订单未完成');
        }
        if (($order->pay_status['value'] ?? '') == 10) {
            return $this->renderError('订单未支付');
        }
        //
        $param['shop_supplier_id'] =  $app->old_company_id;
        $param['app_id'] =   $app->old_company_id;
        // 发送打印
        $orderInvoicePrinterService = new OrderInvoicePrinterService;
        $printerData = $orderInvoicePrinterService->cashierPrint($param, false, $this->cashier['device_id']);
        request()->language = '';
        if (!$printerData) {
            return $this->renderError($orderInvoicePrinterService->getError() ?: '打印失败，未连接打印机');
        }
        return $this->renderSuccess('发送成功', [
            'printer_data' => $printerData
        ]);
    }

     /**
     * @Apidoc\Title("获取发票信息")
     * @Apidoc\Method ("GET")
     * @Apidoc\Url ("/index.php/cashier/order.Order/invoiceInfo")
     * @Apidoc\Param("order_id", type="int", require=true, default="", desc="订单id")
     * @Apidoc\Param("print_lang", type="string",require=false, default="", desc="打印语言")
     * @Apidoc\Param("invoice_info", type="array", desc="发票信息", children={
     *      @Apidoc\Property("company_name",type="string", desc="公司名称"),
     *      @Apidoc\Property("company_addr",type="string", desc="公司地址"),
     *      @Apidoc\Property("company_tax_number",type="string", desc="公司税号"),
     *      @Apidoc\Property("company_phone",type="string", desc="联系电话")
     * })
     * @Apidoc\Returned()
     */
    public function invoiceInfo()
    {
        $app = AppModel::where('uuid', request()->appId)->find();
        if (!$app->old_company_id) {
            return $this->renderError('未存在迁移库');
        }
        request()->newAppId = request()->appId;
        request()->oldAppId = $app->old_company_id;
        request()->appId = $app->old_company_id;
        // 
        $param = $this->postData();
        //
        $order = OldOrderModel::find($param['order_id'] ?? 0);
        if (!$order) {
            return $this->renderError('订单不存在');
        }
        // 
        return $this->renderSuccess('发送成功', $order->invoiceInfo);
    }

    /**
     * @Apidoc\Title("订单退款（v1.0.5）")
     * @Apidoc\Method ("POST, GET")
     * @Apidoc\Desc ("POST 请求时为提交退款数据，GET 请求时为获取退款数据")
     * @Apidoc\Url("/index.php/cashier/order.Order/orderRefund")
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
    public function orderRefund()
    {
        $app = AppModel::where('uuid', request()->appId)->find();
        if (!$app->old_company_id) {
            return $this->renderError('未存在迁移库');
        }
        request()->appId = $app->old_company_id;
        // 
        $params = $this->postData();
        $params['refund_type'] = 1;
        $params['refund_method'] = 1;
        $model = OldOrderModel::detail($params['sale_bill_uuid'] ?? $params['order_id'], ['payType', 'refundType', 'mergeList']);
        if ($this->request->isGet()) {
            $payment_records = $model->payTypeCellRefundMoneys();
            $can_return_amount = array_sum(array_column($payment_records, 'can_return_amount'));
            // $products = $model->getOrderProductList();
            $products = [];
            return $this->renderSuccess('', compact('payment_records', 'products', 'can_return_amount'));
        }
        if ($model?->orderRefund($params)) {
            return $this->renderSuccess('操作成功');
        }
        // 
        return $this->renderError($model?->getError() ?: '操作失败', $model?->getErrorData(), $model?->getErrorCode());
    }
}
