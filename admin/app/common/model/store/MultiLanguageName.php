<?php

namespace app\common\model\store;

use think\facade\Cache;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 * 门店免单标签
 */
class MultiLanguageName extends BaseModel
{
    use SoftDelete;
    protected $name = 'multi_language_name';
    protected $pk   = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    // 缓存前缀
    const CACHE_PREFIX = 'multi_language_name:';

    /**
     * 保存多语言名称
     *
     * @param array|string $names 多语言名称数组或JSON字符串
     * @param int|null $uuid 更新时的UUID
     * @return int 返回UUID
     */
    public function saveNames($names, ?int $uuid = null): int
    {
        // 如果传入的是JSON字符串,则解码
        if (is_string($names)) {
            $names = json_decode($names, true);
        }

        // 准备数据
        $data = [
            'en_name'    => $names['en'] ?? '',
            'zh_name'    => $names['zh'] ?? '',
            'zh_tw_name' => $names['zhtw'] ?? '',
            'th_name'    => $names['th'] ?? '',
            'my_name'    => $names['my'] ?? '',
            'ja_name'    => $names['ja'] ?? '',
            'ko_name'    => $names['ko'] ?? '',
            'tr_name'    => $names['tr'] ?? '',
        ];

        if ($uuid) {
            // 更新
            $this->where('uuid', $uuid)->update($data);
            $this->clearCache($uuid);
            return $uuid;
        } else {
            // 新增
            $data['uuid'] = createUuid();
            $this->save($data);
            return $data['uuid'];
        }
    }

    /**
     * 获取多语言名称
     *
     * @param int $uuid UUID
     * @return string JSON格式的多语言名称
     */
    public function getNames(int $uuid): string
    {
        if (!$uuid) {
            return '';
        }

        // 尝试从缓存获取
        $cacheKey = self::CACHE_PREFIX . $uuid;
        $names = Cache::get($cacheKey);

        if ($names === null) {
            $record = $this->where('uuid', $uuid)
                ->field('en_name,zh_name,zh_tw_name,th_name,my_name,ja_name,ko_name,tr_name')
                ->find();

            if (!$record) {
                return '';
            }

            // 构建返回格式
            $names = json_encode([
                'en'   => $record['en_name'],
                'zh'   => $record['zh_name'],
                'zhtw' => $record['zh_tw_name'],
                'th'   => $record['th_name'],
                'my'   => $record['my_name'],
                'ja'   => $record['ja_name'],
                'ko'   => $record['ko_name'],
                'tr'   => $record['tr_name']
            ], JSON_UNESCAPED_UNICODE);

            // 写入缓存,有效期24小时
            Cache::set($cacheKey, $names, 86400);
        }

        return $names;
    }

    /**
     * 清除缓存
     *
     * @param int $uuid UUID
     */
    protected function clearCache(int $uuid): void
    {
        Cache::delete(self::CACHE_PREFIX . $uuid);
    }
}
