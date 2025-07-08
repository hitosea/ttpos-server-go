<?php

namespace app\shop\controller\user;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\shop\UserShiftLog as UserShiftLogModel;
use app\common\model\shop\UserShiftSnapshot;
use app\shop\service\statistics\UserShiftLogService as UserShiftLogService;

/**
 * 收银交班
 * @Apidoc\Group("statistics")
 * @Apidoc\Sort(7)
 */
class UserShiftLog extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/user.UserShiftLog/index")
     * @Apidoc\Param("user_id", type="string", require=true, default="", desc="收银员id")
     * @Apidoc\Param("date", type="array", require=true, default="", desc="起始时间")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\common\model\shop\UserShiftLog\getList", desc="列表",children={
     *     @Apidoc\Param ("total_money",type="float",desc="营业收入"),
     * })
     * @Apidoc\Returned("cashierList", type="array", ref="app\shop\model\auth\User\getList", desc="收银员列表")
     */
    public function index()
    {
        $data = $this->postData();
        $model = new UserShiftLogModel;
        $list = $model->getList($data);
        $cashierList = $model->getExistUserList();
        return $this->renderSuccess('', compact('list', 'cashierList'));
    }

    /**
     * @Apidoc\Title("详情")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/user.UserShiftLog/detail")
     * @Apidoc\Param("id", type="int", require=true, default=0, desc="id")
     * @Apidoc\Returned("detail", type="array", ref="app\common\model\shop\UserShiftLog\detail", desc="详情", children={
     *  @Apidoc\Returned("withdraw_cash", type="float", desc="取出现金（v1.0.6）"),
     *  @Apidoc\Returned("deposit_cash", type="float", desc="存入现金（v1.0.6）"),
     *  @Apidoc\Returned("exception_remark", type="string", desc="异常报备（v1.0.6）"),
     *  @Apidoc\Returned("order", type="array", desc="订单概况", children={
     *    @Apidoc\Param("total_price",type="int",default=0,desc="营业总额"),
     *    @Apidoc\Param("receivable_price",type="int",default=0,desc="应收"),
     *    @Apidoc\Param("service_money",type="int",default=0,desc="服务费"),
     *    @Apidoc\Param("sales_price",type="int",default=0,desc="销售额"),
     *    @Apidoc\Param("consumption_tax_money",type="int",default=0,desc="消费税"),
     *    @Apidoc\Param("product_num",type="int",default=0,desc="商品数量"),
     *    @Apidoc\Param("discount_money",type="int",default=0,desc="优惠折扣"),
     *    @Apidoc\Param("user_discount_money",type="int",default=0,desc="会员折扣"),
     *    @Apidoc\Param("pay_fee_money",type="int",default=0,desc="支付手续费（v1.0.6）"),
     *    @Apidoc\Param("refund_money",type="int",default=0,desc="退款"),
     *    @Apidoc\Param("received_price",type="int",default=0,desc="实收"),
     *    @Apidoc\Param("total_order_num",type="int",default=0,desc="所有订单数"),
     *    @Apidoc\Param("total_table_num",type="int",default=0,desc="总桌数"),
     *    @Apidoc\Param("total_people_num",type="int",default=0,desc="总人数"),
     *    @Apidoc\Param("min_order_price",type="int",default=0,desc="最小订单金额"),
     *    @Apidoc\Param("max_order_price",type="int",default=0,desc="最大订单金额"),
     *    @Apidoc\Param("avg_order_price",type="int",default=0,desc="平均订单金额"),
     *    @Apidoc\Param("table_order_num",type="int",default=0,desc="桌台-订单数（桌数）"),
     *    @Apidoc\Param("table_people_num",type="int",default=0,desc="桌台-人数"),
     *    @Apidoc\Param("table_min_order_price",type="int",default=0,desc="桌台-最小订单金额"),
     *    @Apidoc\Param("table_max_order_price",type="int",default=0,desc="桌台-最大订单金额"),
     *    @Apidoc\Param("table_avg_order_price",type="int",default=0,desc="桌台-平均订单金额"),
     *    @Apidoc\Param("table_people_avg",type="int",default=0,desc="桌台-人均"),
     *    @Apidoc\Param("cashier_order_num",type="int",default=0,desc="收银-订单数"),
     *    @Apidoc\Param("cashier_people_num",type="int",default=0,desc="收银-人数"),
     *    @Apidoc\Param("cashier_min_order_price",type="int",default=0,desc="收银-最小订单金额"),
     *    @Apidoc\Param("cashier_max_order_price",type="int",default=0,desc="收银-最大订单金额"),
     *    @Apidoc\Param("cashier_avg_order_price",type="int",default=0,desc="收银-平均订单金额"),
     *    @Apidoc\Param("percentage_list",type="array",desc="税收百分比对象列表",children={
     *        @Apidoc\Param("tax_rate", type="int", default=10, desc="税率"),
     *        @Apidoc\Param("consumption_tax", type="int", default=0, desc="消费税"),
     *        @Apidoc\Param("total_price", type="int", default=0, desc="总价格")
     *    }),
     *    @Apidoc\Param("peak_hour_list",type="array",desc="高峰时间段",children={
     *        @Apidoc\Param("time_period", type="string", default="06/26 01:00-02:00", desc="时间段"),
     *        @Apidoc\Param("num", type="int", default=1, desc="订单数"),
     *        @Apidoc\Param("amount", type="int", default=0, desc="订单价格")
     *    }),
     *    @Apidoc\Returned("incomes",type="array",desc="支付方式的收入列表" ,children={
     *        @Apidoc\Param("pay_type",type="string",desc="付款类型"),
     *        @Apidoc\Param("pay_type_name",type="string",desc="付款类型名称"),
     *        @Apidoc\Param("price",type="string",desc="付款价格"),
     *        @Apidoc\Param("order_num",type="int",desc="订单数"),
     *    }),
     *    @Apidoc\Returned("user_count",type="int",desc="有效会员数"),
     * }),
     * })
     * @Apidoc\Returned("salesInfo",type="array",desc="销售信息",children={
     *     @Apidoc\Param ("name",type="string",desc="分类名称"),
     *     @Apidoc\Param ("sales",type="string",desc="销售数量"),
     *     @Apidoc\Param ("prices",type="string",desc="销售金额"),
     * })
     * @Apidoc\Returned("abnormal",type="array",desc="异常信息",children={
     *     @Apidoc\Param ("refund_product_times",type="string",desc="退菜次数"),
     *     @Apidoc\Param ("refund_times",type="string",desc="退款次数"),
     *     @Apidoc\Param ("reverse_settle_times",type="string",desc="反结账次数"),
     *     @Apidoc\Param ("product_free_times",type="string",desc="赠菜次数"),
     *     @Apidoc\Param ("free_order_times",type="string",desc="免单次数"),
     *     @Apidoc\Param ("product_move_times",type="string",desc="转菜次数"),
     *     @Apidoc\Param ("change_price_times",type="string",desc="单品改价次数"),
     *     @Apidoc\Param ("change_order_price_times",type="string",desc="整单改价次数"),
     *     @Apidoc\Param ("discount_order_times",type="string",desc="整单折扣次数"),
     *     @Apidoc\Param ("round_order_times",type="string",desc="整单抹零次数"),
     *
     * })
     */
    public function detail()
    {
        $data = $this->postData();
        $model = new UserShiftLogModel;
        $info = $model->detail($data['id']);
        if (!$info) {
            return $this->renderError('找不到数据');
        }
        $snapshotModel = new UserShiftSnapshot();
        $detail = $snapshotModel->getSnapshot($info);
        if ($detail === false) {
            return $this->renderError($snapshotModel->getError());
        }
        //
        $detail['total_income'] = $snapshotModel->calcBusinessIncome($detail);
        // 
        $salesInfo = $detail['salesInfo'] ?? [];
        // 
        if (isset($detail['abnormal'])) {
            $detail['abnormal']['refund_times'] = $detail['abnormal']['refund_time'] ?? 0;
        }
        // 
        return $this->renderSuccess('', compact('detail', 'salesInfo'));
    }

    /**
     * @Apidoc\Title("导出")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/user.UserShiftLog/export")
     * @Apidoc\Param("user_name", type="string", require=true, default="", desc="收银员名称")
     * @Apidoc\Param("start_time", type="string", require=true, default="", desc="开始时间")
     * @Apidoc\Param("end_time", type="string", require=true, default="", desc="结束时间")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned()
     */
    public function export()
    {
        $data = $this->postData();
        $data['list_rows'] = 1000;
        $model = new UserShiftLogModel;
        $list = $model->getList($data);
        (new UserShiftLogService())->userShiftLogListExport($list);
        return $this->renderSuccess('');
    }
}
