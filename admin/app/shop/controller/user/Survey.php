<?php

namespace app\shop\controller\user;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\user\User as UserModel;

/**
 * 会员概况（v1.1.0）
 * @Apidoc\Group("user")
 * @Apidoc\Sort(7)
 */
class Survey extends Controller
{
    /**
     * @Apidoc\Title("账户数据")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/user.survey/index")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("detail", type="array", desc="订单概况", children={
     *    @Apidoc\Param("balance",type="int",default=0,desc="主账户余额"),
     *    @Apidoc\Param("gift_balance",type="int",default=0,desc="赠送账户余额"),
     * })


     */
    public function index()
    {
        $data = $this->postData() ?: [];
        $model = new UserModel();
        $detail = $model->storeOverview($data);
        return $this->renderSuccess('', compact('detail'));
    }

    /**
     * 充值排行榜
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/user.survey/rechargeRank")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", desc="充值排行榜", children={
     *    @Apidoc\Returned("user_id", type="string", desc="用户ID"),
     *    @Apidoc\Returned("nickname", type="string", desc="昵称"),
     *    @Apidoc\Returned("recharge amount", type="int", desc="充值金额"),
     *    @Apidoc\Returned("gift_amount", type="float", desc="赠送金额"),
     * })
     */
    public function rechargeRank()
    {
        $data = $this->postData() ?: [];
        $model = new UserModel();
        $list = $model->getRechargeRank($data);
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * 消费排行榜
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/user.survey/consumerRank")
     * @Apidoc\Param("sort", type="int", require=false, default="", desc="排序方式 1-消费次数 2-消费金额")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", desc="消费排行榜", children={
     *    @Apidoc\Returned("user_id", type="string", desc="用户ID"),
     *    @Apidoc\Returned("nickname", type="string", desc="昵称"),
     *    @Apidoc\Returned("consumption_num", type="int", desc="消费次数"),
     *    @Apidoc\Returned("consumption_amount", type="float", desc="消费金额"),
     * })
     */
    public function consumerRank()
    {
        $data = $this->postData() ?: [];
        $model = new UserModel();
        $list = $model->getConsumerRank($data);
        return $this->renderSuccess('', compact('list'));
    }
}
