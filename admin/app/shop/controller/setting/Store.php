<?php

namespace app\shop\controller\setting;

use help\HttpHelp;
use help\ImgHelp;
use help\DiskHelp;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\enum\settings\SettingEnum;
use app\common\model\app\App;
use app\shop\model\settings\Setting as SettingModel;

/**
 * 商城设置
 * @Apidoc\Group("system_setting")
 * @Apidoc\Sort(11)
 */
class Store extends Controller
{
    /**
     * @Apidoc\Title("商城设置(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Store/index")
     * @Apidoc\Param("name", type="string", require=true, default="", desc="商城名称")
     * @Apidoc\Param("logoUrl", type="string", require=true, default="", desc="商城logo")
     * @Apidoc\Param("zeroing_method", type="string", require=false, default="", desc="抹零方式")
     * @Apidoc\Param("ip_white_list", type="string", require=false, default="", desc="ip白名单（v1.0.6）")
     * @Apidoc\Param("time_zone", type="string", require=false, default="", desc="时区（v1.0.6）")
     * @Apidoc\Param("no_clear_table", type="string", require=false, default="", desc="结账后不清台 0-清台 1-不清台（v1.0.6）")
     * @Apidoc\Param("company", type="string", require=false, default="", desc="公司名称")
     * @Apidoc\Param("address", type="string", require=false, default="", desc="v1.0.3商城地址")
     * @Apidoc\Param("phone", type="string", require=false, default="", desc="v1.0.3商城电话")
     * @Apidoc\Param("tax_number", type="string", require=false, default="", desc="v1.0.4税号")
     * @Apidoc\Param("language", type="array", require=false, default="", desc="语言列表")
     * @Apidoc\Returned()
     */
    public function index()
    {
        if ($this->request->isGet()) {
            $ret = $this->fetchData();
            $ret['shop']['shop_supplier_id'] = $this->store['app']['uuid'];
            // 获取当前设备机器码
            $ret['shop']['device_code'] = DiskHelp::getMachineCode();
            // 获取语言列表
            $ret['shop']['language'] = $this->store['supplier']->getLanguageList();
            //
            return $this->renderSuccess('', $ret);
        }
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $name = $data['name'] ?? '';
        if (empty($name)) {
            return $this->renderError('名称不能为空');
        }
        // 长度不能超过100个字符
        if (mb_strlen($name) > 100) {
            return $this->renderError('商家名称不能超过100个字符');
        }
        $nameExist = (new App([], 0))->where('name', $name)->where('uuid', '<>', $this->store['user']['shop_supplier_id'])->find();
        if ($nameExist) {
            return $this->renderError('商家名称已存在');
        } else {
            // 更新云平台商家名称
            $supplier = new App([], 0);
            $supplier->where('uuid', $this->store['user']['shop_supplier_id'])->update(['name' => $name]);
        }
        // 判断商家名称是否存在
        if (empty($data['logoUrl'])) {
            return $this->renderError('商城logo不能为空');
        }
        // 判断商家联系电话
        if (empty($data['phone'])) {
            return $this->renderError('联系电话不能为空');
        }
        // 判断商家联系电话长度
        if (mb_strlen($data['phone']) > 20) {
            return $this->renderError('联系电话长度不能超过20个字符');
        }
        // 最长500个字符
        if (mb_strlen($data['address']) > 500) {
            return $this->renderError('地址长度不能超过500个字符');
        }
        // 如果设置了经纬度，判断是否经度取值范围为[-180,180],纬度取值范围为[-90,90] 
        if (!empty($data['coordinates'])) {
            $coordinates = explode(',', $data['coordinates']);
            if (count($coordinates) != 2) {
                return $this->renderError('经纬度格式不正确');
            }
            $lat = floatval($coordinates[0]);
            $lng = floatval($coordinates[1]);
            if ($lat < -90 || $lat > 90 || $lng < -180 || $lng > 180) {
                return $this->renderError('经纬度格式不正确');
            }
        }
        $arr = [
            'name' => $name,
            'sms_open' => $data['sms_open'] ?? 0,
            'is_get_log' => $data['is_get_log'] ?? 0,
            'wx_open' => $data['wx_open'] ?? 0,
            'avatarUrl' => ImgHelp::removeImageDomain($data['avatarUrl'] ?? ''),
            'logoUrl' => ImgHelp::removeImageDomain($data['logoUrl'] ?? ''),
            'zeroing_method' => $data['zeroing_method'] ?? '0',
            'ip_white_list' => $data['ip_white_list'] ?? '',
            'time_zone' => $data['time_zone'] ?? '',
            'no_clear_table' => $data['no_clear_table'] ?? '0',
            'company' => $data['company'] ?? '',
            'address' => $data['address'] ?? '',
            'phone' => $data['phone'] ?? '',
            'tax_number' => $data['tax_number'] ?? '',
            'language' => $data['language'] ?? [],
            'coordinates' => $data['coordinates'] ?? '',
        ];
        if ($model->edit(SettingEnum::STORE, $arr, $this->store['user']['shop_supplier_id'])) {
            $company = $this->store['app'];
            $companySetting = $this->store['app']['supplier'];
 
            if (
                $company->is_enable_erp == 1 &&
                !($companySetting->erpnext_site_code == "1" || $companySetting->erpnext_site_code == "") &&
                $companySetting->erpnext_company_abbr == $companySetting->erpnext_headquarter_abbr
            ) {
                $res = HttpHelp::postRequest('http://nginx/api/v1/admin/erpnext/shop/update_headquarter_supplier_name', json_encode([
                    'site_code' => $companySetting->erpnext_site_code,
                    'company_abbr' => $companySetting->erpnext_company_abbr,
                    'branch' => $companySetting->erpnext_branch_name,
                    'supplier_name' => $data['name'],
                    'company_uuid' => $companySetting->company_uuid,
                ]), [
                    'X-API-KEY: ' . env('JWT_SECRET'),
                    'Accept-Language: ' . request()->header('language'),
                ], 60);

                if (!$res) {
                    return $this->renderError('请求失败');
                }
                $res = json_decode($res, true);
                if ($res['code'] != 0) {
                    return $this->renderError($res['message']);
                }
            }

            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("服务器存储空间(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Store/serverStorage")
     * @Apidoc\Returned("total_space", type="string", desc="总大小")
     * @Apidoc\Returned("free_space", type="string", desc="可用大小")
     * @Apidoc\Returned("used_space", type="string", desc="已用大小")
     * @Apidoc\Returned("used_percentage", type="string", desc="已用百分比")
     */
    public function serverStorage()
    {
        return $this->renderSuccess('',  DiskHelp::getDiskSpaceInfo());
    }

    /**
     * 获取商城配置
     */
    public function fetchData()
    {
        $vars['values'] = SettingModel::getSupplierItem(SettingEnum::STORE, $this->store['user']['shop_supplier_id']);
        $vars['values']['avatarUrl'] = ImgHelp::addImageDomain($vars['values']['avatarUrl'] ?? '');
        $vars['values']['logoUrl'] = ImgHelp::addImageDomain($vars['values']['logoUrl'] ?? '');
        //
        $version = [
            'server_version' => get_version(),
        ];
        //
        return compact('vars', 'version');
    }
}
