<?php

namespace app\shop\controller\store;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;


/**
 * 外送订单管理
 * @Apidoc\Group("order")
 * @Apidoc\Sort(4)
 */
class DeliveryOrder extends Controller
{
    /**
     * @Apidoc\Title("外送订单列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.DeliveryOrder/index")
     * @Apidoc\Param("time_type", type="int", require=false, default="", desc="时间类型 0-全都 1-今天 2-昨天 3-本周")
     * @Apidoc\Param("serial_no", type="string", require=false, default="", desc="外卖序号")
     * @Apidoc\Param("order_no", type="string", require=false, default="", desc="订单号")
     * @Apidoc\Param("time_mode", type="array",require=true, default="0", desc="时间模式 0下单时间 1完成时间，可多选（v2.3.0）")
     * @Apidoc\Param("time", type="array", require=false, default="", desc="时间范围 [2025-01-01, 2025-01-11]")
     * @Apidoc\Param("dataType", type="string", require=false, default="all", desc="订单类型 all-全部 pending_payment-待付款 awaiting_delivery-待配送 delivering-配送中 cancelled-已取消 completed-已完成")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("data", type="array", desc="外送订单列表")
     * @Apidoc\Returned("data[].uuid", type="biginteger", desc="订单唯一标识")
     * @Apidoc\Returned("data[].serial_no", type="string", desc="外卖序号")
     * @Apidoc\Returned("data[].order_no", type="string", desc="订单号")
     * @Apidoc\Returned("data[].status", type="int", desc="状态")
     * @Apidoc\Returned("data[].create_time", type="int", desc="下单时间（时间戳）")
     * @Apidoc\Returned("data[].pay_time", type="int", desc="支付时间（时间戳）")
     * @Apidoc\Returned("data[].order_amount", type="string", desc="订单金额")
     * @Apidoc\Returned("data[].payment_amount", type="string", desc="实际支付金额")
     * @Apidoc\Returned("data[].contact_person", type="string", desc="联系人")
     * @Apidoc\Returned("data[].contact_phone", type="string", desc="联系电话")
     * @Apidoc\Returned("data[].delivery_fee", type="string", desc="配送费")
     * @Apidoc\Returned("data[].payment_method", type="string", desc="支付方式")
     */
    public function index()
    {
        // $data = $this->postData();
        //
        // $res = HttpHelp::getRequest('http://nginx/api/v1/shop/delivery_order/list', $this->buildListQueryParams($data), [
        //     'Authorization: Bearer ' . request()->header('token'),
        //     'Accept-Language: ' . request()->header('language'),
        // ]);
        // if (!$res) {
        //     return $this->renderError('请求失败');
        // } 
        // $result = json_decode($res, true);
        // if (($result['code'] ?? -1) != 0) {
        //     return $this->renderError($result['message'] ?? '请求失败');
        // }

        return $this->renderSuccess('', []);
    }


    /**
     * @Apidoc\Title("外送订单详情")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.DeliveryOrder/detail")
     * @Apidoc\Param("sale_bill_uuid", type="int", require=true, default="", desc="销售账单UUID")
     * @Apidoc\Param("sale_order_uuid", type="int", require=true, default="", desc="销售订单UUID 当查看子订单信息的时候才需要传")
     * @Apidoc\Returned("data.detail", type="object", desc="订单详情")
     * @Apidoc\Returned("data.detail.order_no", type="string", desc="订单号")
     * @Apidoc\Returned("data.detail.member_name", type="string", desc="会员名称")
     * @Apidoc\Returned("data.detail.member_uuid", type="string", desc="会员UUID")
     * @Apidoc\Returned("data.detail.order_amount", type="float", desc="订单金额")
     * @Apidoc\Returned("data.detail.payment_amount", type="float", desc="实付金额")
     * @Apidoc\Returned("data.detail.refund_amount", type="float", desc="退款金额")
     * @Apidoc\Returned("data.detail.pay_types", type="array", desc="支付方式列表")
     * @Apidoc\Returned("data.detail.remark", type="string", desc="备注")
     * @Apidoc\Returned("data.detail.serial_no", type="string", desc="外卖序号")
     * @Apidoc\Returned("data.detail.status", type="int", desc="订单状态")
     * @Apidoc\Returned("data.detail.create_time", type="string", desc="创建时间")
     * @Apidoc\Returned("data.detail.finish_time", type="string", desc="完成时间")
     * @Apidoc\Returned("data.detail.cashier_name", type="string", desc="收银员名称")
     * @Apidoc\Returned("data.detail.sale_orders", type="array", desc="销售订单列表")
     * @Apidoc\Returned("data.detail.sale_orders[].sale_order_uuid", type="string", desc="销售订单UUID")
     * @Apidoc\Returned("data.detail.sale_orders[].bill_type", type="int", desc="单据类型")
     * @Apidoc\Returned("data.detail.sale_orders[].dining_method", type="int", desc="用餐方式")
     * @Apidoc\Returned("data.detail.sale_orders[].serial_no", type="string", desc="销售订单流水号")
     * @Apidoc\Returned("data.detail.sale_orders[].order_no", type="string", desc="订单号")
     * @Apidoc\Returned("data.detail.sale_orders[].status", type="int", desc="销售订单状态")
     * @Apidoc\Returned("data.detail.sale_orders[].finish_time", type="int", desc="完成时间")
     * @Apidoc\Returned("data.detail.sale_orders[].order_amount", type="float", desc="订单金额")
     * @Apidoc\Returned("data.detail.sale_orders[].payment_amount", type="float", desc="支付金额")
     * @Apidoc\Returned("data.detail.sale_orders[].refund_amount", type="float", desc="退款金额")
     * @Apidoc\Returned("data.detail.sale_orders[].pay_type_name", type="string", desc="支付方式名称")
     * @Apidoc\Returned("data.detail.sale_orders[].member_name", type="string", desc="会员名称")
     * @Apidoc\Returned("data.detail.sale_orders[].member_uuid", type="string", desc="会员UUID")
     * @Apidoc\Returned("data.detail.sale_orders[].is_free", type="bool", desc="是否免单")
     * @Apidoc\Returned("data.detail.sale_orders[].free_reason", type="object", desc="免单原因（多语言）")
     * @Apidoc\Returned("data.detail.sale_orders[].products", type="array", desc="商品列表")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].uuid", type="string", desc="商品UUID")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].locale_name", type="object", desc="商品名称（多语言）")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].locale_attribute_name", type="object", desc="商品规格名称（多语言）")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].price", type="float", desc="商品单价")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].num", type="int", desc="商品数量")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].sale_price", type="float", desc="销售价格")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].total_price", type="float", desc="总价")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].refund_amount", type="float", desc="退款金额")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].status", type="int", desc="商品状态")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].remark", type="string", desc="备注")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].is_gift", type="bool", desc="是否赠品")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].is_buffet", type="bool", desc="是否自助餐")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].is_buffet_customer", type="bool", desc="是否自助餐客户")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].is_delay", type="bool", desc="是否延迟")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].is_must", type="bool", desc="是否必选")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].gift_reason", type="string", desc="赠送原因")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].image_url", type="string", desc="商品图片")
     * @Apidoc\Returned("data.detail.sale_orders[].products[].refund_reason", type="string", desc="退款原因")
     * @Apidoc\Returned("data.operation_log", type="object", desc="操作日志")
     * @Apidoc\Returned("data.operation_log.list", type="array", desc="操作日志列表")
     * @Apidoc\Returned("data.operation_log.list[].uuid", type="string", desc="日志UUID")
     * @Apidoc\Returned("data.operation_log.list[].user_name", type="string", desc="操作人名称")
     * @Apidoc\Returned("data.operation_log.list[].user_email", type="string", desc="操作人邮箱")
     * @Apidoc\Returned("data.operation_log.list[].source", type="string", desc="来源")
     * @Apidoc\Returned("data.operation_log.list[].create_time", type="string", desc="创建时间")
     * @Apidoc\Returned("data.operation_log.list[].description", type="string", desc="描述")
     * @Apidoc\Returned("data.operation_log.list[].pay_type", type="array", desc="支付方式")
     * @Apidoc\Returned("data.operation_log.list[].refund_type", type="int", desc="退款类型") 
     */
    public function detail($sale_bill_uuid, $sale_order_uuid = 0)
    {
        // $res = HttpHelp::getRequest('http://nginx/api/v1/shop/delivery_order/info', [
        //     'sale_bill_uuid' => $sale_bill_uuid,
        //     'sale_order_uuid' => $sale_order_uuid,
        // ], [
        //     'Authorization: Bearer ' . request()->header('token'),
        //     'Accept-Language: ' . request()->header('language'),
        // ]);
        // if (!$res) {
        //     return $this->renderError('请求失败');
        // } 

        // $result = json_decode($res, true);
        // if (($result['code'] ?? -1) != 0) {
        //     return $this->renderError($result['message'] ?? '请求失败');
        // }
        // //
        // $data = $result['data'];
        // $data['detail']['create_time'] =  $data['detail']['create_time'] ? date('Y-m-d H:i:s',  $data['detail']['create_time']) : '';
        // $data['detail']['finish_time'] =  $data['detail']['finish_time'] ? date('Y-m-d H:i:s',  $data['detail']['finish_time']) : '';

        // foreach ($data['operation_log']['list'] as &$item) {
        //     $item['create_time'] = $item['create_time'] ? date('Y-m-d H:i:s', $item['create_time']) : '';
        // }

        // 
        return $this->renderSuccess('', []);
    }


    /**
     * 构建列表查询参数
     */
    private function buildListQueryParams($data)
    {
        // 搜索日期类型: '0'-全部 '1'-今天 '2'-昨天 '3'-本周
        $dateType = intval($data['time_type']) - 1;
        // 搜索外卖序号
        $sn = $data['serial_no'];
        // 搜索订单号
        $orderNo = $data['order_no'];
        // 搜索下单时间、完成时间
        $enableCreateTime = false;
        $enablePayTime = false;
        $queryStartTime = 0;
        $queryEndTime = 0;
        $time = $data['time'] ?: [];
        if (!empty($time)) {
            $queryStartTime = strtotime($time[0]);
            $queryEndTime = strtotime($time[1]) + 86399;
        }
        $timeMode = $data['time_mode'] ?: [];
        if (in_array(0, $timeMode)) {
            $enableCreateTime = true;
        }
        if (in_array(1, $timeMode)) {
            $enablePayTime = true;
        }
        // 搜索订单状态
        $status = [
            'all' => -1,
            'pending_payment' => 0,
            'awaiting_delivery' => 1,
            'delivering' => 2,
            'cancelled' => 2,
            'completed' => 2,
        ][$data['dataType']];

        return [
            'sn' => $sn,
            'order_no' => $orderNo,
            'date_type' => $dateType,
            'status' => $status,
            'page_no' => $data['page'] ?? 1,
            'page_size' => $data['list_rows'] ?? 10,
            'enable_create_time' => $enableCreateTime,
            'enable_pay_time' => $enablePayTime,
            'query_start_time' => $queryStartTime,
            'query_end_time' => $queryEndTime,
        ];
    }
}
