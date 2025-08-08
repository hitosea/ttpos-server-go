<?php

namespace app\admin\controller;

use help\StringHelp;
use help\SystemHelp;
use think\facade\Env;
use app\admin\service\AuthService;
use app\common\model\admin\OptLog;
use app\common\enum\http\StatusCode;
use app\controller as BaseController;
use app\common\exception\BaseException;
use app\admin\model\admin\User as UserModel;

/**
 * 商户后台控制器基类
 */
class Controller extends BaseController
{

    // 商家登录信息
    protected $admin;

    // 当前控制器名称
    protected $controller = '';

    // 当前方法名称
    protected $action = '';

    // 当前路由uri
    protected $routeUri = '';

    // 当前路由：分组名称
    protected $group = '';

    // 登录验证白名单
    protected $allowAllAction = [
        // 登录页面
        '/admin/passport/login',
        // 验证码
        '/admin/passport/captcha',
        // 退出登录
        '/admin/passport/logout',
        // 登录Logo
        '/admin/passport/logo',
        // 获取公钥
        '/admin/passport/getKey',
        // 下载客户端
        '/admin/client.client/download',
        // 基础信息
        '/admin/setting.service/info',
        // 获取客户端最新版本
        '/admin/client.client/getNewVersion',
        // 获取安装包
        '/admin/apk/download',
        '/admin/apk/preDownload',
        // 外送渠道
        '/admin/delivery/channels',
        // 可选商家列表 
        '/admin/delivery/companySelect',
        // 商家授权管理，erpnext站点编码
        '/admin/erpnext/siteCode',
        // 商家授权管理，获取site company
        '/admin/erpnext/siteCompany',
    ];

    /**
     * 后台初始化
     */
    public function initialize()
    {
        // 当前路由信息
        $this->getRouteinfo();
        // 验证登录
        $this->checkLogin();
        // 写入操作日志
        $this->saveOptLog();
        // 验证当前页面权限
        $this->checkPrivilege();
    }

    /**
     * 解析当前路由参数 （分组名称、控制器名称、方法名）
     */
    protected function getRouteinfo()
    {
        // 控制器名称
        $this->controller = StringHelp::toUnderScore(request()->controller());
        // 方法名称
        $this->action = request()->action();
        // 控制器分组 (用于定义所属模块)
        $groupstr = strstr($this->controller, '.', true);
        $this->group = $groupstr !== false ? $groupstr : $this->controller;
        // 当前uri
        $this->routeUri = '/admin/' . $this->controller . '/' . $this->action;
        //
        $this->routeUri = str_replace('._', '.', $this->routeUri);
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
        $data = checkToken($token, 'admin');
        if ($data['code'] != 1) {
            throw new BaseException(['msg' => $data['msg'], 'code' => StatusCode::USER_ERROR]);
        }
        if ($data['data']['type'] != 'admin') {
            throw new BaseException(['msg' => '用户信息错误', 'code' => StatusCode::USER_ERROR]);
        }
        if (!$user = UserModel::getUser($data['data'])) {
            throw new BaseException(['msg' => '账号已删除，请联系管理员', 'code' => StatusCode::USER_ERROR]);
        }
        if ($user->status == 0) {
            throw new BaseException(['msg' => '账号被禁用，请联系管理员', 'code' => StatusCode::USER_ERROR]);
        }
        if (($data['data']['pwd'] ?? '') !=  md5($user->password)) {
            throw new BaseException(['msg' => '密码已变更，请重新登录', 'code' => StatusCode::USER_ERROR]);
        }
        $this->admin = [
            'user' => [
                'admin_user_id' => $user['admin_user_id'],
                'user_name' => $user['username'],
            ],
        ];
        return true;
    }

    /**
     * 验证当前页面权限
     */
    private function checkPrivilege()
    {
        if ($this->admin == null) {
            return false;
        }
        $authService = new AuthService($this->admin);
        if (!$authService->checkPrivilege($this->routeUri)) {
            throw new BaseException(['msg' => '当前没有权限使用此功能']);
        }
        return true;
    }

    /**
     * 操作日志
     */
    private function saveOptLog()
    {
        // 过滤循环请求
        $allowLoopUrl = [
            '/admin/admin.loginlog/index',
            '/admin/admin.optlog/index',
            '/admin/role/index',
        ];
        if (in_array($this->routeUri, $allowLoopUrl)) {
            return;
        }
        if (Env::get('env') == 'uat') {
            return;
        }
        if ($this->admin == null) {
            return;
        }
        $admin_user_id = $this->admin['user']['admin_user_id'];
        if (!$admin_user_id) {
            return;
        }
        try {
            $title = (new AuthService($this->admin))::getAccessNameByApiPath($this->routeUri);
            if ($title) {
                (new OptLog)->save([
                    'admin_user_id' => $admin_user_id,
                    'ip' => \request()->ip(),
                    'request_type' => $this->request->isGet() ? 'Get' : 'Post',
                    'url' => $this->routeUri,
                    'content' => json_encode($this->request->param(), JSON_UNESCAPED_UNICODE),
                    'browser' => SystemHelp::getClientBrowser(),
                    'agent' => $_SERVER['HTTP_USER_AGENT'],
                    'title' => $title == '点餐助手' ? '客户端管理' : $title,
                    'create_time' => time(),
                ]);
            }
        } catch (\Throwable $th) {
            //throw $th;
        }
    }
}
