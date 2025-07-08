<?php

namespace app\shop\middleware;

use help\DooHelp;
use help\JsEncrypt;
use think\facade\Cache;
use app\common\exception\BaseException;

/**
 * 删除缓存中间件
 * Class DeleteCache
 * @package app\shop\middleware
 */
class DeleteCache
{
    //自执行中间件方法
    public function handle($request, \Closure $next)
    {
        $response = $next($request);

        $this->deleteCache($request->appId);

        return $response;
    }

    /**
     * 删除缓存
     */
    private function deleteCache($companyUuid)
    {
        // 按前缀删除缓存
        $prefix = '{TTPOS_GORM_CACHE_DATA}:' . $companyUuid;
        // 获取Redis实例
        $redis = Cache::handler();
        // 检查Redis连接是否正常
        if (!$redis) {
            dump('Redis连接失败');
            die;
        }
        // 尝试使用scan命令代替keys（更适合生产环境）
        $iterator = null;
        $keys = [];
        $pattern = $prefix . '*';
        do {
            // scan命令的格式：scan($iterator, $pattern, $count)
            $scanKeys = $redis->scan($iterator, $pattern, 10000);
            if ($scanKeys) {
                $keys = array_merge($keys, $scanKeys);
            }
        } while ($iterator > 0);
        // 如果scan命令不可用，尝试使用keys命令
        if (empty($keys)) {
            $keys = $redis->keys($pattern);
        }
        // 删除找到的键
        $deleteCount = 0;
        if (!empty($keys)) {
            foreach ($keys as $key) {
                if ($redis->del($key)) {
                    $deleteCount++;
                }
            }
        }
    }
}
