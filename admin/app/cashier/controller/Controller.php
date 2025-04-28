<?php

namespace app\cashier\controller;

use think\facade\Cache;
use app\common\enum\http\StatusCode;
use app\controller as BaseController;
use app\common\exception\BaseException;
use app\common\enum\settings\SettingEnum;
use app\cashier\model\cashier\User as UserModel;
use app\common\model\settings\Setting as SettingModel;

/**
 * 商户后台控制器基类
 */
class Controller extends BaseController
{
    /** @var array $store 登录信息 */
    protected $cashier;

    /** @var string $route 当前控制器名称 */
    protected $controller = '';

    /** @var string $route 当前方法名称 */
    protected $action = '';

    /** @var string $route 当前路由uri */
    protected $routeUri = '';

    /**
     * 后台初始化
     */
    public function initialize()
    {
        // 当前路由信息
        $this->getRouteInfo();
        // 验证状态
        $this->checkAuth();
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
        // 当前uri
        $this->routeUri = '/' . $this->controller . '/' . $this->action;
    }

    /**
     * 验证登录状态
     */
    private function checkAuth()
    {
        //
        $token = request()->header('token');
        if (!$token) {
            throw new BaseException(['msg' => '登录失效', 'code' => StatusCode::TOKEN_ERROR]);
        }
        //
        $data = checkToken($token, 'cashier');
        if ($data['code'] != 1) {
            throw new BaseException(['msg' => $data['msg'], 'code' => StatusCode::TOKEN_ERROR]);
        }
        if ($data['data']['type'] != 'cashier') {
            throw new BaseException(['msg' => '用户信息错误', 'code' => StatusCode::USER_ERROR]);
        }
        request()->appId = $data['data']['company_uuid'];
        //
        if (!$user = UserModel::getUser($data['data'])) {
            throw new BaseException(['msg' => '没有找到用户信息', 'code' => StatusCode::USER_ERROR]);
        }
        // 设置id
        request()->appId = $data['data']['company_uuid'];
        request()->cashier_id = $user['uuid'];
        //
        if (!$cashier = Cache::get('cashier_user_info' . $token)) {
            // 商家后台设置的名称
            $shop = SettingModel::getSupplierItem(SettingEnum::STORE, $user['shop_supplier_id'] ?? 0, $user['app_id'] ?? 0);
            // 权限
            $cashier = [
                'device_id' => $data['data']['device_id'] ?? '',
                'user' => [
                    'cashier_id' => $user['uuid'],
                    'user_name' => $user['user_name'],
                    'real_name' => $user['real_name'] ?? $user['user_name'] ?? '',
                    'account' => $user['user_name'],
                    'mobile' => $user['mobile'],
                    'is_super' => $user['is_super'],
                    'name' => $shop['name'],
                    'time_zone' => $shop['time_zone'],
                    'duty_no' => $user['duty_no'] ?? '', // 当班编号
                ],
                'app' => $user['app']->toArray(),
                'supplier' => $user['supplier'],
            ];
            Cache::tag('cashier')->set('cashier_user_info' . $token, $cashier, 3600);
        }
        // 设置时区
        $setting = SettingModel::getSupplierItem(SettingEnum::STORE, request()->shopSupplierId);
        if ($timezone = ($setting['time_zone'] ?? '')) {
            date_default_timezone_set($timezone);
        }
        // 验证设备是否绑定
        request()->userInfo = $this->cashier = $cashier;
        //
        return true;
    }

    /**
     * 获取供应商id
     */
    protected function getCashierId()
    {
        return $this->cashier['user']['cashier_id'];
    }

}
