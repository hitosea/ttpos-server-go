<?php

namespace app\shop\controller\store;

use help\HttpHelp;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
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
    public function index()
    {   
        $data = $this->postData();
        //
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/order/list', $this->buildListQueryParams($data), [
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
        
        foreach ($result['data']['list'] as $key => &$item) {
            $item['finish_time'] = $item['finish_time'] ? date('Y-m-d H:i:s', $item['finish_time']) : '';
            if ($item['sale_orders']) {
                foreach ($item['sale_orders'] as $subKey => &$subItem) {
                    $subItem['finish_time'] = $subItem['finish_time'] ? date('Y-m-d H:i:s', $subItem['finish_time']) : '';
                }
            }
        }
        
        $result['data']['ex_style']  = DeliveryTypeEnum::store();
        
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
        if (($result['code'] ?? -1) != 0) {
            return $this->renderError($result['message'] ?? '请求失败');
        }
        //
        $data = $result['data'];
        $data['detail']['create_time'] =  $data['detail']['create_time'] ? date('Y-m-d H:i:s',  $data['detail']['create_time']) : '';
        $data['detail']['finish_time'] =  $data['detail']['finish_time'] ? date('Y-m-d H:i:s',  $data['detail']['finish_time']) : '';

        foreach ($data['operation_log']['list'] as &$item) {
            $item['create_time'] = $item['create_time'] ? date('Y-m-d H:i:s', $item['create_time']) : '';
        }

        // 
        return $this->renderSuccess('', $data);
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
}
