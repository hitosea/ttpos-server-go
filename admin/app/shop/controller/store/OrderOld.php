<?php

namespace app\shop\controller\store;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\shop\model_old\order\Order as OldOrderModel;
use app\common\model\app\App as AppModel;

/**
 * 订单管理-旧
 * @Apidoc\Group("order")
 * @Apidoc\Sort(4)
 */
class OrderOld extends Controller
{
    /**
     * @Apidoc\Title("订单列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.orderOld/index")
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
        $app = AppModel::where('uuid', request()->appId)->find();
        if (!$app->old_company_id) {
            return $this->renderError('未存在迁移库');
        }
        request()->appId = $app->old_company_id;
        $data['order_type'] = 1;
        $data['parent_id'] = 0;
        return $this->renderSuccess('', OldOrderModel::getLists($data));
    }

    /**
     * @Apidoc\Title("订单详情")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.orderOld/detail")
     * @Apidoc\Param("sale_bill_uuid", type="int", require=true, default="", desc="销售账单UUID")
     * @Apidoc\Param("sale_order_uuid", type="int", require=true, default="", desc="销售订单UUID 当查看子订单信息的时候才需要传")
     * @Apidoc\Returned("detail", type="array", ref="app\shop\model\order\Order\detail", desc="订单详情")
     */
    public function detail($sale_bill_uuid)
    {
        $app = AppModel::where('uuid', request()->appId)->find();
        if (!$app->old_company_id) {
            return $this->renderError('未存在迁移库');
        }
        request()->appId = $app->old_company_id;
        return $this->renderSuccess('', OldOrderModel::details($sale_bill_uuid));
    }

   
}
