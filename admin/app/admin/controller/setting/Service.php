<?php

namespace app\admin\controller\setting;

use help\ImgHelp;
use think\facade\Request;
use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use app\admin\validate\SettingValidate;
use app\common\enum\settings\SettingEnum;
use app\admin\model\settings\Setting as SettingModel;

/**
 * 系统设置
 * @Apidoc\Group("base")
 * @Apidoc\Sort(5)
 */
class Service extends Controller
{
    /**
     * @Apidoc\Title("系统设置")
     * @Apidoc\Desc("get请求是获取，post请求是提交修改")
     * @Apidoc\Method("GET,POST")
     * @Apidoc\Url("/api/admin/setting.service/index")
     * @Apidoc\Param("brand_name", type="string", default="", desc="品牌名称")
     * @Apidoc\Param("brand_logo", type="string", default="", desc="品牌LOGO 正方型")
     * @Apidoc\Param("brand_logo_long", type="string", default="", desc="品牌LOGO 长方型")
     * @Apidoc\Param("browser_logo", type="string", default="", desc="v1.0.4.1浏览器LOGO")
     * @Apidoc\Param("browser_title", type="string", default="", desc="v1.0.4.1浏览器标题")
     * @Apidoc\Param("expiration_reminder", type="int", default="", desc="到期提醒 0为永久")
     * @Apidoc\Param("member_default_avatar", type="string", default="", desc="会员默认头像")
     * @Apidoc\Param("auth_code_bind_validity_period", type="string", default="", desc="验证码绑定有效期")
     */
    public function index(SettingValidate $validate)
    {
        if ($this->request->isGet()) {
            return $this->fetchData();
        }
        $param = $validate->goCheck('edit');
        $param['brand_logo'] = ImgHelp::removeImageDomain($param['brand_logo'] ?? '');
        $param['brand_logo_long'] = ImgHelp::removeImageDomain($param['brand_logo_long'] ?? '');
        $param['browser_logo'] = ImgHelp::removeImageDomain($param['browser_logo'] ?? '');
        $param['member_default_avatar'] = ImgHelp::removeImageDomain($param['member_default_avatar'] ?? '');
        $model = new SettingModel;
        if ($model->add($param)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * 获取系统设置
     */
    public function fetchData()
    {
        $vars['values'] = SettingModel::getSysConfig(SettingEnum::SYS_ADMIN_CONFIG);
        //
        // $file = root_path() . '/views/shop/package.json';
        // $version = json_decode(file_get_contents($file), true);
        $vars['version'] = [
            'client_version' => get_version(),
            'server_version' => get_version(),
        ];
        //
        return $this->renderSuccess('', compact('vars'));
    }

    /**
     * @Apidoc\Title("获取基本信息")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/setting.service/info")
     * @Apidoc\Returned("base", type="array", children={
     *      @Apidoc\Property("brand_name",type="string", desc="品牌名称"),
     *      @Apidoc\Property("brand_logo",type="string", desc="正方形品牌logo"),
     *      @Apidoc\Property("brand_logo_long",type="string", desc="长方形品牌logo"),
     *      @Apidoc\Property("browser_logo",type="string", desc="浏览器LOGO"),
     *      @Apidoc\Property("browser_title",type="string", desc="浏览器标题"),
     *      @Apidoc\Property("expiration_reminder",type="string", desc="剩余多少天到期提示"),
     * })
     */
    public function info(Request $request)
    {
        // todo: 获取系统设置
        // $values = SettingModel::getSysConfig(SettingEnum::SYS_ADMIN_CONFIG);
        $base = [
            // 'brand_name' => $values['brand_name'] ?? '',                                        //  品牌名称
            // 'brand_logo' => ImgHelp::addImageDomain($values['brand_logo'] ?? ''),               //  正方形品牌logo
            // 'brand_logo_long' => ImgHelp::addImageDomain($values['brand_logo_long'] ?? ''),     //  长方形品牌logo
            // 'browser_logo' => ImgHelp::addImageDomain($values['browser_logo'] ?? ''),           //  浏览器LOGO
            // 'browser_title' => $values['browser_title'] ?? '',                                  //  浏览器标题
            // 'expiration_reminder' => $values['expiration_reminder'] ?? 0,                       //  剩余多少天到期提示
        ];
        //
        return $this->renderSuccess('',  compact('base'));
    }
}
