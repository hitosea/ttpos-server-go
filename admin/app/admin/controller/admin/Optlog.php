<?php

namespace app\admin\controller\admin;

use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use app\common\model\admin\OptLog as OptLogModel;

/**
 * 操作日志
 * @Apidoc\Group("user")
 * @Apidoc\Sort(4)
 */
class Optlog extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/api/admin/admin.optlog/index")
     * @Apidoc\Param("username", type="string", require=true, default="", desc="用户名")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\common\model\admin\OptLog\getList", desc="操作日志列表")
     */
    public function index()
    {
        $model = new OptLogModel;
        $list = $model->getList($this->postData());
        return $this->renderSuccess('', compact('list'));
    }
}
