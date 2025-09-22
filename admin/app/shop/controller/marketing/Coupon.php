<?php

namespace app\shop\controller\marketing;

use app\shop\service\marketing\MarketingCouponService;
use think\App;
use think\Request;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\enum\http\StatusCode;
use app\common\exception\BaseException;

/**
 * 优惠券
 * @Apidoc\Group("marketing")
 * @Apidoc\Sort(1)
 */
class Coupon extends Controller
{
    protected $service;

    public function __construct(App $app)
    {
        parent::__construct($app);
        // 验证功能是否开启
        if ((request()->userInfo['supplier']['is_open_coupon'] ?? 0) != 1) {
            throw new BaseException(['msg' => '该营销活动需商家开通优惠券功能，请联系销售代表处理', 'code' => StatusCode::ERROR]);
        }
        // 
        $this->service = new MarketingCouponService();
    }

    /**
     * @Apidoc\Title("优惠券列表")
     * @Apidoc\Method ("GET")
     * @Apidoc\Url ("/index.php/shop/marketing.coupon/list")
     * @Apidoc\Query("name", type="string", require=false, desc="优惠券名称")
     * @Apidoc\Query("type", type="string", require=false, desc="优惠券类型")
     * @Apidoc\Query(ref="pageParam")
     * @Apidoc\Returned("list", type="array", children={
     *      @Apidoc\Returned("uuid", type="string", desc="优惠券Uuid"),
     *      @Apidoc\Returned("sort", type="integer", desc="排序"),
     *      @Apidoc\Returned("name", type="string", desc="优惠券名称"),
     *      @Apidoc\Returned("type", type="string", desc="优惠券类型: deduction - 抵扣券"),
     *      @Apidoc\Returned("amount", type="float", desc="金额"),
     *      @Apidoc\Returned("valid_date", type="string", desc="有效日期"),
     *      @Apidoc\Returned("valid_day_time_range", type="string", desc="适用时间"),
     *      @Apidoc\Returned("count", type="integer", desc="数量"),
     *      @Apidoc\Returned("create_time", type="integer", desc="添加时间"),
     *      @Apidoc\Returned("status", type="integer", desc="状态 1 开启 0 禁用"),
     *      @Apidoc\Returned("can_delete", type="boolean", desc="是否可以删除 true 可以 false 不可以"),
     * })
     */
    public function list(Request $request)
    {
        $params = $request->get();
        $list = $this->service->getList($params);
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("创建优惠券")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/marketing.coupon/add")
     * @Apidoc\Param("name", type="string", require=true, desc="优惠券名称")
     * @Apidoc\Param("sort", type="integer", require=true, desc="排序, 1-99")
     * @Apidoc\Param("type", type="string", desc="优惠券类型: deduction - 抵扣券")
     * @Apidoc\Param("deduction_type", type="string", desc="抵扣类型: taxed - 税后抵扣")
     * @Apidoc\Param("amount", type="float", desc="金额", require=true)
     * @Apidoc\Param("count", type="integer", desc="优惠券数量, 最大999999", require=true)
     * @Apidoc\Param("day_start_time", type="string", desc="每日适用时段开始时间, hh:mm 格式", require=true)
     * @Apidoc\Param("day_end_time", type="string", desc="每日适用时段结束时间, hh:mm 格式, 00:00 到 23:59 表示全天", require=true)
     * @Apidoc\Param("requirement", type="string", require=true, desc="获得优惠券所需条件: none - 都可以获取; marketing - 营销活动")
     * @Apidoc\Param("valid_start_time", type="string", desc="优惠券有效开始时间, requirement = none 时有效, yyyy-mm-dd 格式")
     * @Apidoc\Param("valid_end_time", type="string", desc="优惠券有效结束时间, requirement = none 时有效, yyyy-mm-dd 格式")
     * @Apidoc\Param("valid_days", type="integer", desc="领取优惠券后n天内有效, requirement = marketing 时有效")
     * @Apidoc\Returned("uuid", type="string", desc="优惠券Uuid")
     */
    public function add(Request $request)
    {
        $data = $request->post();
        $result = $this->service->create($data);
        if (isset($result['code']) && $result['code'] !== 0) {
            return $this->renderError($result['msg'] ?? '创建失败');
        }
        return $this->renderSuccess('创建成功', $result);
    }

    /**
     * @Apidoc\Title("编辑优惠券")
     * @Apidoc\Method ("POST,GET")
     * @Apidoc\Desc("GET请求为编辑页面，POST请求为编辑保存")
     * @Apidoc\Url ("/index.php/shop/marketing.coupon/edit")
     * @Apidoc\Query("uuid", type="string", require=true, desc="优惠券uuid")
     * @Apidoc\Returned("name", type="string", require=true, desc="优惠券名称")
     * @Apidoc\Returned("sort", type="integer", require=true, desc="排序, 1-99")
     * @Apidoc\Returned("type", type="string", desc="优惠券类型: deduction - 抵扣券")
     * @Apidoc\Returned("deduction_type", type="string", desc="抵扣类型: taxed - 税后抵扣")
     * @Apidoc\Returned("amount", type="float", desc="金额", require=true)
     * @Apidoc\Returned("count", type="integer", desc="优惠券数量, 最大999999", require=true)
     * @Apidoc\Returned("day_start_time", type="string", desc="每日适用时段开始时间, hh:mm 格式", require=true)
     * @Apidoc\Returned("day_end_time", type="string", desc="每日适用时段结束时间, hh:mm 格式, 00:00 到 23:59 表示全天", require=true)
     * @Apidoc\Returned("requirement", type="string", require=true, desc="获得优惠券所需条件: none - 都可以获取; marketing - 营销活动")
     * @Apidoc\Returned("valid_start_time", type="string", desc="优惠券有效开始时间, requirement = none 时有效, yyyy-mm-dd 格式")
     * @Apidoc\Returned("valid_end_time", type="string", desc="优惠券有效结束时间, requirement = none 时有效, yyyy-mm-dd 格式")
     * @Apidoc\Returned("valid_days", type="integer", desc="领取优惠券后n天内有效, requirement = marketing 时有效")
     */
    public function edit(Request $request, $uuid)
    {
        if ($request->isGet()) {
            $result = $this->service->getDetail($uuid);
            if (!$result) {
                return $this->renderError('优惠券不存在');
            }
            return $this->renderSuccess('', [
                'detail' => $result,
            ]);
        }
        $data = $request->post();
        $result = $this->service->update($uuid, $data);
        if (isset($result['code']) && $result['code'] !== 0) {
            return $this->renderError($result['msg'] ?? '编辑失败');
        }
        return $this->renderSuccess('编辑成功', $result);
    }


    /**
     * @Apidoc\Title("优惠券记录")
     * @Apidoc\Method ("GET")
     * @Apidoc\Url ("/index.php/shop/marketing.coupon/record")
     * @Apidoc\Query("coupon_name", type="string", require=false, desc="优惠券名称")
     * @Apidoc\Query("coupon_type", type="string", require=false, desc="优惠券类型")
     * @Apidoc\Query("record_type", type="string", require=false, desc="记录类型")
     * @Apidoc\Param("create_time", type="array", require=false, default="", desc="时间")
     * @Apidoc\Query(ref="pageParam")
     * @Apidoc\Returned("list", type="array", desc="优惠券记录列表", children={
     *      @Apidoc\Returned("coupon_name", type="string", desc="优惠券名称"),
     *      @Apidoc\Returned("serial_no", type="string", desc="记录编号"),
     *      @Apidoc\Returned("record_type", type="string", desc="记录类型：1-首次添加、2-调整添加、3-调整扣减、4-活动扣减、5、奖励领取（冻结）、6、核销扣减"),
     *      @Apidoc\Returned("create_time", type="integer", desc="时间"),
     *      @Apidoc\Returned("coupon_count", type="integer", desc="数量"),
     *      @Apidoc\Returned("left_count", type="integer", desc="剩余有效张数"),
     * })
     */
    public function record(Request $request)
    {
        $params = $request->get();
        $list = $this->service->getRecord($params);
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("删除优惠券")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/marketing.coupon/delete")
     * @Apidoc\Param("uuid", type="string", require=true, desc="优惠券uuid")
     * @Apidoc\Returned()
     */
    public function delete(Request $request)
    {
        $params = $request->post();
        $result = $this->service->delete($params['uuid']);
        if (isset($result['code']) && $result['code'] !== 0) {
            return $this->renderError($result['msg'] ?? '删除失败');
        }
        return $this->renderSuccess('删除成功', $result);
    }

    /**
     * @Apidoc\Title("修改优惠券状态")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/marketing.coupon/status")
     * @Apidoc\Param("uuid", type="string", require=true, desc="优惠券uuid")
     * @Apidoc\Param("status", type="integer", require=true, desc="状态 1 开启 0 禁用")
     * @Apidoc\Returned()
     */
    public function status(Request $request)
    {
        $params = $request->post();
        $result = $this->service->status($params['uuid'], $params);
        if (isset($result['code']) && $result['code'] !== 0) {
            return $this->renderError($result['msg'] ?? '修改状态失败');
        }
        return $this->renderSuccess('修改状态成功', $result);
    }
} 