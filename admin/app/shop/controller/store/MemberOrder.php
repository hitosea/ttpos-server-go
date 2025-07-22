<?php

namespace app\shop\controller\store;

use help\HttpHelp;
use app\shop\controller\Controller;
use app\shop\service\order\MemberOrderExportService;
use hg\apidoc\annotation as Apidoc;


/**
 * 外送订单管理
 * @Apidoc\Group("order")
 * @Apidoc\Sort(4)
 */
class MemberOrder extends Controller
{
    /**
     * @Apidoc\Title("外送订单列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.MemberOrder/index")
     * @Apidoc\Param("date_range", type="int", require=false, default="", desc="日期类型 -1=全都、 0=今天、 1=昨天、 2=本周")
     * @Apidoc\Param("serial_no", type="string", require=false, default="", desc="订单序号")
     * @Apidoc\Param("order_no", type="string", require=false, default="", desc="订单编号")
     * @Apidoc\Param("status", type="string", require=false, default="all", desc="状态: unaccept-待接单, accept-备餐中, undelivery-待配送, delivery-配送中, completed-已完成, cancel-已取消")
     * @Apidoc\Param("time_type", type="int",require=true, default="0", desc="时间类型  1=下单时间、 2=支付时间")
     * @Apidoc\Param("query_start_time", type="int", require=false, default="", desc="查询开始时间戳")
     * @Apidoc\Param("query_end_time", type="int", require=false, default="", desc="查询结束时间戳")
     * @Apidoc\Param("page_no", type="int", require=false, default="", desc="页码")
     * @Apidoc\Param("page_size", type="int", require=false, default="", desc="每页条数") 
     */
    public function index()
    {
        $data = $this->postData(); 
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/member_order/list', $data, [
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

        return $this->renderSuccess('', $result['data']);
    }

    /**
     * @Apidoc\Title("外送订单详情")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.MemberOrder/detail")
     * @Apidoc\Param("member_sale_order_uuid", type="int", require=true, default="", desc="订单UUID") 
     */
    public function detail($member_sale_order_uuid)
    {
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/member_order/info', [
            'member_sale_order_uuid' => $member_sale_order_uuid,
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
         
        return $this->renderSuccess('', $data);
    }

    /**
     * @Apidoc\Title("拒单外送订单")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.MemberOrder/reject")
     * @Apidoc\Param("member_sale_order_uuid", type="int", require=true, default="", desc="订单UUID") 
     */
    public function reject($member_sale_order_uuid)
    {
        $res = HttpHelp::postRequest('http://nginx/api/v1/shop/member_order/reject', json_encode($this->postData()), [
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
        
        return $this->renderSuccess('');
    }

    /**
     * @Apidoc\Title("取消外送订单")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.MemberOrder/cancel")
     * @Apidoc\Param("member_sale_order_uuid", type="int", require=true, default="", desc="订单UUID")
     * @Apidoc\Param("cancel_reason", type="string", require=true, default="", desc="取消原因")
     * @Apidoc\Param("bank_code", type="string", require=false, default="", desc="银行代码 - 暂时不用")
     * @Apidoc\Param("account_no", type="string", require=false, default="", desc="账号 - 暂时不用")
     * @Apidoc\Param("account_name", type="string", require=false, default="", desc="账户名称- 暂时不用")
     */
    public function cancel()
    {
        $res = HttpHelp::postRequest('http://nginx/api/v1/shop/member_order/cancel', json_encode($this->postData()), [
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

        return $this->renderSuccess('');
    }


    /**
     * @Apidoc\Title("外送订单列表导出")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.MemberOrder/export")
     * @Apidoc\Param("date_range", type="int", require=false, default="", desc="日期类型 -1=全都、 0=今天、 1=昨天、 2=本周")
     * @Apidoc\Param("serial_no", type="string", require=false, default="", desc="订单序号")
     * @Apidoc\Param("order_no", type="string", require=false, default="", desc="订单编号")
     * @Apidoc\Param("status", type="string", require=false, default="all", desc="状态: unaccept-待接单, accept-备餐中, undelivery-待配送, delivery-配送中, completed-已完成, cancel-已取消")
     * @Apidoc\Param("time_type", type="int",require=true, default="0", desc="时间类型  1=下单时间、 2=支付时间")
     * @Apidoc\Param("query_start_time", type="int", require=false, default="", desc="查询开始时间戳")
     * @Apidoc\Param("query_end_time", type="int", require=false, default="", desc="查询结束时间戳")
     */
    public function export()
    {
        $data = $this->postData();
        $data['page_no'] = 1;
        $data['page_size'] = 1000;
        // 请求获取充值订单列表接口
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/member_order/list', $data, [
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

        $list = $result['data']['list'];

        // 导出excel文件
        return (new MemberOrderExportService)->export($list);
    }

    /**
     * @Apidoc\Title("外送订单退款弹窗信息")
     * @Apidoc\Method ("GET")
     * @Apidoc\Url ("/index.php/shop/store.MemberOrder/return_info")
     * @Apidoc\Param("member_sale_order_uuid", type="int", require=true, default="", desc="订单UUID")
     */
    public function return_info()
    {
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/member_order/return_info', $this->getData(), [
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
        return $this->renderSuccess('', $result['data']);
    }

    /**
     * @Apidoc\Title("外送订单退款")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.MemberOrder/return")
     */
    public function return()
    {
        $res = HttpHelp::postRequest('http://nginx/api/v1/shop/member_order/return', json_encode($this->postData()), [
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
        return $this->renderSuccess('');
    }

}