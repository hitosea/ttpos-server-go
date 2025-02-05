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
        if ($model->edit(SettingEnum::POINTS, $this->postData(), $this->store['user']['shop_supplier_id'])) {
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
