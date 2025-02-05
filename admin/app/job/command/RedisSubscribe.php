<?php

namespace app\job\command;

use think\facade\Cache;
use think\console\Input;
use think\console\Output;
use think\console\Command;
use app\job\service\WebSocketService;

// Redis订阅
class RedisSubscribe extends Command
{
    protected function configure()
    {
        $this->setName('redis:subscribe')->setDescription('Start a Redis subscription');
    }

    protected function execute(Input $input, Output $output)
    {
        $serverIdentifier = php_uname('n') . '_' . gethostbyname(php_uname('n'));
        // 获取 Redis 实例
        $redis = Cache::store('redis')->handler();
        // 设置读取永不超时
        $redis->setOption(\Redis::OPT_READ_TIMEOUT, -1);
        // 订阅频道
        $redis->subscribe(['web_socket_send_' . (getmypid() % 2 === 0 ? 1 : 0)], function ($redis, $channel, $message) use ($output, $serverIdentifier) {
            // 多线程推送 - 隔离限制
            Cache::set($serverIdentifier . '_' . $channel . '_' . md5($message), 1, 3);
            // 处理接收到的消息
            try {
                $data = json_decode($message, true);
                WebSocketService::pushAllClient($data['app_id'], $data['client'], $data['msg_type'], $data['msg'], $data['msg_time'], $channel);
            } catch (\Throwable $th) {
                $output->writeln($th->getMessage());
                \think\facade\Log::error($th->getMessage());
            }
        });
    }
}
