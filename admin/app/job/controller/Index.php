<?php
namespace app\job\controller;

use think\Request;
use think\facade\Db;
use app\controller as BaseController;

/**
 * 定时类接口
 */
class Index extends BaseController
{
    /**
     * @Apidoc\Title("定时任务")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/api/job/index/index")
     * @Apidoc\Returned()
     */
    public function index(Request $request)
    {
        if($request->param('app_secret') != md5(env('APP_ID'))) {
            return '';
        }
        $type = $request->param('type');
        //
        if ($type == 'JobScheduler') {
            $appIds = Db::name('company')->where('is_enable_erp', 0)->column('uuid');
            foreach ($appIds as $appId) {
               request()->appId = $appId;
               event('JobScheduler');
            }
        }
        //
        if ($type == 'ShopPrint') {
            request()->appId = $request->param('app_id');
            event('ShopPrint');
        }
        return '';
    }
}
