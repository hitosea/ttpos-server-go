<?php

namespace app\common\controller;

use think\Request;
use think\facade\Db;
use think\facade\Cache;
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

    /**
     * @Apidoc\Title("health")
     * @Apidoc\Method ("GET")
     * @Apidoc\Url("/api/common/publics/health")
     * @Apidoc\Param("token", type="string", require=true, desc="token")
     * @Apidoc\Returned()
     */
    public function health()
    {
        $token =  request()->get("token");
        if (!$token || $token != env('HEALTH_TOKEN')) {
            return "unhealthy";
        }
        $result = Db::connect()->query('SELECT 1');
        if (!$result) {
            return "unhealthy";
        } 
        if (!Cache::set($token, "healthy") || !Cache::delete($token)){
            return "unhealthy";
        } 
        return "healthy";
    }
}
