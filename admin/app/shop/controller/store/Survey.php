<?php

namespace app\shop\controller\store;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\shop\model\order\Order as OrderModel;
use app\shop\service\statistics\SurveyService;

/**
 * 店内概况
 * @Apidoc\Group("statistics")
 * @Apidoc\Sort(7)
 */
class Survey extends Controller
{
    /**
     * @Apidoc\Title("店内概况")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.survey/index")
     * @Apidoc\Param("date", type="array", require=false, default="", desc="起始日期")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("detail", type="array", desc="订单概况", children={
     *    @Apidoc\Param("total_price",type="int",default=0,desc="营业总额"),
     *    @Apidoc\Param("receivable_price",type="int",default=0,desc="应收"),
     *    @Apidoc\Param("service_money",type="int",default=0,desc="服务费"),
     *    @Apidoc\Param("sales_price",type="int",default=0,desc="销售额"),
     *    @Apidoc\Param("consumption_tax_money",type="int",default=0,desc="消费税"),
     *    @Apidoc\Param("product_num",type="int",default=0,desc="商品数量"),
     *    @Apidoc\Param("discount_money",type="int",default=0,desc="优惠折扣"),
     *    @Apidoc\Param("discount_ratio",type="int",default=0,desc="优惠占比 （v1.0.7）"),
     *    @Apidoc\Param("user_discount_money",type="int",default=0,desc="会员折扣"),
     *    @Apidoc\Param("pay_fee_money",type="int",default=0,desc="支付手续费（v1.0.6）"),
     *    @Apidoc\Param("refund_money",type="int",default=0,desc="退款"),
     *    @Apidoc\Param("received_price",type="int",default=0,desc="实收"),
     *    @Apidoc\Param("business_price",type="int",default=0,desc="营业收入（v1.0.7） = 实收金额-支付手续费-税费"),
     *    @Apidoc\Param("free_product_price",type="int",default=0,desc="赠菜折扣：赠菜的总金额（v1.0.7）"),
     *    @Apidoc\Param("free_product_num",type="int",default=0,desc="赠菜数量：赠菜的数量（v1.0.7）"),
     *    @Apidoc\Param("free_order_price",type="int",default=0,desc="免单折扣 = 免单的总金额（v1.0.7）"),
     *    @Apidoc\Param("free_order_num",type="int",default=0,desc="免单数量（v1.0.7）"),
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
     *        @Apidoc\Param("amount", type="int", default=0, desc="订单价格"),
     *    }),
     *    @Apidoc\Returned("incomes",type="array",desc="支付方式的收入列表" ,children={
     *        @Apidoc\Param("pay_type",type="string",desc="付款类型"),
     *        @Apidoc\Param("pay_type_name",type="string",desc="付款类型名称"),
     *        @Apidoc\Param("price",type="string",desc="付款价格"),
     *        @Apidoc\Param("order_num",type="int",desc="订单数")
     *    }),
     *    @Apidoc\Returned("user_count",type="int",desc="有效会员数"),
     * })
     * * @Apidoc\Returned("regionData", type="array", desc="区域数据（v1.0.7）", children={
     *    @Apidoc\Returned("area_name", type="string", desc="区域名称"),
     *    @Apidoc\Returned("product_num", type="string", desc="商品数量"),
     *    @Apidoc\Returned("business_price", type="int", desc="营业收入"),
     *    @Apidoc\Returned("sales_price", type="float", desc="总销售额"),
     * })
     * @Apidoc\Returned("salesNumRank", type="array", desc="销量排行", children={
     *    @Apidoc\Returned("product_name", type="string", desc="商品名称（多语言）"),
     *    @Apidoc\Returned("product_name_text", type="string", desc="商品名称"),
     *    @Apidoc\Returned("total_num", type="int", desc="销量"),
     *    @Apidoc\Returned("total_price", type="float", desc="销售额"),
     * })
     * @Apidoc\Returned("salesMoneyRank", type="array", desc="销售额排行", children={
     *    @Apidoc\Returned("product_name_text", type="string", desc="商品名称"),
     *    @Apidoc\Returned("product_name", type="string", desc="商品名称（多语言）"),
     *    @Apidoc\Returned("total_num", type="int", desc="销量"),
     *    @Apidoc\Returned("total_price", type="float", desc="销售额"),
     * })
     */
    public function index()
    {
        $data = $this->postData() ?: [];
        $shopSupplierId = $this->store['user']['shop_supplier_id'];
        $model = new OrderModel();

        // 店內概況
        $detail = [];
        // 区域数据
        $regionData =[];
        // 销量排行
        $salesNumRank = [];
        // 销售额排行
        $salesMoneyRank = [];

        // todo 兼容
        // // 店內概況
        // $detail = $model->storeOverview($data);
        // // 区域数据
        // $regionData = $model->regionData($data);
        // // 销量排行
        // $salesNumRank = $model->getProductRank(0, 1, $shopSupplierId, $data);
        // // 销售额排行
        // $salesMoneyRank = $model->getProductRank(1, 1, $shopSupplierId, $data);
        //
        return $this->renderSuccess('', compact('detail', 'regionData', 'salesNumRank', 'salesMoneyRank'));
    }

    /**
     *
     * @Apidoc\Title("导出")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.survey/export")
     * @Apidoc\Param("date", type="array", require=false, default="", desc="起始日期")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned()
     */
    public function export()
    {
        $data = $this->postData();
        if (empty($data['date'])) {
            return $this->renderError('请选择时间');
        }
        if (strtotime($data['date'][1]) - strtotime($data['date'][0]) > 30 * 86400) {
            return $this->renderError('时间范围不能超过31天');
        }
        $model = new OrderModel;
        $list = $model->storeOverviewByDate($data);
        if (empty($list)) {
            return $this->renderError('导出数据为空');
        }
        (new SurveyService())->surveyExport($list);
        return $this->renderSuccess('');
    }
}
