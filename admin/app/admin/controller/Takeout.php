<?php

namespace app\admin\controller;

use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use help\HttpHelp;

/**
 * 外卖日志管理
 * @Apidoc\Group("base")
 * @Apidoc\Sort(8)
 */
class Takeout extends Controller
{
    /**
     * @Apidoc\Title("外卖导入日志列表")
     * @Apidoc\Desc("分页查询外卖导入日志，支持按平台、类型、状态筛选")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/takeout/logs")
     * @Apidoc\Query("platform", type="string", require=false, default="", desc="外卖平台(grab/lineman等)，为空查询所有")
     * @Apidoc\Query("import_type", type="int", require=false, default=0, desc="导入类型(1-TTPOS推送到平台 2-平台推送到TTPOS)，0 查询所有")
     * @Apidoc\Query("status", type="int", require=false, default=-1, desc="导入状态(0-进行中 1-成功 2-失败)，-1 查询所有")
     * @Apidoc\Query("page", type="int", require=false, default=1, desc="页码，默认 1")
     * @Apidoc\Query("page_size", type="int", require=false, default=20, desc="每页数量，默认 20，最大 100")
     * @Apidoc\Query("company_uuid", type="string", require=false, default="", desc="商户 UUID(平台管理员可指定，商户管理员只能查自己的)")
     * @Apidoc\Returned("list", type="array", desc="日志列表", children={
     *      @Apidoc\Returned("uuid", type="int", desc="日志UUID"),
     *      @Apidoc\Returned("platform", type="string", desc="外卖平台"),
     *      @Apidoc\Returned("import_type", type="int", desc="导入类型"),
     *      @Apidoc\Returned("import_direction", type="string", desc="导入方向描述"),
     *      @Apidoc\Returned("status", type="int", desc="导入状态"),
     *      @Apidoc\Returned("progress", type="int", desc="进度百分比"),
     *      @Apidoc\Returned("success_count", type="int", desc="成功数量"),
     *      @Apidoc\Returned("failure_count", type="int", desc="失败数量"),
     *      @Apidoc\Returned("total_count", type="int", desc="总数量"),
     *      @Apidoc\Returned("error_message", type="string", desc="错误信息"),
     *      @Apidoc\Returned("start_time", type="int", desc="开始时间"),
     *      @Apidoc\Returned("end_time", type="int", desc="结束时间"),
     *      @Apidoc\Returned("duration", type="int", desc="耗时(秒)"),
     *      @Apidoc\Returned("createtime", type="int", desc="创建时间"),
     * })
     * @Apidoc\Returned("total", type="int", desc="总记录数")
     * @Apidoc\Returned("page", type="int", desc="当前页码")
     * @Apidoc\Returned("page_size", type="int", desc="每页数量")
     */
    public function logs()
    {
        
        if (!$this->getData('company_uuid')) {
            return $this->renderError('商户UUID不能为空');
        }

        // 获取请求参数
        $params = $this->getData();
        
        // 构建查询参数
        $queryParams = http_build_query([
            'platform' => $params['platform'] ?? '',
            'import_type' => $params['import_type'] ?? 0,
            'status' => $params['status'] ?? -1,
            'page' => $params['page'] ?? 1,
            'page_size' => $params['page_size'] ?? 20,
            'company_uuid' => $params['company_uuid'] ?? '',
        ]);

        // 调用 Go Main 接口
        $url = 'http://nginx/api/v1/admin/takeout/logs?' . $queryParams;
        $res = HttpHelp::getRequest($url, [], [
            'X-API-KEY: ' . env('JWT_SECRET'),
            'Accept-Language: ' . request()->header('language'),
        ]);

        if (!$res) {
            dump($res);
            return $this->renderError('请求失败');
        }

        $res = json_decode($res, true);
        if ($res['code'] != 0) {
            return $this->renderError($res['message']);
        }

        return $this->renderSuccess('', $res['data']);
    }
}

