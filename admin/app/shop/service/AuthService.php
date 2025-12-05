<?php

namespace app\shop\service;

use think\facade\Cache;
use think\facade\Session;
use app\shop\model\auth\User;
use app\common\model\shop\Access;
use app\shop\model\auth\UserRole;
use app\shop\model\auth\RoleAccess;

/**
 * 商家后台权限业务
 */
class AuthService
{
    // 存放实例
    static public $instance;

    // 商家登录信息
    private $store;

    // 商家用户信息
    private $user;

    // 权限验证白名单
    protected $allowAllAction = [
        // 首页
        '/index/index',
        '/index/base',
        '/index/authCode',
        '/index/verifyAuthCode',
        '/index/bindList',
        '/index/unbind',
        '/index/aiTranslate',
        '/index/checkNameExist',
        '/index/productList',

        // 用户登录
        '/passport/login',
        '/passport/logout',
        '/passport/editPass',
        '/passport/saasEditPassword',
        '/user/survey/rechargeRank',
        '/user/survey/consumerRank',

        // 图片上传
        '/file/file/*',
        '/file/upload/image',
        '/file/upload/batchReplaceProductImage',

        // 数据选择
        '/data/*',

        // 商品分类
        '/product/store/category/parent',
        '/product/buffet/buffet/customerList', // 自助餐
        '/product/spec/*', // 添加商品规格
        '/product/tax/tax/list', // 税率列表
        '/product/store/product/imports', // 商品导入
        '/product/expand/spec/skuProduct', // 规格商品修改价格
        '/product/store/product/batchUpdateTax', // 批量修改税率
        '/product/store/product/batchUpdateCategory', // 批量修改分类
        '/product/store/product/batchUpdateOpenOverallDiscount', // 批量修改整单折扣

        // 订单
        '/store/operate/orderProductList', // 订单商品列表
        '/store/operate/orderRefund',   // 订单退款
        '/store/operate/orderRefundAgain',   // 订单再次退款

        // 用户信息
        '/auth/user/getUserInfo',
        '/auth/user/getRoleList', // 角色列表
        '/auth/user/index',

        // 统计
        '/statistics/sales/order',
        '/statistics/sales/product',
        '/statistics/user/scale',
        '/statistics/user/new_user',
        '/statistics/user/pay_user',

        // 各端设置
        '/setting/terminal/editCashierAdvancedPassword',
        '/setting/terminal/editCashierPassword',
        '/setting/terminal/editAdvancedPassword',
        '/setting/terminal/editKitchenAdvancedPassword',
        '/setting/terminal/editAssistantAdvancedPassword',
        '/setting/terminal/editAssistantLockPassword',
        '/setting/paytype/defaultPay',
        '/setting/business/freeTag',
        '/setting/business/returnReason',
        '/setting/business/orderRemark',
        '/setting/business/orderItemRemark',
        '/setting/business/qrcode',
        '/setting/business/companyQrcode',

        // 采购管理
        '/purchase/ErpPurchaseOrder/adjust',
        '/purchase/ErpSupplier/list',

        // 库存管理
        '/inventory/erpinventory/list',

        // 门店管理
        '/supplier/printerTemplate/setTemplate',

        // 添加会员获取会员卡列表
        '/card/card/getCardList',

        // 营销活动
        '/marketing/activity/couponlist',
    ];

    /** @var array $accessUrls 商家用户权限url */
    private $accessUrls = [];

    /**
     * 公有化获取实例方法
     */
    public static function getInstance()
    {
        if (!(self::$instance instanceof AuthService)) {
            self::$instance = new self;
        }
        return self::$instance;
    }

    /**
     * 私有化构造方法
     */
    public function __construct($store)
    {
        // 商家登录信息
        $this->store = $store;
        // 当前用户信息
        $this->user = User::detail($this->store['user']['uuid']);
        // 把白名单uri转换成小写，兼容旧版本请求不区分大小写
        $allowAllAction = array_map('strtolower', $this->allowAllAction);
        $this->allowAllAction = $allowAllAction;
    }

    /**
     * 私有化克隆方法
     */
    private function __clone() {}

    /**
     * 验证指定url是否有访问权限
     */
    public function checkPrivilege($url, $strict = true)
    {
        if (!is_array($url)) :
            return $this->checkAccess($url);
        else :
            foreach ($url as $val) :
                if ($strict && !$this->checkAccess($val)) {
                    return false;
                }
                if (!$strict && $this->checkAccess($val)) {
                    return true;
                }
            endforeach;
        endif;
        return true;
    }

    /**
     * @param string $url
     */
    private function checkAccess($url)
    {
        // 超级管理员无需验证
        if ($this->user && $this->user['is_super'] ?? '') {
            return true;
        }
        if (strstr($url, 'orderold') !== false) {
            return true;
        }

        //
        $url = strtolower($url);
        $this->allowAllAction = array_map('strtolower', $this->allowAllAction);

        // 验证当前请求是否在白名单
        if (in_array($url, $this->allowAllAction)) {
            return true;
        }

        // 通配符支持
        foreach ($this->allowAllAction as $action) {
            if (
                strpos($action, '*') !== false
                && preg_match('/^' . str_replace('/', '\/', $action) . '/', $url)
            ) {
                return true;
            }
        }

        // 获取当前用户的权限url列表
        if (!in_array($url, $this->getAccessUrls())) {
            return false;
        }
        return true;
    }

    /**
     * 获取当前用户的权限url列表
     */
    private function getAccessUrls()
    {
        if ($this->user && empty($this->accessUrls)) {
            // 获取当前用户的角色集
            $roleIds = UserRole::getRoleIds($this->user['shop_user_id']);
            // 根据已分配的权限
            $accessIds = RoleAccess::getAccessIds($roleIds);
            // 获取当前角色所有权限链接
            $this->accessUrls = Access::getApiAccessUrls($accessIds, $this->user['user_type'], $this->store['supplier']);
        }
        return array_map('strtolower', $this->accessUrls);
    }

    /**
     * 前端路由权限验证
     *
     * @param [type] $path
     * @param [type] $app_id
     * @return string
     */
    public static function getAccessNameByPath($path, $app_id)
    {
        $arr = Cache::get('{cache}_' . 'path_access_' . $app_id);
        if (!$arr) {
            // 查找访问资源
            $list = (new Access())->withoutGlobalScope()->field(['name', 'path'])->select();
            foreach ($list as $access) {
                $arr[$access['path']] = $access['name'];
            }
            Cache::tag('cache')->set('{cache}_' . 'path_access_' . $app_id, $arr, 3600);
        }
        $url = isset($arr[$path]) ? $arr[$path] : '';
        return $url;
    }

    /**
     * 后端路由权限验证
     *
     * @param [type] $path
     * @param [type] $app_id
     * @return string
     */
    public static function getAccessNameByApiPath($path, $app_id)
    {
        $arr = Cache::get('{cache}_' . 'shop_api_path_access_' . $app_id);
        if (!$arr) {
            // 查找访问资源
            $list = (new Access())->withoutGlobalScope()->field(['name', 'api_path'])->select();
            foreach ($list as $access) {
                $arr[$access['api_path']] = $access['name'];
            }
            Cache::tag('cache')->set('{cache}_' . 'shop_api_path_access_' . $app_id, $arr, 3600);
        }
        $url = isset($arr[$path]) ? $arr[$path] : '';
        return $url;
    }
}
