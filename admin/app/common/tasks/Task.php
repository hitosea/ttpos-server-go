<?php

namespace app\common\tasks;

use Exception;
use think\Request;
use think\facade\Db;

/**
 * 异步任务基类
 */
class Task
{
    // 应用ID
    protected $appId;

    // 构造函数
    public function __construct() {}

    /**
     * 此函数用于获取参数。
     * @param string $paramName 要获取的参数名称。
     * @return mixed 获取的参数值。
     */
    public function getParams($parameters)
    {
        $paramDetails = [];
        foreach ($parameters as $param) {
            $paramName = $param->getName();
            // 获取参数的值，如果没有则使用默认值
            $paramDetails[] = $this->$paramName ?? $param->getDefaultValue() ?: null;
        }
        return $paramDetails;
    }

    /**
     * 任务投递
     * @param self $task 任务
     * @return void
     */
    public static function deliver($task)
    {
        $reflector = new \ReflectionClass($task);
        $constructor = $reflector->getConstructor();
        $parameters = $constructor->getParameters();
        $paramDetails = $task->getParams($parameters);
        // 记录启动任务
        $taskWorkerId = Db::name('task_worker')->insertGetId([
            'name' => get_class($task),
            'args' => json_encode($paramDetails, JSON_UNESCAPED_UNICODE),
            'create_time' => time()
        ]);
        // 初始化cURL会话
        $ch = curl_init();
        // 设置cURL选项
        curl_setopt($ch, CURLOPT_URL, "http://nginx/api/common/publics/asynTask");  // 目标URL
        curl_setopt($ch, CURLOPT_POST, true);                                       // 设置为POST请求
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
            'task_id' => $taskWorkerId,
            'task_name' => get_class($task),
            'task_data' => $paramDetails,
            'app_id' => request()->appId,
        ]));
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json'
        ]);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);                             // 返回结果而不是输出
        curl_setopt($ch, CURLOPT_HEADER, 0);
        curl_setopt($ch, CURLOPT_TIMEOUT_MS, 10);                                   // 设置超时时间（毫秒）
        // 执行cURL会话
        curl_exec($ch);
        // 关闭cURL会话
        curl_close($ch);
    }

    /**
     * 启动任务
     * @param Request $request
     * @return void
     */
    public function starts(Request $request)
    {
        $param = $request->param();
        // 查找未启动的任务
        $taskWorker = Db::name('task_worker')->where('id', $param['task_id'])->where('start_at', 0)->find();
        if ($taskWorker) {
            Db::name('task_worker')->where('id', $param['task_id'])->update([
                'start_at' => time(),
                'update_time' => time()
            ]);
            // 尝试启动任务
            $error = '';
            try {
                $classSpace = $param['task_name'];
                if (!class_exists($classSpace)) {
                    throw new Exception('未找到存储引擎类: ' . $classSpace);
                }
                // 
                request()->appId = $param['app_id'];
                // 
                $reflectionClass = new \ReflectionClass($classSpace);
                $instance = $reflectionClass->newInstanceArgs($param['task_data'] ?? []);
                $instance->start();
            } catch (\Throwable $th) {
                $error = $th->getMessage();
            }
            // 记录任务结束时间
            Db::name('task_worker')->where('id', $param['task_id'])->update([
                'end_at' => time(),
                'error' => $error,
                'update_time' => time()
            ]);
        }
        // 删除 task_worker 表 7天前的数据
        Db::name('task_worker')->where('end_at', '>' , 0)->where('create_time', '<', time() - 7 * 24 * 60 * 60)->delete();
    }
}
