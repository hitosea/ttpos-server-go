<?php

namespace app\scan\controller;

use app\common\model\app\App;
use app\common\model\store\Table;
use app\common\enum\http\StatusCode;
use app\controller as BaseController;
use app\common\exception\BaseException;
use app\common\service\qrcode\AuthService;
use app\common\enum\settings\SettingEnum;
use app\common\model\settings\Setting as SettingModel;


/**
 * API控制器基类
 */
class Controller extends BaseController
{

    // app_id
    protected int $app_id;

    // 桌台
    protected array $table;

    /** @var string $route 当前控制器名称 */
    protected string $controller = '';

    /** @var string $route 当前方法名称 */
    protected string $action = '';

    /** @var string $route 当前路由uri */
    protected string $routeUri = '';

    /** @var array $allowAllAction 验证白名单 */
    protected array $allowAllAction = [];

    /**
     * 后台初始化
     */
    public function initialize()
    {
        // 当前路由信息
        $this->getRouteInfo();
        $this->check();
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
        $this->action = request()->action();
        // 当前uri
        $this->routeUri = '/' . $this->controller . '/' . $this->action;
    }

    /**
     * 验证状态
     * @throws BaseException
     */
    private function check()
    {
        $token = request()->header('token');
        $auth = (new AuthService);
        $token = $auth->decodeToken($token);
        if (!$token) {
            throw new BaseException(['msg' => '访问失效', 'code' => StatusCode::VISIT_ERROR]);
        }
        if (!isset($token['s']) || !isset($token['t']) || !isset($token['a']) || !isset($token['q'])) {
            throw new BaseException(['msg' => '访问失效', 'code' => StatusCode::VISIT_ERROR]);
        }
        // 设置id
        request()->appId = $appId = $token['a'];
        request()->shopSupplierId = $token['s'];
        //
        $table = Table::detail($token['t']);
        if (!$table || $table->qrcode_value != ($token['q'] ?: '')) {
            throw new BaseException(['msg' => '访问失效', 'code' => StatusCode::VISIT_ERROR]);
        }
        if (!$table || $table?->switch_status != 1) {
            throw new BaseException(['msg' => '访问失效', 'code' => StatusCode::VISIT_ERROR]);
        }
        //
        request()->licenses = $license = App::where('uuid', $appId)->find()?->getLicense();
        //
        if (($license['is_open_scan'] ?? 0) != 1) {
            throw new BaseException(['msg' => '当前未开启扫码点餐功能，请联系销售代表', 'data' => [], 'code' => StatusCode::USER_ERROR]);
        }
        // 系统设置状态
        $settingData = SettingModel::getAll(request()->appId ?? 0, request()->shopSupplierId ?? 0);
        $cashier = $settingData[SettingEnum::CASHIER]['values'] ?? [];
        if ($cashier['order_method']['is_table_order'] != 1) {
            throw new BaseException(['msg' => '桌台已关闭', 'code' => StatusCode::VISIT_ERROR]);
        }
        //
        request()->userInfo = $this->table = [
            'shop_supplier_id' => $token['s'] ?? 0,
            'table_id' => $token['t'] ?? 0,
            'app_id' => $token['a'] ?? 0,
            'qrcode_value' => $token['q'] ?? 0,
            'setting_data' => $settingData,
        ];
        // 设置时区
        $setting = SettingModel::getSupplierItem(SettingEnum::STORE, request()->shopSupplierId);
        if ($timezone = ($setting['time_zone'] ?? '')) {
            date_default_timezone_set($timezone);
        }
        return true;
    }
}
