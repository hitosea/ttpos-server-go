<?php

declare(strict_types=1);

namespace app\job\command;

use help\HttpHelp;
use Workerman\Worker;
use think\console\Input;
use Workerman\Lib\Timer;
use think\console\Output;
use think\console\Command;
use think\console\input\Option;
use think\console\input\Argument;
use app\job\service\WebSocketService;
use app\job\service\RabbitmqConsumerService;

/**
 * 初始队列启动器
 */
class Job extends Command
{
    private $second = 60;
    protected function configure()
    {
        // 指令配置
        $this->setName('job')
            ->addArgument('action', Argument::OPTIONAL, "start|stop|restart|reload|status|connections", 'start')
            ->addOption('mode', 'm', Option::VALUE_OPTIONAL, 'Run the workerman server in daemon mode.')
            ->setDescription('定时任务');
    }

    protected function execute(Input $input, Output $output)
    {
        // 指令输出
        $action = $input->getArgument('action');
        $mode = $input->getOption('mode');
        // 重新构造命令行参数, 以便兼容workerman的命令
        global $argv;
        $argv = [];
        array_unshift($argv, 'think', $action);
        if ($mode == 'd') {
            $argv[] = '-d';
        } else if ($mode == 'g') {
            $argv[] = '-g';
        }

        // 设置 PID 文件路径到 runtime 目录
        Worker::$pidFile = runtime_path() . 'workerman.pid';

        // 修复 workerman 的原始bug
        $this->repairIntBug();

        // 创建消费者
        // $this->startRabbitmqConsumer();

        // 创建定时
        $this->startTimer();

        // 创建 WebSocket - 暂时不启用
        // $this->startWebsocket();

        // 输出当前的启动时间 - 以便用于日志追查
        echo("\n[" . date("Y-m-d H:i:s") . "]\n");

        // 启动
        Worker::runAll();
    }

    /**
     * 启动 Rabbitmq-消费者
     * @return false|int
     */
    public function startRabbitmqConsumer()
    {
        $worker = new Worker();
        $worker->onWorkerStart = function ($connection) {
            $service = new RabbitmqConsumerService();
            $service->create();
        };
    }

    /**
     * 启动定时器
     * @return false|int
     */
    public function startTimer()
    {
        $worker = new Worker();
        $worker->count = 1;
        $worker->onWorkerStart = function () {
            // 定时任务
            Timer::add($this->second, function () {
                try {
                    $appSecret = md5(env('APP_ID'));
                    HttpHelp::postRequest("http://nginx/index.php/job/index/index?app_secret=$appSecret&type=JobScheduler");
                } catch (\Throwable $e) {
                    echo 'ERROR: ' . $e->getMessage() . PHP_EOL;
                }
            });
        };
    }

    /**
     * 启动 Websocket
     * @return false|int
     */
    public function startWebsocket()
    {
        $service = new WebSocketService();
        // 创建WebSocket
        $worker = new Worker('websocket://0.0.0.0:2345');
        // 因为需要主动推送的关系，进程数设置为1：
        $worker->count = WebSocketService::PROCESSES_COUNT;
        // 处理启动事件
        $worker->onWorkerStart = [$service, 'onWorkerStart'];
        // 处理客户端连接事件
        $worker->onConnect = [$service, 'onConnect'];
        // 处理客户端消息事件
        $worker->onMessage = [$service, 'onMessage'];
        // 处理客户端断开连接事件
        $worker->onClose = [$service, 'onClose'];
    }

    /**
     * 修复 workerman 的原始bug
     * @return false|int
     */
    public function repairIntBug()
    {
        $selectPath = root_path("vendor/workerman/workerman/Events") . 'Select.php';
        $content = file_get_contents($selectPath);
        if (strpos($content, 'usleep($this->_selectTimeout);') !== false) {
            $content = str_replace('usleep($this->_selectTimeout);', 'usleep((int)$this->_selectTimeout);', $content);
            $content = str_replace('$ret = stream_select($read, $write, $except, 0, $this->_selectTimeout);', '$ret = stream_select($read, $write, $except, 0, (int)$this->_selectTimeout);', $content);
            file_put_contents($selectPath, $content);
        }
    }
}
