<?php
namespace app\job\queue;

use think\queue\Job;
use think\facade\Log;
use app\common\library\printer\Driver;
use app\common\model\settings\PrinterLog;

class PrintJob extends BaseJob
{
    /**
     * 处理打印任务
     */
    protected function handle(Job $job, $data)
    {
        $printerLog = (new PrinterLog([], $data['app_id']))->where('id', $data['id'])->find();
        if (!$printerLog || $printerLog->status != 1) {
            Log::error('打印队列任务执行失败, 找不到打印日志: ' . $data['id']);
            return false;
        }
        if (!$printerLog->data) {
            Log::error('打印队列任务执行失败, 找不到打印数据: ' . $data['id']);
            return false;
        }
        // 
        $num = $printerLog->num + 1;
        // 
        $printerDriver = $printerLog->printer ?? null;
        if ($printerDriver) {
            $driver = new Driver($printerDriver);
            $result = $driver->printTicket($printerLog->data ?? '', $printerLog->print_method);
            if ($result) {
                $printerLog->reason = "打印成功";
                $printerLog->status = 2;
            } else {
                $printerLog->status = $num >= 3 ? 0 : 1;
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
        // 
        if ($printerLog->status == 1) {
            return false;
        } 
        return true;
        
    }
}