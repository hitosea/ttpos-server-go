<?php

namespace app\admin\service;

use app\common\model\admin\Access;
use think\facade\Cache;
use app\admin\model\admin\User;
use app\admin\model\admin\UserRole;
use app\admin\model\admin\RoleAccess;

/**
 * 商家后台权限业务
 */
class AuthService
{
    // 存放实例
    static public $instance;

    // 商家登录信息
    private $admin;

    // 商家用户信息
    private $user;

    // 权限验证白名单
    protected $allowAllAction = [
        // 用户登录
        '/admin/passport/login',
        // 退出登录
        '/admin/passport/logout',
        // 首页
        '/admin/index/index',
        /*登录信息*/
        '/admin/index/base',
        // 图片上传
        '/admin/file/file/*',
        '/admin/file._upload/image',
        // 用户信息
        '/admin/user/getUserInfo',
        // 角色列表
        '/admin/user/getRoleList',
        // 修改密码
        '/admin/passport/renew',
        //
        '/admin/access/getRoleList',
        //
    ];

    /** @var array $accessUrls 商家用户权限url */
    private $accessUrls = [];

    /**
     * 私有化构造方法
     */
    public function __construct($admin)
    {
        // 商家登录信息
        $this->admin = $admin;
        // 当前用户信息
        $this->user = User::detail($this->admin['user']['admin_user_id']);
        // 把白名单uri转换成小写，兼容旧版本请求不区分大小写
        $allowAllAction = array_map('strtolower', $this->allowAllAction);
        //
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
        if (!is_array($url)):
            return $this->checkAccess($url);
        else:
            foreach ($url as $val):
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
        $urls = str_replace('._', '.', $url);

        // 超级管理员无需验证
        if ($this->user && $this->user['is_super'] ?? '') {
            return true;
        }

        // 验证当前请求是否在白名单
        if (in_array(strtolower($url), $this->allowAllAction)) {
            return true;
        }
        if ($urls == '/admin/setting.service/index' && request()->isGet()) {
            return true;
        }

        // 不需要权限管理
        if (!Access::where('api_path', $url)->whereOr('api_path', strtolower($url))->whereOr('api_path', $urls)->value('id')) {
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
        $accessUrls = $this->getAccessUrls();
        if (!in_array(strtolower($url), $accessUrls) && !in_array($url, $accessUrls) && !in_array($urls, $accessUrls)) {
            return false;
        }

        return true;
    }

    /**
     * 获取当前用户的权限url列表
     */
    private function getAccessUrls()
    {
        if (empty($this->accessUrls)) {
            // 获取当前用户的角色集
            $roleIds = UserRole::getRoleIds($this->user['admin_user_id']);
            // 根据已分配的权限
            $accessIds = RoleAccess::getAccessIds($roleIds);
            //
            // 获取当前角色所有权限链接
            $this->accessUrls = Access::getApiAccessUrls($accessIds);
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
    public static function getAccessNameByApiPath($path)
    {
        $arr = Cache::get('{cache}_' . 'admin_api_path_access');
        if (!$arr) {
            // 查找访问资源
            $list = (new Access())->withoutGlobalScope()->field(['name', 'api_path'])->select();
            foreach ($list as $access) {
                $arr[$access['api_path']] = $access['name'];
            }
            Cache::tag('cache')->set('{cache}_' . 'admin_api_path_access', $arr, 3600);
        }
        $path = str_replace('._', '.', $path);
        $url = isset($arr[$path]) ? $arr[$path] : '';
        return $url;
    }
}
