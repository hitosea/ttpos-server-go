<?php

namespace app\shop\controller;

use captcha\facade\Captcha;
use app\shop\model\shop\User;
use app\admin\model\CompanyStaff;
use hg\apidoc\annotation as Apidoc;
use app\common\enum\http\StatusCode;
use app\common\enum\settings\SettingEnum;
use app\common\service\keypair\KeypairService;
use app\shop\model\settings\Setting as SettingModel;

/**
 * 商户认证
 * @Apidoc\Group("home")
 * @Apidoc\Sort(0)
 */
class Passport extends Controller
{
    /**
     * @Apidoc\Title("获取登录验证码")
     * @Apidoc\Desc("直接用于 img 的src属性")
     * @Apidoc\Method("get")
     * @Apidoc\Url("/index.php/shop/passport/captcha")
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
     * @Apidoc\Url("/api/shop/passport/getKey")
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
     * @Apidoc\Title("商户后台登录")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/shop/passport/login")
     * @Apidoc\Param("username", type="string", require=true, default="000000", desc="用户名")
     * @Apidoc\Param("password", type="string", require=true, default="666666", desc="密码")
     * @Apidoc\Param("code", type="int", require=true, default="123456", desc="验证码")
     * @Apidoc\After(event="setGlobalHeader",key="Token",value="res.data.data.token",desc="我的全局Header参数")
     */
    public function login()
    {
        $user = $this->postData();
        // 不再预加密密码，将明文密码传递给 checkLogin 方法
        // checkLogin 方法内部会使用 verify_password() 处理 MD5 和 bcrypt 两种格式
        // 云上部署时，判断规则不同
        if (!Captcha::check($user['code']) && !(env('APP_DEBUG') == true && $user['code'] == 123456)) {
            return $this->renderError('验证码错误');
        }

        // 集团检测
        $uuid = (new CompanyStaff([], 0))->setAppId(0)->checkExistInCompany($user['username']);
        if (!$uuid) {
            return $this->renderError('用户不存在');
        }
        request()->appId = $uuid;

        $model = new User();
        if ($userInfo = $model->checkLogin($user)) {
            // saas首次修改密码
            if ($userInfo['password_change_count'] == 0) {
                return $this->renderError('首次登录需修改密码', ['token' => $userInfo['token']], StatusCode::UNBIND_ERROR);
            }
            $companyUuid = $userInfo['company_uuid'] ?? 0;
            $setting = SettingModel::getSupplierItem(SettingEnum::STORE, $companyUuid, $companyUuid);
            //
            return $this->renderSuccess('登录成功', [
                'app_id' => $companyUuid,
                'user_name' => $userInfo['username'],
                'token' => $userInfo['token'],
                'shop_name' => $setting['name'],
                'supplier_name' => $userInfo['app']['name'],
                'user_type' => $userInfo['user_type'],
                'version' => get_version(),
                'logoUrl' => $setting['logoUrl'],
                'currency' => $userInfo['currency'],
            ]);
        }
        return $this->renderError($model->getError() ?: '登录失败');
    }

    /**
     * @Apidoc\Title("退出登录")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/shop/passport/logout")
     */
    public function logout()
    {
        return $this->renderSuccess('退出成功');
    }


    /**
     * @Apidoc\Title("修改密码")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/shop/passport/editPass")
     * @Apidoc\Param("oldpass", type="string", require=true, default="", desc="旧密码")
     * @Apidoc\Param("password", type="string", require=true, default="", desc="新密码")
     * @Apidoc\Param("confirmPass", type="string", require=true, default="", desc="确认密码")
     */
    public function editPass()
    {
        $model = new User();
        if ($model->editPass($this->postData(), $this->store['user'])) {
            return $this->renderSuccess('修改成功');
        }
        return $this->renderError($model->getError() ?: '修改失败');
    }

    /**
     * @Apidoc\Title("SAAS-首次修改密码")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/shop/passport/saasEditPassword")
     * @Apidoc\Param("username", type="string", require=true, default="", desc="用户名")
     * @Apidoc\Param("old_password", type="string", require=true, default="", desc="旧密码")
     * @Apidoc\Param("password", type="string", require=true, default="", desc="新密码")
     * @Apidoc\Param("confirm_password", type="string", require=true, default="", desc="确认密码")
     */
    public function saasEditPassword()
    {
        $data = $this->postData();
        $model = new User();
        if ($model->saasEditPassword($data, $this->store['user'])) {
            return $this->renderSuccess('修改成功');
        }
        return $this->renderError($model->getError() ?: '修改失败');
    }
}
