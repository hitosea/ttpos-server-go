<?php

namespace app\shop\controller;

use help\SystemHelp;
use think\facade\Env;
use app\admin\model\CompanyStaff;
use app\common\model\shop\OptLog;
use app\shop\service\AuthService;
use app\common\enum\http\StatusCode;
use app\controller as BaseController;
use app\common\exception\BaseException;
use app\common\enum\settings\SettingEnum;
use app\shop\model\shop\User as UserModel;
use app\common\model\settings\Setting as SettingModel;

/**
 * 商户后台控制器基类
 */
class Controller extends BaseController
{
    /** @var array $store 商家登录信息 */
    protected $store;

    /** @var string $route 当前控制器名称 */
    protected $controller = '';

    /** @var string $route 当前方法名称 */
    protected $action = '';

    /** @var string $route 当前路由uri */
    protected $routeUri = '';

    /** @var string $route 当前路由：分组名称 */
    protected $group = '';

    /** @var string $route 当前路由：分组名称 */
    protected $menu = '';
    /** @var array $allowAllAction 登录验证白名单 */
    protected $allowAllAction = [
        '/passport/login',              // 登录页面
        '/passport/captcha',            // 登录验证码
        '/passport/getKey',             // 加密key
        // 语言列表
        '/index/public',
        '/index/authCode',
        '/index/getNewVersion',
    ];

    /**
     * 后台初始化
     */
    public function initialize()
    {
        // 当前路由信息
        $this->getRouteInfo();
        //  验证登录状态
        $this->checkLogin();
        // 写入操作日志
        $this->saveOptLog();
        // 验证当前页面权限
        $this->checkPrivilege();
    }

    /**
     * 验证当前页面权限
     */
    private function checkPrivilege()
    {
        if ($this->store == null) {
            return false;
        }
        $authService = new AuthService($this->store);
        // 把uri转换成小写，兼容旧版本请求不区分大小写
        $this->routeUri = strtolower($this->routeUri);
        if (!$authService->checkPrivilege($this->routeUri)) {
            throw new BaseException(['msg' => '当前没有权限使用此功能']);
        }
        return true;
    }

    /**
     * 解析当前路由参数 （分组名称、控制器名称、方法名）
     */
    protected function getRouteInfo()
    {
        // 控制器名称
        $this->controller = strtolower($this->request->controller());
        $this->controller = str_replace(".", "/", $this->controller);
        // 方法名称
        $this->action = Request()->action();
        // 控制器分组 (用于定义所属模块)
        $groupstr = strstr($this->controller, '.', true);
        $this->group = $groupstr !== false ? $groupstr : $this->controller;
        // 当前uri
        $this->routeUri = '/' . $this->controller . '/' . $this->action;
    }

    /**
     * 验证登录状态
     */
    private function checkLogin()
    { 
        // 验证当前请求是否在白名单
        if (in_array($this->routeUri, $this->allowAllAction)) {
            return true;
        }
        $token = request()->header('token');
        if (!$token) {
            $token = Request()->param('token');
        }
        if (!$token) {
            throw new BaseException(['msg' => '缺少必要的参数：token', 'code' => StatusCode::USER_ERROR]);
        }
        $data = checkToken($token, 'shop');
        // 验证登录状态
        if ($data['code'] != 1) {
            throw new BaseException(['msg' => $data['msg'], 'code' => StatusCode::USER_ERROR]);
        }
        if ($data['data']['type'] != 'shop') {
            throw new BaseException(['msg' => '用户信息错误', 'code' => StatusCode::USER_ERROR]);
        }
        // 集团检测
        if (!$companyUuid = (new CompanyStaff([], 0))->setAppId(0)->where('uuid', $data['data']['uid'] ?? 0)->value('company_uuid')) {
            throw new BaseException(['msg' => '没有找到用户信息', 'code' => StatusCode::USER_ERROR]);
        }
        request()->appId = $companyUuid;
        //
        if (!$user = UserModel::getUser($data['data'])) {
            throw new BaseException(['msg' => '没有找到用户信息', 'code' => StatusCode::USER_ERROR]);
        }
        if ($user->delete_time > 0) {
            throw new BaseException(['msg' => '账号已删除，请联系管理员', 'code' => StatusCode::USER_ERROR]);
        }
        if ($user->is_disable == 1) {
            throw new BaseException(['msg' => '账号被禁用，请联系管理员', 'code' => StatusCode::USER_ERROR]);
        }
        if (($data['data']['pwd'] ?? '') != md5($user->password)) {
            throw new BaseException(['msg' => '密码已变更，请重新登录', 'code' => StatusCode::USER_ERROR]);
        }
        if (!$user->app) {
            throw new BaseException(['msg' => '店铺状态异常，请联系销售代表', 'code' => StatusCode::USER_ERROR]);
        }
        // 授权
        request()->licenses = $user->app->getLicense();
        // 保存登录状态
        request()->userInfo = $this->store = [
            'user' => [
                'uuid' => $user['uuid'],
                'company_uuid' => $companyUuid,
                'user_name' => $user['username'],
                'real_name' => $user['real_name'],
                'user_type' => $user['user_type'],
                'shop_user_id' => $user['uuid'], // 保留，兼容旧版本
                'shop_supplier_id' => $companyUuid, // 保留，兼容旧版本
            ],
            'app' => $user->app,
            'supplier' => $user->supplier()->field('
                company_uuid as app_id, company_uuid, sale_stock, cash_limit, kitchen_limit, tablet_limit, assistant_limit, 
                is_open_member, is_open_tablet, is_open_coupon, is_open_marketing, languages,
                is_open_h5 as is_open_scan, is_open_assistant, is_open_kitchen_kds, is_open_buffet, table_limit, printer_limit, timezone, delivery_status,
                enable_grab_delivery, enable_lineman_delivery,
                enable_kiosk
            ')->find(),
        ];
        // 设置时区
        $setting = SettingModel::getSupplierItem(SettingEnum::STORE, request()->shopSupplierId);
        if ($timezone = ($setting['time_zone'] ?? '')) {
            date_default_timezone_set($timezone);
        }
        //
        return true;
    }

    /**
     * 操作日志
     */
    private function saveOptLog()
    {
        // 过滤循环请求
        $allowLoopUrl = [
            '/auth/loginlog/index',
            '/auth/optlog/index',
            '/auth/role/index'
        ];
        if (in_array($this->routeUri, $allowLoopUrl)) {
            return;
        }
        if (Env::get('env') == 'uat') {
            return;
        }
        if ($this->store == null) {
            return;
        }
        // 如果不记录查询日志
        //
        try {
            $title = (new AuthService($this->store))::getAccessNameByApiPath($this->routeUri, $this->store['app']['app_id']);
            if ($title) {
                (new OptLog)->save([
                    'uuid' => createUuid(),
                    'staff_uuid' => $this->store['user']['uuid'] ?? 0,
                    'ip' => \request()->ip(),
                    'url' => $this->routeUri,
                    'type' => $this->request->isGet() ? 'Get' : 'Post',
                    'request_data' => json_encode($this->request->param(), JSON_UNESCAPED_UNICODE),
                    'source' => SystemHelp::getClientBrowser(),
                    'agent' => $_SERVER['HTTP_USER_AGENT'],
                    'title' => $title,
                ]);
            }
        } catch (\Throwable $th) {
            //throw $th;
        }
    }
}
