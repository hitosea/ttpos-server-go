<?php

namespace app\admin\controller;

use captcha\facade\Captcha;
use hg\apidoc\annotation as Apidoc;
use app\admin\validate\PassportValidate;
use app\admin\validate\AdminUserValidate;
use app\common\enum\settings\SettingEnum;
use app\admin\model\admin\User as UserModel;
use app\common\service\keypair\KeypairService;
use app\admin\model\admin\User as AdminUserModel;
use app\common\model\settings\Setting as SettingModel;

/**
 * 登录相关
 * @Apidoc\Group("base")
 * @Apidoc\Sort(1)
 */
class Passport extends Controller
{
    /**
     * @Apidoc\Title("登录页logo")
     * @Apidoc\Method("GET")
     * @Apidoc\Desc("直接用于img标签的src")
     * @Apidoc\Url("/api/admin/passport/logo")
     */
    public function logo()
    {
        $setting = SettingModel::getSysConfig(SettingEnum::SYS_ADMIN_CONFIG);
        $brandLogo = $setting['brand_logo'];
        $brandLogos = explode('?', $brandLogo);
        $image_data = file_get_contents("." . $brandLogos[0] ?? '');
        header('Content-Type: image/jpeg');
        exit($image_data);
    }

    /**
     * @Apidoc\Title("获取登录验证码")
     * @Apidoc\Desc("直接用于 img 的src属性")
     * @Apidoc\Method("get")
     * @Apidoc\Url("/api/admin/passport/captcha")
     * @Apidoc\Query("v", type="int", require=true, default="1", desc="版本： 0=以前的验证方式并将弃用，1新版本使用")
     * @Apidoc\Returned("sign", type="string", desc="当前验证码的签名，登录的时候需要传到头部去，头部参数名：sign")
     * @Apidoc\Returned("base64", type="string", desc="当前验证码的base64图片")
     * NotHeaders
     */
    public function captcha()
    {
        return Captcha::create();
    }

    /**
     * @Apidoc\Title("客户端KEY")
     * @Apidoc\Desc("获取客户端KEY，用于加密数据发送给服务端")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/passport/getKey")
     * @Apidoc\Param("client_id", type="int", require=true, default="dk1h278t12e2ty7g1bh2", desc="客户端ID（希望不变的，除非清除浏览器缓存或者卸载应用）")
     * @Apidoc\Returned("type", type="string", desc="类型")
     * @Apidoc\Returned("id", type="string", desc="客户端ID")
     * @Apidoc\Returned("key", type="string", desc="public_key")
     * NotHeaders
     */
    public function getKey($client_id)
    {
        return $this->renderSuccess('', KeypairService::getKey($client_id));
    }

    /**
     * @Apidoc\Title("后台登录")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/api/admin/passport/login")
     * @Apidoc\Param("username", type="string", require=true, default="admin", desc="用户名")
     * @Apidoc\Param("password", type="string", require=true, default="666666", desc="密码")
     * @Apidoc\Param("code", type="int", require=true, default="123456", desc="验证码")
     * @Apidoc\Returned("user_name", type="string", desc="用户名")
     * @Apidoc\Returned("token", type="string", desc="token")
     * @Apidoc\After(event="setGlobalHeader",key="Token",value="res.data.data.token",desc="我的全局Header参数")
     */
    public function login(PassportValidate $validate)
    {
        $data = $validate->goCheck('login');
        //
        if (!Captcha::check($data['code']) && !(env('APP_DEBUG') == true && $data['code'] == 123456)) {
            return $this->renderError('验证码错误');
        }
        //
        $model = new UserModel;
        if ($user = $model->login($data)) {
            return $this->renderSuccess('登录成功', [
                'user_name' => $user['user_name'],
                'token' => $user['token']
            ]);
        }
        //
        return $this->renderError($model->getError() ?: '用户名或者密码错误！');
    }

    /**
     * @Apidoc\Title("修改密码")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/passport/renew")
     * @Apidoc\Param("pass", type="string", default="", desc="密码")
     * @Apidoc\Param("checkPass", type="string", default="", desc="确认密码")
     * @Apidoc\Param("oldPass", type="string", default="", desc="原密码")
     */
    public function renew(AdminUserValidate $validate)
    {
        $param = $validate->goCheck('pass');
        /** @var AdminUserModel $model */
        $model = AdminUserModel::detail($this->admin['user']['admin_user_id']);
        if ($model->renew($param)) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("退出登录")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/api/admin/passport/logout")
     */
    public function logout()
    {
        return $this->renderSuccess('退出成功');
    }
}
