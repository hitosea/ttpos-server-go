<?php

namespace app\scan\controller\base;

use app\common\model\order\Order;
use hg\apidoc\annotation as Apidoc;
use app\scan\controller\Controller;
use app\common\model\store\Table;
use app\common\model\supplier\Supplier as SupplierModel;


/**
 * 基础
 */
class Base extends Controller
{
    /**
     * @Apidoc\Title("基础信息")
     * @Apidoc\Desc("基础信息")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/scan/base.base/getInfo")
     * @Apidoc\Returned("currency", type="object", desc="货币", children={
     *    @Apidoc\Param ("unit",type="string",desc="主单位"),
     *    @Apidoc\Param ("unit_position",type="string",desc="v1.0.3主货币显示位置 0-金额前 1-金额后"),
     *    @Apidoc\Param ("is_open",type="string",desc="是否开启副单位 0-关闭 1-开启"),
     *    @Apidoc\Param ("vices",type="object",desc="副单位",children={
     *       @Apidoc\Param ("vice_unit",type="string",desc="副单位"),
     *       @Apidoc\Param ("vice_unit_position",type="string",desc="v1.0.3副货币显示位置 0-金额前 1-金额后"),
     *       @Apidoc\Param ("unit_rate",type="float",desc="副单位比例"),
     *   }),
     * })
     * @Apidoc\Returned("h5", type="object", desc="扫码H5设置", children={
     *    @Apidoc\Param ("is_call_service",type="string",desc="是否开启呼叫服务员 0-关闭 1-开启"),
     *    @Apidoc\Param ("is_customer_order",type="string",desc="是否开启顾客自助下单 0-关闭 1-开启"),
     *    @Apidoc\Param ("is_voice_remind",type="string",desc="是否开启声音提醒 0-关闭 1-开启（v1.0.5）"),
     *    @Apidoc\Param ("is_show_sold_out",type="string",desc="是否显示售罄商品 0-关闭 1-开启"),
     *    @Apidoc\Param("is_buffet_order_limit", type="string", require=true, desc="v1.0.4是否开启自助餐下单限制 0-关闭 1-开启"),
     *    @Apidoc\Param("buffet_order_limit", type="object", require=true, desc="v1.0.4自助餐下单限制", children={
     *       @Apidoc\Param("is_limit_time", type="int", require=true, desc="是否开启时间限制 0-关闭 1-开启"),
     *       @Apidoc\Param("limit_time", type="int", require=true, desc="时间限制（分钟）"),
     *       @Apidoc\Param("is_limit_num", type="int", require=true, desc="是否开启数量限制 0-关闭 1-开启"),
     *       @Apidoc\Param("limit_num", type="int", require=true, desc="数量限制"),
     *   }),
     *    @Apidoc\Param("is_order_limit", type="string", require=true, desc="v1.0.4是否开启非自助餐下单限制 0-关闭 1-开启"),
     *    @Apidoc\Param("order_limit", type="object", require=true, desc="v1.0.4自助餐下单限制", children={
     *       @Apidoc\Param("is_limit_time", type="int", require=true, desc="是否开启时间限制 0-关闭 1-开启"),
     *       @Apidoc\Param("limit_time", type="int", require=true, desc="时间限制（分钟）"),
     *       @Apidoc\Param("is_limit_num", type="int", require=true, desc="是否开启数量限制 0-关闭 1-开启"),
     *       @Apidoc\Param("limit_num", type="int", require=true, desc="数量限制"),
     *   }),
     *    @Apidoc\Param ("language",type="array",desc="常用语言，默认th, en, zh, zhtw"),
     *    @Apidoc\Param ("default_language",type="array",desc="默认语言，默认en"),
     * })
     * @Apidoc\Returned("buffet", type="object", desc="自助餐设置", children={
     *    @Apidoc\Param ("is_open",type="string",desc="是否开启自助餐 0-关闭 1-开启"),
     *    @Apidoc\Param ("tablet_end_time",type="string",desc="平板结束时间提醒（分）"),
     *    @Apidoc\Param("is_remain_continue", type="int", require=true, desc="剩余xx分不可继续点餐开关 0-关闭 1-开启（v1.0.4弃用）"),
     *    @Apidoc\Param("remain_continue_time", type="string", require=true, default="5", desc="剩余xx分不可继续点餐（v1.0.4弃用）"),
     *    @Apidoc\Param("remain_continue_notice_time", type="string", require=true, default="5", desc="剩余xx分提醒不可继续点餐（v1.0.4弃用）"),
     *    @Apidoc\Param ("is_buy_continue",type="string",desc="非自助餐商品到时是否能继续选购 0-关闭 1-开启"),
     *    @Apidoc\Param ("is_add_clock",type="string",desc="是否开启加钟 0-关闭 1-开启"),
     * })
     * @Apidoc\Returned("cashier", type="object", desc="收银机设置", children={
     *    @Apidoc\Param("order_method", type="array", require=true, desc="v1.0.3用餐方式 收银-is_cashier_order（0-关闭 1-开启） 桌台-is_table_order（0-关闭 1-开启）", children={
     *      @Apidoc\Param("is_cashier_order", type="int", require=true, desc="收银用餐"),
     *      @Apidoc\Param("is_table_order", type="int", require=true, desc="桌台用餐"),
     *    })
     * })
     * @Apidoc\Returned("cloud_basic", type="object", desc="v1.0.2云端基础信息", children={
     * @Apidoc\Param("base", type="array", require=true, desc="云端基础设置", children={
     *     @Apidoc\Param("brand_name", type="string", require=true, desc="品牌名称"),
     *     @Apidoc\Param("brand_logo", type="string", require=true, desc="正方形品牌logo"),
     *     @Apidoc\Param("brand_logo_long", type="string", require=true, desc="长方形品牌logo"),
     *     @Apidoc\Param("browser_logo", type="string", require=true, desc="v1.0.4.1浏览器logo"),
     *     @Apidoc\Param("browser_title", type="string", require=true, desc="v1.0.4.1浏览器标题"),
     *     @Apidoc\Param("expiration_reminder", type="string", require=true, desc="剩余多少天到期提示"),
     *  }),
     * @Apidoc\Param ("shop",type="object",desc="云端门店信息设置",children={
     *      @Apidoc\Param ("name",type="string",desc="店铺名称"),
     *      @Apidoc\Property("sale_stock",type="int", desc="进销存"),
     *      @Apidoc\Property("reserve",type="int", desc="预订"),
     *      @Apidoc\Property("cash_limit",type="int", desc="收银机上限"),
     *      @Apidoc\Property("kitchen_limit",type="int", desc="厨显上限"),
     *      @Apidoc\Property("tablet_limit",type="int", desc="平板上限")
     *  }),
     * })
     * @Apidoc\Param ("kitchen",type="object",desc="v1.0.4厨显设置(暂时不用做)",children={
     *      @Apidoc\Param ("is_open",type="int",desc="是否开启厨显 0-关闭 1-开启"),
     *  }),
     * })
     * @Apidoc\Returned("is_cloud_deploy", type="bool", desc="是否是云端部署 true-云端部署 false-本地部署")
     */
    public function getInfo()
    {
        $table_id = $this->table['table_id'] ?? 0;
        $table = Table::detail($table_id);
        if (!$table) {
            return $this->renderError('桌台不存在');
        }
        $detail = Order::getTableUnderwayOrder($table_id);
        $orderInfo['is_buffet'] = $detail ? $detail->is_buffet : 0;
        $info = SupplierModel::getScanBaseInfo();
        $reData = [
            'tableInfo' => $table,
            'baseInfo' => $info,
            'orderInfo' => $orderInfo,
        ];
        return $this->renderSuccess('基础信息', $reData);
    }
}
