<?php

namespace app\shop\controller\marketing;

use think\App;
use think\Request;
use Endroid\QrCode\QrCode;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\enum\http\StatusCode;
use Endroid\QrCode\Writer\PngWriter;
use app\common\exception\BaseException;
use app\shop\service\marketing\MarketingActivityService;

/**
 * 邀请有礼活动
 * @Apidoc\Group("marketing")
 * @Apidoc\Sort(1)
 */
class Activity extends Controller
{
    protected $service;

    public function __construct(App $app)
    {
        parent::__construct($app);
        // 验证功能是否开启
        if ((request()->userInfo['supplier']['is_open_member'] ?? 0) != 1) {
            throw new BaseException(['msg' => '该营销活动需商家开通会员中心功能，请联系销售代表处理', 'code' => StatusCode::ERROR]);
        }
        // 验证功能是否开启
        if ((request()->userInfo['supplier']['is_open_marketing'] ?? 0) != 1) {
            throw new BaseException(['msg' => '该营销活动需商家开通营销活动功能，请联系销售代表处理', 'code' => StatusCode::ERROR]);
        }
        // 
        $this->service = new MarketingActivityService();
    }

    /**
     * @Apidoc\Title("邀请有礼活动列表")
     * @Apidoc\Method ("GET")
     * @Apidoc\Url ("/index.php/shop/marketing.activity/list")
     * @Apidoc\Query("name", type="string", require=false, desc="活动名称")
     * @Apidoc\Query(ref="pageParam")
     * @Apidoc\Returned("list", type="array", children={
     *      @Apidoc\Returned("uuid", type="string", desc="活动uuid"),
     *      @Apidoc\Returned("name", type="string", desc="活动名称"),
     *      @Apidoc\Returned("type", type="integer", desc="活动类型 0邀请有礼"),
     *      @Apidoc\Returned("start_time", type="integer", desc="开始时间"),
     *      @Apidoc\Returned("end_time", type="integer", desc="结束时间"),
     *      @Apidoc\Returned("status", type="integer", desc="状态 0未开始 1进行中 2已结束"),
     *      @Apidoc\Returned("status_text", type="string", desc="状态文本 未开始 进行中 已结束"),
     *      @Apidoc\Returned("create_time", type="integer", desc="创建时间"),
     *      @Apidoc\Returned("prizes", type="array", children={
     *          @Apidoc\Returned("prize_type", type="integer", desc="奖品类型 1优惠券"),
     *          @Apidoc\Returned("prize_uuid", type="string", desc="奖品uuid"),
     *          @Apidoc\Returned("coupon_name", type="string", desc="优惠券名称")
     *      }),
     * })
     */
    public function list(Request $request)
    {
        $params = $request->get();
        $list = $this->service->getList($params);
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("创建邀请有礼活动")
     * @Apidoc\Method ("GET,POST")
     * @Apidoc\Desc("GET请求为创建页面，POST请求为创建保存")
     * @Apidoc\Url ("/index.php/shop/marketing.activity/add")
     * @Apidoc\Param("name", type="string", require=true, desc="活动名称（json多语言）", mock="{'th':'测试'}")
     * @Apidoc\Param("description", type="string", require=true, desc="活动描述 （json多语言）", mock="{'th':'描述测试'}")
     * @Apidoc\Param("start_time", type="string", require=true, desc="开始时间", mock="2025-05-30 10:00:00")
     * @Apidoc\Param("end_time", type="string", require=true, desc="结束时间", mock="2025-05-30 12:00:00")
     * @Apidoc\Param("reward_condition_amount", type="float", require=true, desc="奖励条件金额", mock="100")
     * @Apidoc\Param("is_open_reward_limit", type="integer", require=true, desc="是否开启奖励次数限制 0否 1是", mock="1")
     * @Apidoc\Param("reward_limit", type="integer", require=true, desc="奖励次数限制", mock="10")
     * @Apidoc\Param("prize_list", type="array", require=true, desc="奖品列表", children={
     *      @Apidoc\Param("prize_type", type="integer", require=true, desc="奖品类型：1优惠券 2未知", mock="1"),
     *      @Apidoc\Param("prize_uuid", type="string", require=true, desc="奖品UUID 优惠券时必填，为优惠券的uuid", mock="1731155609"),
     * })
     * @Apidoc\Param("image_base64", type="string", require=false, desc="活动图片")
     * @Apidoc\Returned("uuid", type="string", desc="活动uuid")
     */
    public function add(Request $request)
    {
        if ($request->isGet()) {
            $url = env('MEMBER_BASE_URL') . "/home?cid={$request->appId}";
            $qrCode = new QrCode($url);
            return $this->renderSuccess('', [
                'qr_code_url' => $url,
                'qr_code' => (new PngWriter())->write($qrCode)->getDataUri()
            ]);
        }
        $data = $request->post();
        $result = $this->service->create($data);
        if (isset($result['code']) && $result['code'] !== 0) {
            return $this->renderError($result['msg'] ?? '创建失败');
        }
        return $this->renderSuccess('创建成功', $result);
    }

    /**
     * @Apidoc\Title("编辑邀请有礼活动")
     * @Apidoc\Method ("POST,GET")
     * @Apidoc\Desc("GET请求为编辑页面，POST请求为编辑保存")
     * @Apidoc\Url ("/index.php/shop/marketing.activity/edit")
     * @Apidoc\Query("uuid", type="string", require=true, desc="活动uuid")
     * @Apidoc\Param("image_base64", type="string", require=false, desc="活动图片")
     * @Apidoc\Param("prize_list", type="array", require=true, desc="奖品列表", children={
     *      @Apidoc\Param("prize_type", type="integer", require=true, desc="奖品类型：1优惠券 2未知", mock="1"),
     *      @Apidoc\Param("prize_uuid", type="string", require=true, desc="奖品UUID 优惠券时必填，为优惠券的uuid", mock="1731155609"),
     * })
     * @Apidoc\Returned("uuid", type="string", desc="活动uuid")
     * @Apidoc\Returned("name", type="string", desc="活动名称"),
     * @Apidoc\Returned("description", type="string", desc="活动描述"),
     * @Apidoc\Returned("start_time", type="integer", desc="开始时间"),
     * @Apidoc\Returned("end_time", type="integer", desc="结束时间"),
     * @Apidoc\Returned("status", type="integer", desc="状态 0未开始 1进行中 2已结束"),
     * @Apidoc\Returned("status_text", type="string", desc="状态文本 未开始 进行中 已结束"),
     * @Apidoc\Returned("reward_condition_amount", type="float", desc="奖励金额"),
     * @Apidoc\Returned("reward_limit", type="integer", desc="奖励次数限制"),
     * @Apidoc\Returned("is_open_reward_limit", type="integer", desc="是否开启奖励次数限制 0否 1是"),
     * @Apidoc\Returned("create_time", type="integer", desc="创建时间"),
     * @Apidoc\Returned("image_base64", type="string", desc="活动图片"),
     */
    public function edit(Request $request, $uuid)
    {
        if ($request->isGet()) {
            $result = $this->service->getDetail($uuid);
            if (!$result) {
                return $this->renderError('活动不存在');
            }
            $url = env('MEMBER_BASE_URL') . "/home?cid={$request->appId}";
            $qrCode = new QrCode($url);
            return $this->renderSuccess('', [
                'detail' => $result,
                'qr_code_url' => $url,
                'qr_code' => (new PngWriter())->write($qrCode)->getDataUri()
            ]);
        }
        // 
        $data = $request->post();
        $result = $this->service->update($uuid, $data);
        if (isset($result['code']) && $result['code'] !== 0) {
            return $this->renderError($result['msg'] ?? '编辑失败');
        }
        return $this->renderSuccess('编辑成功', $result);
    }

    /**
     * @Apidoc\Title("失效邀请有礼活动")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/marketing.activity/disable")
     * @Apidoc\Param("uuid", type="string", require=true, desc="活动uuid")
     * @Apidoc\Returned("result", type="bool", desc="操作结果")
     */
    public function disable(Request $request, $uuid)
    {
        $result = $this->service->disable($uuid);
        if (!$result) {
            return $this->renderError('操作失败');
        }
        return $this->renderSuccess('操作成功', ['result' => true]);
    }

    /**
     * @Apidoc\Title("邀请有礼奖励记录")
     * @Apidoc\Method ("GET")
     * @Apidoc\Url ("/index.php/shop/marketing.activity/record")
     * @Apidoc\Query("activity_uuid", type="string", require=false, desc="活动uuid")
     * @Apidoc\Query("keyword", type="string", require=false, desc="暱稱/手機號/ID")
     * @Apidoc\Returned("list", type="array", desc="奖励记录列表", children={
     *      @Apidoc\Returned("uuid", type="string", desc="记录uuid"),
     *      @Apidoc\Returned("member_uuid", type="string", desc="会员uuid"),
     *      @Apidoc\Returned("nickname", type="string", desc="会员昵称"),
     *      @Apidoc\Returned("phone", type="string", desc="会员手机号"),
     *      @Apidoc\Returned("reward_count", type="integer", desc="奖励次数"),
     *      @Apidoc\Returned("last_reward_time", type="string", desc="最后奖励时间"),
     * })
     */
    public function record(Request $request)
    {
        $params = $request->get();
        $list = $this->service->getRecord($params);
        return $this->renderSuccess('', compact('list'));
    }
} 