<?php

namespace help;

use think\facade\Cache;

class QueueHelp
{
    // 
    protected $cacheKey;

    // 
    protected $queueUuid;

    /**
     * 架构方法 设置参数
     * @access public
     * @param int  $imageWidth | height
     * @param int $direction 方向 
     */
    public function __construct($cacheKey)
    {
        $this->cacheKey = $cacheKey;
        $this->queueUuid = StringHelp::uuid();
    }

    /**
     * 开始队列
     * @param $cacheKey, $queueUuid
     * @return void
     */
    public function while()
    {
        // Cache::hset($this->cacheKey, $this->queueUuid, time());
        // while (($firstKey = (array_keys(Cache::hgetall($this->cacheKey))[0] ?? '')) != $this->queueUuid) {
        //     if ($firstKey) {
        //         $firstKeyTIme = Cache::hget($this->cacheKey, $firstKey);
        //         if ((time() - $firstKeyTIme) > 5) {
        //             Cache::hdel($this->cacheKey, $firstKey);
        //         }
        //     } else {
        //         Cache::hdel($this->cacheKey, $firstKey);
        //     }
        //     usleep(5000); 
        // }
        // return $this;
    }

    /**
     * 释放
     * @param $cacheKey, $queueUuid
     * @return void
     */
    public function release()
    {
        Cache::hdel($this->cacheKey, $this->queueUuid);
    }

    /**
     * 监听销毁-释放
     */
    public function __destruct() {
        $this->release();
    }
}
