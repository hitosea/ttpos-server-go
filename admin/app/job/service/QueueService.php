<?php
namespace app\job\service;

use think\facade\Queue;

class QueueService
{
    /**
     * 获取队列名称
     */
    public static function getQueueName($type, $appId)
    {
        return sprintf('%d-%s', $appId, $type);
    }

    /**
     * 投递任务到队列
     */
    public static function push($jobClass, $appId, $data=[], $type = 'print')
    {
        // 添加商户信息
        $data['app_id'] = $appId;
        // 生成队列名称
        // $queueName = self::getQueueName($type, $data['app_id']);
        // 
        return Queue::push($jobClass, $data, $type);
    }

    /**
     * 延迟投递任务
     */
    public static function later($delay, $jobClass, $appId, $data, $type = 'print')
    {
        $data['app_id'] = $appId;
        // 生成队列名称
        // $queueName = self::getQueueName($type, $data['app_id']);
        // 
        return Queue::later($delay, $jobClass, $data, $type);
    }
}