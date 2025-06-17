<?php

namespace app\shop\controller\user;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\enum\settings\SettingEnum;
use app\shop\model\settings\Setting as SettingModel;
use app\shop\model\user\PointsLog as PointsLogModel;

/**
 * 积分设置
 * @Apidoc\Group("user")
 * @Apidoc\Sort(5)
 */
class Points extends Controller
{

    /**
     * @Apidoc\Title("积分设置")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/user.points/setting")
     * @Apidoc\Param("deduction_order", type="string", require=true, default="", desc="扣款顺序状态")
     * @Apidoc\Param("deduct_order", type="string", require=true, default="", desc="扣款顺序 1-先主账户后赠送账户 2-先赠送账户后主账户 3-按比例（v1.0.8）")
     * @Apidoc\Param("deduct_ratio_main", type="string", require=true, default="", desc="主账户扣款比例0-100（v1.0.8）")
     * @Apidoc\Param("deduct_ratio_gift", type="string", require=true, default="", desc="赠送账户扣款比例0-100（v1.0.8）")
     * @Apidoc\Param("points_name", type="string", require=true, default="", desc="积分名称")
     * @Apidoc\Param("is_shopping_gift", type="int", require=true, default="", desc="是否开启购物送积分 0-关闭 1-开启")
     * @Apidoc\Param("gift_ratio", type="int", require=true, default="", desc="积分赠送比例")
     * @Apidoc\Param("is_shopping_discount", type="int", require=true, default="", desc="是否允许下单使用积分抵扣 0-关闭 1-开启")
     * @Apidoc\Param("discount", type="array", require=true, default="", desc="充值参数", children={
     *    @Apidoc\Param("discount_ratio", type="float", require=true, default="", desc="积分抵扣比例"),
     *    @Apidoc\Param("max_money_ratio", type="float", require=true, default="", desc="最高可抵扣订单额百分比"),
     *    @Apidoc\Param("full_order_price", type="float", require=true, default="", desc="订单满[?]元"),
     * })
     * @Apidoc\Param("describe", type="string", require=true, default="", desc="积分说明")
     * @Apidoc\Param("shopping_gift_rules", type="array", desc="赠送规则", children={
     *   @Apidoc\Param("type", type="string", desc="赠送类型: payment_amount-按付款金额比例赠送, desk-按桌台人数赠送"),
     *   @Apidoc\Param("is_open", type="string", desc="是否开启: 1-开启, 0-关闭"),
     *   @Apidoc\Param("is_member_level_related", type="string", desc="是否按会员等级赠送: 1-是, 0-否"),
     *   @Apidoc\Param("value", type="string", desc="积分比例/数量"),
     *   @Apidoc\Param("payment_amount_requirement", type="string", desc="付款金额要求"),
     *   @Apidoc\Param("meal_type", type="array", desc="就餐类型: buffet-自助餐, non-buffet-非自助餐"),
     *   @Apidoc\Param("balance_payment_get_points", type="string", desc="会员余额支付是否赠送: 1-是, 0-否"),
     *   @Apidoc\Param("refund_return_points", type="string", desc="退款自动扣积分: 1-是, 0-否"),
     *   @Apidoc\Param("member_levels", type="array", desc="会员等级")
     * }),
     * @Apidoc\Param("exchange", type="array", desc="积分兑换设置", children={
     *   @Apidoc\Param("open_points_exchange", type="string", desc="会员积分抵扣订单金额: 1-开启, 0-关闭"),
     *   @Apidoc\Param("points_exchange_rate", type="string", desc="每积分抵扣应付金额"),
     *   @Apidoc\Param("auto_points_exchange", type="string", desc="是否自动抵扣: 1-是, 0-否")
     * })
     * @Apidoc\Returned()
     */
    public function setting()
    {
        if ($this->request->isGet()) {
            $values = SettingModel::getSupplierItem(SettingEnum::POINTS, $this->store['user']['shop_supplier_id']);
            return $this->renderSuccess('', compact('values'));
        }
        // 判断比例相加等于100
        $data = $this->postData();
        if ($data['deduct_order'] == 3) {
            if ($data['deduct_ratio_main'] + $data['deduct_ratio_gift'] != 100) {
                return $this->renderError('扣款比例相加必须等于100');
            }
        }
        $model = new SettingModel;
        $rateMemberLevels = [];
        $quantityMemberLevels = [];
        foreach ($data["shopping_gift_rules"] as $key => $rule) {
            switch ($rule["type"]) {
                case "payment_amount":
                    $rateMemberLevels = $rule["member_levels"];
                    break;
                case "desk":
                    $quantityMemberLevels = $rule["member_levels"];
                    break;
            }
            unset($data["shopping_gift_rules"][$key]["member_levels"]);
        }
        
        // 修改按照会员等级获得积分的值
        foreach ($rateMemberLevels as $level) {
            $update = ["points_rate" => $level["value"]];
            \app\common\model\user\Grade::where("uuid", "=", $level["uuid"])->update($update);
        }
        foreach ($quantityMemberLevels as $level) {
            $update = ["points_quantity" => $level["value"]];
            \app\common\model\user\Grade::where("uuid", "=", $level["uuid"])->update($update);
        }

        if ($model->edit(SettingEnum::POINTS, $data, $this->store['user']['shop_supplier_id'])) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("积分明细")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/user.points/log")
     * @Apidoc\Param("keyword", type="string", require=false, default="", desc="搜索关键字")
     * @Apidoc\Param("date", type="array", require=false, desc="起始日期")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\user\PointsLog\getList", desc="列表")
     */
    public function log()
    {
        $model = new PointsLogModel;
        $list = $model->getList($this->request->param());
        $attributes = $model::getAttributes();
        return $this->renderSuccess('', compact('list', 'attributes'));
    }
}
