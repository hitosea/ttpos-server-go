<?php

namespace app\admin\controller;

use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use app\admin\model\app\App as AppModel;
use app\common\model\supplier\Supplier;
use app\admin\model\Setting;


/**
 * 商家授权管理
 * @Apidoc\Group("base")
 * @Apidoc\Sort(7)
 */
class ERPNext extends Controller
{
    /**
     * @Apidoc\Title("授权商家列表")
     * @Apidoc\Desc("get请求是获取，返回数组")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/erpnext/index")
     * @Apidoc\Param("keyword", type="string", require=true, desc="商家名称/ID")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("uuid", type="biginteger", desc="商家ID")
     * @Apidoc\Returned("name", type="string", desc="商家名称")
     * @Apidoc\Returned("link_phone", type="string", desc="联系电话")
     * @Apidoc\Returned("admin", type="string", desc="超管用户名")
     */
    public function index()
    {
        $param = $this->getData();
        $param = array_merge($param, ["configured" => true]);
        $list = (new AppModel)->getErpnextCompanyList($param)?->toArray();
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("新增授权商家")
     * @Apidoc\Desc("新增授权商家")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/api/admin/erpnext/add")
     * @Apidoc\Param("uuid", type="biginteger", require=true, desc="商家ID")
     * @Apidoc\Param("erpnext_code", type="string", require=true, desc="ERPNext编码")
     * @Apidoc\Param("erpnext_name", type="string", require=true, desc="ERPNext名称")
     */
    public function add()
    {
        $param = $this->postData();
        if (empty($param['uuid']) || empty($param['erpnext_code']) || empty($param['erpnext_name'])) {
            return $this->renderError('参数错误');
        }
        $companySetting = Supplier::where("uuid", $param['uuid'])->find();
        if (empty($companySetting)) {
            return $this->renderError('商家不存在');
        }
        if (!empty($companySetting->erpnext_code) && !empty($companySetting->erpnext_name)) {
            return $this->renderError('商家已授权');
        }
        // 读取setting表的erpnext_site
        $erpnextSite = Setting::where("key", "erpnext_site")->find();
        if (empty($erpnextSite)) {
            return $this->renderError('ERPNext站点不存在');
        }
        $erpnextSite->values = array_filter($erpnextSite->values, function ($item) use ($param) {
            return $item['code'] == $param['erpnext_code'];
        });
        if (empty($erpnextSite->values)) {
            return $this->renderError('ERPNext站点不存在');
        }
        if (!$companySetting->allowField(['erpnext_code', 'erpnext_name'])->save([
            'erpnext_code' => $param["erpnext_code"],
            'erpnext_name' => $param["erpnext_name"],
        ])) {
            return $this->renderError('保存失败');
        }

        // TODO 同步信息到erpnext站点

        return $this->renderSuccess('操作成功');
    }

    /**
     * 获取ERPNext站点公司名称
     * @param string $erpnext_code
     * @return array
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/erpnext/siteCompany")
     * @Apidoc\Param("erpnext_code", type="string", require=true, desc="ERPNext编码")
     * @Apidoc\Returned("name", type="string", desc="公司名称")
     * @Apidoc\Returned("code", type="int", desc="公司编码")
     * @Apidoc\Returned("url", type="string", desc="公司URL")
     */
    function siteCompany($erpnext_code)
    {
        // TODO 调用中台接口，获取ERPNext站点公司名称，返回下来列表，比如 
        // [
        //     "name" => "TTPOS",
        //     "code" => 1,
        //     "url" => "http://192.168.100.206:15080"
        // ]
        // 
        // $erpnextSite->values = array_filter($erpnextSite->values, function($item){
        //     return $item['code'] == $param['erpnext_code'];
        // });
        // if (empty($erpnextSite->values)) {
        //     return $this->renderError('ERPNext站点不存在');
        // }
    }
}
