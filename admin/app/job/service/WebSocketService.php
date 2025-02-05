<?php

namespace app\job\service;

use Workerman\Worker;
use think\facade\Cache;
use think\worker\Server;
use Workerman\Lib\Timer;
use app\common\model\shop\BindRecord;
use app\common\model\websock\WebSocketMsg;
use app\cashier\model\cashier\User as UserModel;

class WebSocketService extends Server
{
    protected $connections = array();

    // 进程数
    const PROCESSES_COUNT = 4;

    // 缓存key
    const PROT_KEY = '__WORKER_START_PROTS__';

    // 心跳间隔时间（秒）
    // doto - 时间需要还原
    const HEARTBEAT_INTERVAL = 111155;

    // N秒内只允许推送一次相同的数据
    const CHECKSUM_INTERVAL = 1;
    
    // 离线消息过期时间（秒）
    const OFFLINE_MESSAGE_EXPIRE = 7200;

    protected $heartbeatTimer;
    protected $retryAttempts = [];

    // 消息类型
    const MSG_TYPE_HEARTBEAT = 'heartbeat';                 // 心跳
    const MSG_TYPE_UPDATE_PRODUCT = 'update_product';       // 更新商品


    /**
     * 处理启动事件
     */
    public function onWorkerStart($connection)
    {
        // 开启心跳定时器
        $this->heartbeatTimer = Timer::add(self::HEARTBEAT_INTERVAL, function() {
            $this->checkHeartbeat();
        });
        
        // 开启一个内部端口，方便内部系统推送数据，Text协议格式 文本+换行符
        $innerTextWorker = new Worker('text://0.0.0.0:123' . $connection->id);
        $innerTextWorker->onMessage = function ($c, $buffer) use ($connection) {
            try {
                // $data数组格式，里面有uid，表示向那个uid的页面推送数据
                $data = json_decode($buffer, true);
                // 处理内部消息
                $ret = $this->handleInternalMessage($data);
                // 返回推送结果
                $c->send($ret ? 'ok' : 'fail');
            } catch (\Exception $e) {
                $this->handleError($e);
                $c->send('error');
            }
        };
        $innerTextWorker->listen();
    }

    /**
     * 处理客户端连接事件
     */
    public function onConnect($connection)
    {
        $connection->onWebSocketConnect = function ($connection, $httpBuffer) {
            try {
                if (!$bindRecord = $this->validateConnection($connection, $httpBuffer)) {
                    return $connection->close();
                }
                // 发送链接成功
                $connection->send(json_encode([
                    'state' => 200,
                    'msg' => 'Connected successfully'
                ]));
                // 初始化链接信息
                $this->initializeConnection($connection, $bindRecord);
                // 补发离线消息
                $this->reissueSend($connection);
                // 
            } catch (\Exception $e) {
                $this->handleError($e);
                $connection->close();
            }
        };
    }

    /**
     * 处理客户端消息事件
     */
    public function onMessage($connection, $data)
    {
        try {
            $message = json_decode($data, true);
            // 处理心跳消息
            if (isset($message['type']) && $message['type'] === 'heartbeat') {
                $connection->lastHeartbeat = time();
                $connection->send(json_encode([
                    'type' => 'heartbeat',
                    'state' => 200
                ]));
                return;
            }
            // 处理其他消息
            $this->processMessage($connection, $message);
            // 
        } catch (\Exception $e) {
            $this->handleError($e);
            $connection->send(json_encode([
                'state' => 500,
                'msg' => 'Message processing failed'
            ]));
        }
    }

    /**
     * 处理客户端断开连接事件
     */
    public function onClose($connection)
    {
        try {
            /**
             * 清理连接
             */
            if (isset($connection->appId) && isset($connection->uid)) {
                unset($this->connections[$connection->appId][$connection->uid]);
                // 缓存连接 - 5分钟内都算在线
                Cache::hdel('connection_' . $connection->appId, $connection->uid);
                Cache::hset('connection_' . $connection->appId, $connection->uid, time());
                // 
                if (count($this->connections[$connection->appId]) == 0) {
                    unset($this->connections[$connection->appId]);
                }
            }
        } catch (\Exception $e) {
            $this->handleError($e);
        }
    }

    /**
     * 处理客户端上线补发
     */
    private function reissueSend($connection)
    {
        try {
            if (!isset($connection->uid) && !isset($connection->appId)) {
                return;
            }
            // 
            $socketMsgList = (new WebSocketMsg([], 0))->where([
                    'uid' => $connection->uid, 
                    'app_id' => $connection->appId, 
                    'status' => 0
                ])->select();
            foreach ($socketMsgList as $socketMsg) {
                $connection->send($socketMsg->msg);
                $socketMsg->status = 1;
                $socketMsg->update_time = time();
                $socketMsg->save();
            }
            // 
        } catch (\Exception $e) {
            $this->handleError($e);
        }
    }

    /**
     * 验证连接
     */
    private function validateConnection()
    {
        $parts = parse_url($_SERVER['REQUEST_URI'] ?? '');
        parse_str($parts['query'] ?? '', $queryParams);
        $token = $queryParams['token'] ?? '';
        $client = $queryParams['client'] ?? '';
        if (!$token || !$client) {
            return false;
        }
        // 校验token
        $data = checkToken($token, $client);
        if ($data['code'] != 1) {
            return false;
        }
        // 获取用户信息
        if (!$user = UserModel::getUser($data['data'])) {
            return false;
        }
        // 
        request()->appId = $user['app_id'];
        request()->shopSupplierId = $user['shop_supplier_id'];
        // 验证设备是否绑定
        $bindRecord = (new BindRecord([], $user['app_id']))::where('key',  $data['data']['device_id'] ?? '')->find();
        if (!$bindRecord) {
            return false;
        }
        return $bindRecord;
    }

    /**
     * 初始化连接
     */
    private function initializeConnection($connection, $bindRecord)
    {
        $connection->appId = $bindRecord->app_id;
        $connection->uid = $bindRecord->key;
        $connection->sourceClient = $bindRecord->source;
        $connection->lastHeartbeat = time();
        $connection->isAlive = true;
        // 处理离线缓存
        Cache::hdel('connection_' . $connection->appId, $connection->uid);
        // 
        $this->connections[$connection->appId][$connection->uid] = $connection;
    }

    /**
     * 检查心跳状态
     */
    private function checkHeartbeat()
    {
        $now = time();
        foreach ($this->connections as $connections) {
            foreach ($connections as $connection) {
                if ($now - $connection->lastHeartbeat > self::HEARTBEAT_INTERVAL * 2) {
                    $connection->close();
                }
            }
        }
    }

    /**
     * 错误处理
     */
    private function handleError(\Exception $e)
    {
        // 记录错误日志
        \think\facade\Log::error('WebSocket Error: ' . $e->getMessage());
    }

    /**
     * 处理内部消息
     */
    private function handleInternalMessage($data)
    {
        $appid = $data['appid'];
        $uid = $data['uid'];
        $message = $data['data'];
        $msgType = $data['type'];
        $cs = $data['client'];
        $time = $data['time'] ?? 0;
        // 针对uid推送数据
        return $this->sendMessageByUid($appid, $cs, $uid, $msgType, $message, $time);
    }

    /**
     * 处理消息
     */
    private function processMessage($connection, $message)
    {
        // 已读删除
        if (isset($message['type']) && isset($message['msg_id']) && $message['type'] === 'reply') {
            $connection->send(json_encode([
                'type' => 'reply',
                'state' => 200,
                'msg_id' => $message['msg_id'],
                'msg' => 'Reply successfully'
            ]));
            (new WebSocketMsg([], 0))->where(['id' => $message['msg_id'], 'app_id' => $connection->appId])->update(['status' => 1, 'update_time' => time()]);
        }
    }


    /**
     * 限制频繁推送, N秒内只允许推送一次相同的数据
     */
    private function handleFrequentlyMessage($message, $appid, $uid, $msg_time)
    {
        // N秒内只允许推送一次相同的数据
        $checksum = md5($message);
        if (Cache::has('web_socket_send_checksum_' . $checksum . $appid . $uid . $msg_time)) {
            return false;
        }
        Cache::set('web_socket_send_checksum_' . $checksum . $appid . $uid . $msg_time, 1, self::CHECKSUM_INTERVAL);
        return true;
    }

    /**
     * 针对uid推送数据
     */
    private function sendMessageByUid($appid, $cs = '*', $uid = '*',string $msgType = '', string $message = "", $msg_time = 0)
    {
        $data = json_decode($message, true);
        // 
        if ($uid == '*') {
            // 
            $uids = [];
            $webSocketMsg = (new WebSocketMsg([], 0));
            // 处理在线消息
            foreach (($this->connections[$appid] ?? []) as $item) {
                if ($cs != '*' && $cs != $item->sourceClient) {
                    continue;
                }
                $uids[] = $item->uid;

                // N秒内只允许推送一次相同的数据
                if (!$this->handleFrequentlyMessage($message, $appid, $item->uid, $msg_time)) {
                    continue;
                }

                // 1. 先删后加 - 同一个类型，只保留最新的
                $webSocketMsg->where(['uid' => $item->uid, 'type' => $msgType, 'app_id' => $item->appId, 'status' => 0])->delete();
                $res = $webSocketMsg->create([
                    'uid' => $item->uid,
                    'app_id' => $item->appId,
                    'source_client' => $item->sourceClient,
                    'type' => $msgType,
                    'msg' => json_encode($data),
                ]);
                $data['msg_id'] = $res['id'];

                // 推送
                $item->send(json_encode($data));
            }

            // 处理离线消息
            $connections = Cache::hgetall('connection_' . $appid);
            foreach ($connections as $uid => $time) {
                if (in_array($uid, $uids)) {
                    Cache::hdel('connection_' . $appid, $uid);
                    continue;
                }

                // 多线程推送 - 隔离限制
                $crc32 = crc32($uid . $appid . $msgType);
                if (Cache::has('connection_' . $appid . $crc32)) {
                    continue;
                }
                Cache::set('connection_' . $appid . $crc32, 1, 2);

                // N秒内只允许推送一次相同的数据
                if (!$this->handleFrequentlyMessage($message, $appid, $uid, $msg_time)) {
                    continue;
                }

                // 
                if ((time() - $time) < self::OFFLINE_MESSAGE_EXPIRE) {
                    // 1. 先删后加 - 同一个类型，只保留最新的
                    $bindRecord = (new BindRecord([],$appid))->where('key', $uid)->find();
                    if (!$bindRecord) {
                        continue;
                    }
                    if ($cs != '*' && $cs != $bindRecord->source) {
                        continue;
                    }
                    // 
                    $webSocketMsg = (new WebSocketMsg([], 0));
                    $webSocketMsg->where(['uid' => $uid, 'type' => $msgType, 'app_id' => $appid, 'status' => 0])->delete();
                    $webSocketMsg->create([
                        'uid' => $uid,
                        'app_id' => $appid,
                        'type' => $msgType,
                        'source_client' => $bindRecord->source,
                        'is_offline' => 1,
                        'msg' => json_encode($data),
                    ]);
                } else {
                    Cache::hdel('connection_' . $appid, $uid);
                }
            }
            // 
            return true;
        } else {
            $webSocketMsg = (new WebSocketMsg([], 0));
            if (isset($this->connections[$appid][$uid])) {
                $data['uid'] = $this->connections[$appid][$uid]->uid;

                // N秒内只允许推送一次相同的数据
                if (!$this->handleFrequentlyMessage($message, $appid, $uid, $msg_time)) {
                    return false;
                }

                // 推送
                $webSocketMsg->where(['uid' => $uid, 'type' => $msgType, 'app_id' => $appid, 'status' => 0])->delete();
                $res = $webSocketMsg->create([
                    'uid' => $this->connections[$appid][$uid]->uid,
                    'app_id' => $this->connections[$appid][$uid]->appId,
                    'type' => $msgType,
                    'source_client' => $this->connections[$appid][$uid]->sourceClient,
                    'status' => 0,
                    'msg' => json_encode($data),
                ]);
                $data['msg_id'] = $res['id'];
                $this->connections[$appid][$uid]->send(json_encode($data));
                // 
                return true;
            }
        }
        // 
        return false;
    }

    /**
     * 推送给指定客户端
     */
    private static function pushClient($appid, $cs = '*', $uid = "", $msgType = '', array $data = [], $msg_time = 0)
    {
        for ($i = 0; $i < self::PROCESSES_COUNT; $i++) {
            try {
                $client = stream_socket_client('tcp://127.0.0.1:123' . $i, $errno, $errmsg, 1);
                if (!$client) {
                    continue;
                }
                // 推送的数据，包含uid字段，表示是给这个uid推送
                // 发送数据，注意5678端口是Text协议的端口，Text协议需要在数据末尾加上换行符
                fwrite($client, json_encode([
                    'appid' => $appid,
                    'client' => $cs,
                    'uid' => $uid,
                    'type' => $msgType,
                    'time' =>$msg_time,
                    'data' => json_encode(['type' => $msgType, 'data' => $data])
                ]) . "\n");
                // 读取推送结果
                // $res = fread($client, 8192);
                //
            } catch (\Throwable $th) {
                //throw $th;
            }
        }
    }

    /**
     * 推送给所有客户端 - 禁止使用
     */
    public static function pushAllClient($appid, $cs = '*', $msgType = '', array $data = [], $msg_time = 0, $channel = '')
    {
        self::pushClient($appid, $cs, "*", $msgType, $data, $msg_time, $channel);
    }

    /**
     * 通过订阅发布推送给客户端
     */
    public static function publish($appid, $cs = '*', $msgType = '', array $data = [])
    {
        $msg = json_encode([
            'app_id' => $appid,
            'client' => $cs,
            'msg_type' => $msgType,
            'msg_time' =>1,
            'msg' => $data,
        ]);
        // 需要实现一秒只推送最后一条
        $redis = Cache::store('redis')->handler();
        $redis->publish('web_socket_send_0', $msg);
        usleep(50000);
        $serverIdentifier = php_uname('n') . '_' . gethostbyname(php_uname('n'));
        if (!$redis->get($serverIdentifier . '_' .'web_socket_send_0' . '_' . md5($msg))) {
            $redis->publish('web_socket_send_1', $msg);
            trace("sdasdasdsadasdasdasdsa");
        }
    }
}
