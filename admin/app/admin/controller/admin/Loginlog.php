<?php

namespace app\admin\controller\admin;

use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use app\common\model\admin\LoginLog as LoginLogModel;

/**
 * 登录日志
 * @Apidoc\Group("user")
 * @Apidoc\Sort(3)
 */
class Loginlog extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/api/admin/admin.loginlog/index")
     * @Apidoc\Param("username", type="string", require=true, default="", desc="用户名")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\common\model\admin\LoginLog\getList", desc="登录日志列表")
     */
    public function index()
    {
        $model = new LoginLogModel;
        $list = ($model)->getList($this->postData());
        return $this->renderSuccess('', compact('list'));
    }
}
