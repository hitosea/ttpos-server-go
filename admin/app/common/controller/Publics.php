<?php
namespace app\common\controller;

use think\Request;
use app\common\tasks\Task;
use app\controller as BaseController;

/**
 * 公共类接口
 * @Apidoc\Sort(2)
 * @Apidoc\Group("home")
 */
class Publics extends BaseController
{
    /**
     * @Apidoc\Title("asynTask")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/api/common/publics/asynTask")
     * @Apidoc\Param("x", type="", require=true, default="2", desc="倍数")
     * @Apidoc\Returned()
     */
    public function asynTask(Request $request)
    {
        fastcgi_finish_request(); /* 响应完成, 关闭连接 */
        // 执行任务
        (new Task)->starts($request);
    }
}
