<?php

namespace app\job\service;

use think\facade\Cache;
use app\shop\model\settings\Printer;
use app\common\library\printer\Driver;
use app\common\model\settings\PrinterLog;
use PhpAmqpLib\Connection\AMQPStreamConnection;

/**
 *  消费者服务
 */
class RabbitmqConsumerService
{

    // 生成消费者
    public function create()
    {
        try {
            $config = config('rabbitmq');
            $connection = new AMQPStreamConnection($config['host'], $config['port'], $config['user'], $config['password']);
            // 获取连接的通道
            $channel = $connection->channel();
            //
            for ($i = 0; $i < 10; $i++) {
                $channel->queue_declare('print-data-system-' . $i, false, false, false, false);
                $channel->basic_qos(null, 1, false);
                for ($j = 0; $j < 2; $j++) {
                    $channel->basic_consume('print-data-system-' . $i, '', false, false, false, false, function ($msg) {
                        $this->consume($msg);
                    });
                }
            }
            //
            $channel->consume();
            $channel->close();
            $connection->close();
        } catch (\Throwable $exception) {
            log_write('RabbitmqConsumerService 队列连接失败 :' . $exception->getMessage(), 'task');
            throw new \Exception("队列连接失败: " . $exception->getMessage());
        }
    }

    // 执行消费者服务
    private function consume($msg)
    {
        $data = json_decode($msg->body, true);
        if (!$data || !isset($data['printer_id']) || !isset($data['data']) || !isset($data['id'])) {
            $msg->delivery_info['channel']->basic_ack($msg->delivery_info['delivery_tag']);
            return false;
        }
        // 间隔10秒
        if (Cache::get('__RABBITMQCONSUMERSERVICE__' . $data['id'])) {
            $msg->delivery_info['channel']->basic_nack($msg->delivery_info['delivery_tag'], false, true);
            return false;
        }
        // 查数据
        $printerDriver = Printer::find($data['printer_id']);
        if (!$printerDriver) {
            PrinterLog::where('id', '=', $data['id'])->update([
                'status' => 0,
                'reason' => '打印机不存在'
            ]);
            $msg->delivery_info['channel']->basic_ack($msg->delivery_info['delivery_tag']);
            return false;
        }
        // 发送打印
        $result = (new Driver($printerDriver))->timingPrintTicket($data['data'] ?? '');
        // 查打印日志
        $printerLog = PrinterLog::find($data['id']);
        if (!$printerLog) {
            $msg->delivery_info['channel']->basic_ack($msg->delivery_info['delivery_tag']);
            return false;
        }
        // 计数
        $num = $printerLog->num + 1;
        if ($result) {
            $printerLog->reason = "打印成功";
            $printerLog->status = 2;
            $msg->delivery_info['channel']->basic_ack($msg->delivery_info['delivery_tag']);
        } else {
            $printerLog->status = $num >= 5 ? 0 : 1;
            if ($printerLog->status == 0) {
                $printerLog->reason = "打印失败，未连接打印机";
                $msg->delivery_info['channel']->basic_ack($msg->delivery_info['delivery_tag']);
            } else {
                # 发送重新入队消息
                $msg->delivery_info['channel']->basic_nack($msg->delivery_info['delivery_tag'], false, true);
                Cache::set('__RABBITMQCONSUMERSERVICE__' . $data['id'], 1, 10);
            }
        }
        //
        $printerLog->printer_time = time();
        $printerLog->num = $num;
        $printerLog->save();
    }
}
