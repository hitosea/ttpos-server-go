<?php

namespace app\admin\controller;

use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use app\admin\model\app\App as AppModel;
use app\common\model\supplier\Supplier;
use app\admin\model\Setting;
use help\HttpHelp;
use app\admin\model\admin\User;


/**
 * 商家授权管理
 * @Apidoc\Group("base")
 * @Apidoc\Sort(7)
 */
class Erpnext extends Controller
{
    /**
     * @Apidoc\Title("授权商家列表")
     * @Apidoc\Desc("授权商家列表")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/erpnext/index")
     * @Apidoc\Param("keyword", type="string", require=true, desc="商家名称/ID")
     * @Apidoc\Param("configured", type="int", require=false, default="0", desc="是否仅获取已授权的商家：0-否；1-是")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", desc="商家列表", children={
     *      @Apidoc\Returned("total", type="int", desc="总数"),
     *      @Apidoc\Returned("page", type="int", desc="页码"),
     *      @Apidoc\Returned("limit", type="int", desc="每页条数"),
     *      @Apidoc\Returned("data", type="array", desc="商家列表", children={
     *          @Apidoc\Returned("uuid", type="biginteger", desc="商家ID"),
     *          @Apidoc\Returned("name", type="string", desc="商家名称"),
     *          @Apidoc\Returned("link_phone", type="string", desc="联系电话"),
     *          @Apidoc\Returned("admin", type="string", desc="超管用户名"),
     *          @Apidoc\Returned("erpnext_site_code", type="string", desc="ERPNext站点编码"),
     *          @Apidoc\Returned("erpnext_company_abbr", type="string", desc="ERPNext公司缩写"),
     *          @Apidoc\Returned("erpnext_branch_name", type="string", desc="ERPNext分支名称"),
     *          @Apidoc\Returned("erpnext_pos_profile_name", type="string", desc="ERPNextPos Profile名称"),
     *      })
     * })
     */
    public function index()
    {
        $param = $this->getData();
        $list = (new AppModel)->getErpnextCompanyList($param)?->toArray();
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("新增授权商家")
     * @Apidoc\Desc("新增授权商家")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/api/admin/erpnext/add")
     * @Apidoc\Param("uuid", type="biginteger", require=true, desc="商家ID")
     * @Apidoc\Param("erpnext_site_code", type="string", require=true, desc="ERPNext编码")
     * @Apidoc\Param("erpnext_company_abbr", type="string", require=true, desc="ERPNext公司缩写")
     * @Apidoc\Param("erpnext_default_company_abbr", type="string", require=true, desc="ERPNext默认公司缩写，用于同步单位和属性")
     * @Apidoc\Param("password", type="string", require=true, desc="密码验证")
     */
    public function add()
    {
        $param = $this->postData();
        if (
            empty($param['uuid']) ||
            !isset($param['erpnext_site_code']) ||
            empty($param['erpnext_company_abbr']) ||
            !isset($param['headquarter_abbr']) ||
            empty($param['password'])
        ) {
            return $this->renderError('参数错误');
        }
        // 验证用户名密码是否正确
        $user = User::withTrashed()->whereRaw('BINARY username = :username', ['username' => $this->admin['user']['user_name']])->order('admin_user_id', 'desc')->order('delete_time')->find();
        if (!$user || $user->password != salt_hash($param['password'] ?? '')) {
            return $this->renderError('密码错误');
        }

        $companySetting = Supplier::where("uuid", $param['uuid'])->with('app')->find();
        if (empty($companySetting)) {
            return $this->renderError('商家不存在');
        }

        if (!empty($companySetting->erpnext_site_code) && !empty($companySetting->erpnext_company_name) && !empty($companySetting->erpnext_company_abbr) && !empty($companySetting->erpnext_branch_name)) {
            return $this->renderError('商家已授权');
        }
        // 读取setting表的erpnext_site
        $erpnextSite = Setting::where("key", "erpnext_site")->find();
        if (empty($erpnextSite)) {
            return $this->renderError('ERPNext站点不存在');
        }
        $erpnextSite->values = array_filter($erpnextSite->values, function ($item) use ($param) {
            return $item['code'] == $param['erpnext_site_code'];
        });
        if (empty($erpnextSite->values)) {
            return $this->renderError('ERPNext站点不存在');
        }

        $params = [
            'site_code' => $param['erpnext_site_code'],
            'company_abbr' => $param['erpnext_company_abbr'],
            'company_uuid' => $param['uuid'],
            'headquarter_abbr' => $param['headquarter_abbr'],
        ];
        $res = HttpHelp::postRequest('http://nginx/api/v1/admin/erpnext/shop/init', json_encode($params), [
            'X-API-KEY: ' . env('JWT_SECRET'),
            'Accept-Language: ' . request()->header('language'),
        ]);
        if (!$res) {
            return $this->renderError('请求失败');
        }
        $res = json_decode($res, true);
        if ($res['code'] != 0) {
            return $this->renderError($res['message']);
        }

        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("获取ERPNext站点编码列表")
     * @Apidoc\Desc("获取ERPNext站点编码列表，用于下拉框")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/erpnext/siteCode")
     * @Apidoc\Returned("list", type="array", desc="ERPNext站点列表", children={
     *      @Apidoc\Returned("name", type="string", desc="ERPNext站点名称"),
     *      @Apidoc\Returned("code", type="string", desc="ERPNext站点编码"),
     * })
     */
    function siteCode()
    {
        $erpnextSite = Setting::where("key", "erpnext_site")->find();
        return $this->renderSuccess('', ['list' => $erpnextSite->values]);
    }

    /**
     * @Apidoc\Title("获取ERPNext站点公司名称")
     * @Apidoc\Desc("获取ERPNext站点公司名称，如果company_abbr不为空，则根据company_abbr查询")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/erpnext/siteCompany")
     * @Apidoc\Param("site_code", type="string", require=true, desc="ERPNext编码")
     * @Apidoc\Param("company_name", type="string", require=false, desc="根据公司名称筛选")
     * @Apidoc\Param("company_abbr", type="string", require=false, desc="根据公司缩写编码筛选")
     * @Apidoc\Param("parent_company", type="string", require=false, desc="根据父公司名称筛选")
     * @Apidoc\Returned("list", type="array", desc="公司列表、树形结构", children={
     *      @Apidoc\Returned("company_name", type="string", desc="公司名称"),
     *      @Apidoc\Returned("company_abbr", type="string", desc="公司缩写"),
     *      @Apidoc\Returned("is_used", type="boolean", desc="是否已被使用"),
     *      @Apidoc\Returned("children", type="array", desc="子公司列表", children={
     *          @Apidoc\Returned("company_name", type="string", desc="公司名称"),
     *          @Apidoc\Returned("company_abbr", type="string", desc="公司缩写"),
     *          @Apidoc\Returned("is_used", type="boolean", desc="是否已被使用"),
     *      }),
     * })
     */
    function siteCompany()
    {
        $res = HttpHelp::getRequest('http://nginx/api/v1/admin/erpnext/site/company', $this->getData(), [
            'X-API-KEY: ' . env('JWT_SECRET'),
            'Accept-Language: ' . request()->header('language'),
        ]);
        if (!$res) {
            return $this->renderError('请求失败');
        }
        $res = json_decode($res, true);
        if ($res['code'] != 0) {
            return $this->renderError($res['message']);
        }
        return $this->renderSuccess('', $res['data']);
    }


    /**
     * @Apidoc\Title("获取ERPNext支付方式列表")
     * @Apidoc\Desc("获取ERPNext支付方式列表")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/erpnext/paymentMethodList")
     * @Apidoc\Param("company_uuid", type="biginteger", require=true, desc="公司ID")
     * @Apidoc\Returned("list", type="array", desc="公司支付方式", children={
     *      @Apidoc\Returned("name", type="string", desc="支付方式名称"),
     *      @Apidoc\Returned("is_addable", type="boolean", desc="是否可添加"),
     * })
     */
    function paymentMethodList()
    {
        $res = HttpHelp::getRequest('http://nginx/api/v1/admin/erpnext/payment_method/list', $this->getData(), [
            'X-API-KEY: ' . env('JWT_SECRET'),
            'Accept-Language: ' . request()->header('language'),
        ]);
        if (!$res) {
            return $this->renderError('请求失败');
        }
        $res = json_decode($res, true);
        if ($res['code'] != 0) {
            return $this->renderError($res['message']);
        }
        return $this->renderSuccess('', $res['data']);
    }

    /**
     * @Apidoc\Title("给授权erpnext的商家添加支付方式")
     * @Apidoc\Desc("给授权erpnext的商家添加支付方式")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/api/admin/erpnext/addPaymentMethod")
     * @Apidoc\Param("company_uuid", type="biginteger", require=true, desc="公司UUID")
     * @Apidoc\Param("erpnext_payment", type="string", require=true, desc="ERPNext支付方式")
     * @Apidoc\Param("name", type="string", require=true, desc="支付方式名称")
     * @Apidoc\Param("fee", type="float", require=true, desc="手续费")
     * @Apidoc\Param("sort", type="int", require=true, desc="排序")
     * @Apidoc\Param("status", type="int", require=true, desc="状态: 0-禁用 1-启用")
     * @Apidoc\Param("checkout_show", type="array", require=false, desc="结账显示，可选cashier、assistant")
     * @Apidoc\Param("member_recharge_show", type="array", require=false, desc="会员充值显示，可选cashier")
     */
    function addPaymentMethod()
    {
        $res = HttpHelp::postRequest('http://nginx/api/v1/admin/erpnext/payment_method/add', json_encode($this->postData()), [
            'X-API-KEY: ' . env('JWT_SECRET'),
            'Accept-Language: ' . request()->header('language'),
        ]);
        if (!$res) {
            return $this->renderError('请求失败');
        }
        $res = json_decode($res, true);
        if ($res['code'] != 0) {
            return $this->renderError($res['message']);
        }
        return $this->renderSuccess('添加成功');
    }
}
