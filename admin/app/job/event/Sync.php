<?php

namespace app\job\event;

use think\facade\Cache;
use Workerman\Lib\Timer;
use app\common\enum\sync\SyncEnum;
use app\common\service\sync\SyncService;

/**
 * 同步
 */
class Sync
{
    /**
     * 执行函数
     */
    public function handle()
    {
        //
        if (!Cache::get('__SYNC_INDEX__')) {
            //
            Cache::set('__SYNC_INDEX__', 1);
            // 同步基础数据相关
            Timer::add(env('AYNC_TIME', 60), function () {
                $syncService = new SyncService();
                // 同步云端基础信息
                try {
                    $syncService->syncBaseInfo();
                } catch (\Throwable $th) {
                    log_write('Sync TASK Error: sync_base :' . $th, 'task');
                }
            });
        }
    }
}
