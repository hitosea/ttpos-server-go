<?php
namespace app\job\queue;

use think\queue\Job;
use think\facade\Log;

abstract class BaseJob
{
    /**
     * 执行任务
     */
    public function fire(Job $job, $data)
    {
        try {
            // 验证商户权限
            if (!$this->validateMerchant($data)) {
                $job->delete();
                return;
            }

            // 执行具体任务
            $result = $this->handle($job, $data);
            
            if ($result) {
                $job->delete();
            } else {
                if ($job->attempts() < 3) {
                    $job->release(3);
                } else {
                    $job->delete();
                    $this->failed($data);
                }
            }
        } catch (\Exception $e) {
            if ($job->attempts() < 3) {
                $job->release(3);
            } else {
                $this->failed($data);
                Log::error('打印队列任务执行失败: ' . $e->getMessage());
                $job->delete();
            }
        }
    }

    /**
     * 验证商户权限
     */
    protected function validateMerchant($data)
    {
        request()->appId = $data['app_id'] ?? 0;
        // 
        if (!request()->appId) {
            return false;
        }
        // 验证商户是否存在和有效
        return true; // 实现您的验证逻辑
    }

    /**
     * 具体任务处理
     */
    abstract protected function handle(Job $job, $data);

    /**
     * 任务失败处理
     */
    protected function failed($data)
    {
        Log::error('打印队列任务执行失败: ' . json_encode($data));
    }
}