<?php

namespace app\job\event;

use think\facade\Cache;
use Workerman\Lib\Timer;
use app\common\library\printer\Driver;
use app\common\model\settings\PrinterLog;

/**
 * 打印
 */
class ShopPrint
{

    /**
     * 执行函数
     *
     * 该函数用于执行打印任务。
     * 如果缓存中不存在 "__SHOP_PRINT_INDEX__" 键，则设置该键的值为 1，并创建一个定时器，每秒钟执行一次打印操作。
     * 打印操作会查询数据库中满足条件的打印日志记录，并逐个进行打印。
     * 打印成功后，更新打印日志的状态和打印次数。
     * 如果打印失败，则根据打印次数判断是否继续尝试打印。
     *
     * @return bool 返回执行结果，始终为 true。
     */
    public function handle()
    {
        $appId = request()->appId;
        // 禁止并发操作
        if (Cache::get('__SYSTEM_JOB_SHOP_PRINT__' . $appId)) {
            return;
        }
        Cache::set('__SYSTEM_JOB_SHOP_PRINT__' . $appId, 1, 10);
        //
        try {
            //
            $printerLogs = (new PrinterLog([], $appId))
                ->whereRaw('((printer_time + 10) < UNIX_TIMESTAMP() OR num = 0)')
                ->where('status', 1)
                ->where('first_execution', 0)
                ->where('type', 0)
                ->order('create_time, num')
                ->group('printer_id')
                ->limit(10)
                ->select();
            foreach ($printerLogs as $printerLog) {
                $printerDriver = $printerLog->printer ?? null;
                $num = $printerLog->num + 1;
                if ($printerDriver) {
                    $driver = new Driver($printerDriver);
                    $result = $driver->printTicket($printerLog->data ?? '', $printerLog->print_method);
                    if ($result) {
                        $printerLog->reason = "打印成功";
                        $printerLog->status = 2;
                    } else {
                        $printerLog->status = $num >= 5 ? 0 : 1;
                        if ($printerLog->status == 0) {
                            $printerLog->reason = $driver->getError() ?: '打印失败，未连接打印机';
                        }
                    }
                } else {
                    $printerLog->reason = "打印机不存在";
                    $printerLog->status = 0;
                }
                //
                $printerLog->printer_time = time();
                $printerLog->num = $num;
                $printerLog->save();
            }
            //
            Cache::set('__SYSTEM_JOB_SHOP_PRINT__' . $appId, 0, 0);
            //
        } catch (\Throwable) {
            Cache::set('__SYSTEM_JOB_SHOP_PRINT__' . $appId, 0, 0);
        }
    }
}
