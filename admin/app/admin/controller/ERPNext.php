<?php

namespace app\admin\controller;

use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use app\admin\model\app\App as AppModel;
use app\common\model\supplier\Supplier;
use app\admin\model\Setting;
use help\HttpHelp;


/**
 * 商家授权管理
 * @Apidoc\Group("base")
 * @Apidoc\Sort(7)
 */
class ERPNext extends Controller
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
     * @Apidoc\Param("erpnext_company_abbr", type="string", require=true, desc="ERPNext缩写")
     */
    public function add()
    {
        $param = $this->postData();
        if (empty($param['uuid']) || empty($param['erpnext_site_code']) || empty($param['erpnext_company_abbr'])) {
            return $this->renderError('参数错误');
        }
        $companySetting = Supplier::where("uuid", $param['uuid'])->find();
        if (empty($companySetting)) {
            return $this->renderError('商家不存在');
        }
        if (!empty($companySetting->erpnext_site_code) && !empty($companySetting->erpnext_company_name) && !empty($companySetting->erpnext_company_abbr)) {
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

        // TODO 同步信息到erpnext站点，获得branch_name，进行保存
        $branchName = '';

        if (!$companySetting->allowField(['erpnext_site_code', 'erpnext_company_abbr'])->save([
            'erpnext_site_code' => $param["erpnext_site_code"],
            'erpnext_company_abbr' => $param["erpnext_company_abbr"],
            'erpnext_branch_name' => $branchName,
        ])) {
            return $this->renderError('保存失败');
        }

        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("获取ERPNext站点编码列表")
     * @Apidoc\Desc("获取ERPNext站点编码列表，用于下拉框")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/erpnext/site_code")
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
     * @Apidoc\Param("company_abbr", type="string", require=false, desc="公司缩写")
     * @Apidoc\Returned("list", type="array", desc="公司列表", children={
     *      @Apidoc\Returned("company_name", type="string", desc="公司名称"),
     *      @Apidoc\Returned("company_abbr", type="string", desc="公司缩写"),
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
}
