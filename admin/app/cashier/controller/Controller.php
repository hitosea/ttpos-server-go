<?php

namespace app\cashier\controller;

use think\facade\Cache;
use app\common\enum\http\StatusCode;
use app\common\model\shop\BindRecord;
use app\controller as BaseController;
use app\common\exception\BaseException;
use app\common\enum\settings\SettingEnum;
use app\cashier\model\cashier\User as UserModel;
use app\common\model\shop\Access as AccessModel;
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

    /** @var array $allowAllAction 验证全局白名单 */
    protected $allowAllAction = [
        '/index/bind',                          // 绑定设备
    ];
    /** @var array $allowCashierAction 收银机验证白名单 */
    protected $allowCashierAction = [
        '/index/getNewVersion',                         // 版本信息
        '/passport/getKey',                             // 获取key
        '/passport/login',                              // 登录页面
        '/passport/captcha',                            // 验证码
        '/index/base',                                  // 登录信息
        '/payment/callback/lianlianCallback',           // lianlian支付回调
        '/payment/callback/lianlianRefundCallback',     // lianlian支付退款回调
    ];
    /** @var array $allowCashierOpenStatusAction 收银机操作功能名单 */
    protected $allowCashierOpenStatusAction = [
        '/order/cart/add',
        '/order/cart/delProduct',
        '/order/cart/stay',
        '/order/cart/pick',
        '/order/cart/delStay',
        '/order/cart/sendKitchen',
        '/order/cart/moveProduct',
        '/order/cart/changeMoney',
        '/order/order/buy'
    ];

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
        // 验证当前请求是否在白名单
        if (in_array($this->routeUri, $this->allowAllAction)) {
            return true;
        }
        // 验证当前请求是否在白名单
        if (in_array($this->routeUri, $this->allowCashierAction)) {
            return true;
        }
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
        //
        if (!$user = UserModel::getUser($data['data'])) {
            throw new BaseException(['msg' => '没有找到用户信息', 'code' => StatusCode::USER_ERROR]);
        }
        // 设置id
        request()->appId = $user['app_id'];
        request()->shopSupplierId = $user['shop_supplier_id'];
        request()->licenses = $user->app->getLicense();
        request()->cashier_id = $user['shop_user_id'];
        //
        if (!$cashier = Cache::get('cashier_user_info' . $token)) {
            // 商家后台设置的名称
            $shop = SettingModel::getSupplierItem(SettingEnum::STORE, $user['shop_supplier_id'] ?? 0, $user['app_id'] ?? 0);
            // 权限
            $supplier = [
                'name' => isset($user['supplier']) && $user['supplier'] ? $user['supplier']['name'] : '',
                'category_set' => isset($user['supplier']) && $user['supplier'] ? $user['supplier']['category_set'] : 10,
                'is_main' => isset($user['supplier']) && $user['supplier'] ? $user['supplier']['is_main'] : 1,
            ];
            $permission = (new AccessModel)->getPermission(AccessModel::CASHIER_ROUTE_NAME, $user, $supplier);
            $cashier = [
                'device_id' => $data['data']['device_id'] ?? '',
                'user' => [
                    'shop_user_id' => $user['shop_user_id'],
                    'cashier_id' => $user['shop_user_id'],
                    'user_name' => $user['user_name'],
                    'real_name' => $user['real_name'] ?? $user['user_name'] ?? '',
                    'account' => $user['user_name'],
                    'mobile' => $user['mobile'],
                    'is_super' => $user['is_super'],
                    'shop_supplier_id' => $user['shop_supplier_id'],
                    'name' => $shop['name'],
                    'time_zone' => $shop['time_zone'],
                    'app_id' => $user['app_id'],
                    'permission' => $permission,
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
        $device_id = $cashier['device_id'] = $data['data']['device_id'] ?? '';
        $bindRecord = BindRecord::where('key', $device_id)->where('source', BindRecord::SOURCE_CASHIER)->find();
        if (!$bindRecord) {
            throw new BaseException(['msg' => '设备已解绑，请重新绑定', 'data' => Cache::pull("un_printer_{$device_id}"), 'code' => StatusCode::UNBIND_ERROR]);
        }
        //
        request()->userInfo = $this->cashier = $cashier;
        //
        $this->checkCashierOpen();
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


    /**
     * 检查收银是否开启
     */
    public function checkCashierOpen()
    {
        $settingData = SettingModel::getAll($this->cashier['user']['app_id'] ?? 0, $this->cashier['user']['shop_supplier_id'] ?? 0);
        $cashier = $settingData[SettingEnum::CASHIER]['values'] ?? [];
        if (!$cashier['order_method']['is_cashier_order'] && in_array($this->routeUri, $this->allowCashierOpenStatusAction)) {
            throw new BaseException(['msg' => '收银用餐已关闭，请选择其他用餐方式', 'code' => StatusCode::TABLE_ERROR]);
        }
    }
}
