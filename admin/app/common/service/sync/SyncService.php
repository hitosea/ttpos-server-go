<?php

namespace app\common\service\sync;

use help\ImgHelp;
use help\HttpHelp;
use think\facade\Cache;
use app\common\enum\settings\SettingEnum;

/**
 * 同步服务
 */
class SyncService
{
    protected $error = '';
    protected $cloud_platform_host;

    // 初始化
    public function __construct()
    {
        $this->cloud_platform_host = env('CLOUD_PLATFORM_HOST');
    }

    /**
     * 同步基础信息
     */
    public function syncBaseInfo()
    {
        $url = $this->cloud_platform_host . "/api/admin/setting.service/info";
        $cloudBasic = HttpHelp::postRequest($url);
        if (!$cloudBasic) {
            return false;
        }
        //
        $md5Key = md5($cloudBasic);
        if (Cache::get(SettingEnum::CLOUD_BASIC) && $md5Key == Cache::get("__SYNCBASEINFO__")) {
            return Cache::get(SettingEnum::CLOUD_BASIC);
        }
        //
        $cloudBasic = json_decode($cloudBasic ?? '{}', true);
        if ($cloudBasic && isset($cloudBasic['data'])) {
            $values = $cloudBasic['data'];
            // 处理图片
            $brandLogoPath = ImgHelp::downloadCloudImage($values['base']['brand_logo'] ?? '', 'uploads/cloud/');
            $values['base']['brand_logo'] = ImgHelp::removeImageDomain($brandLogoPath);
            $brandLogoLongPath = ImgHelp::downloadCloudImage($values['base']['brand_logo_long'] ?? '', 'uploads/cloud/');
            $values['base']['brand_logo_long'] = ImgHelp::removeImageDomain($brandLogoLongPath);
            $browserLogoPath = ImgHelp::downloadCloudImage($values['base']['browser_logo'] ?? '', 'uploads/cloud/');
            $values['base']['browser_logo'] = ImgHelp::removeImageDomain($browserLogoPath);
            //
            Cache::set(SettingEnum::CLOUD_BASIC, $values);
            Cache::set("__SYNCBASEINFO__", $md5Key);
            //
            return $values;
        }
        return false;
    }

    /**
     * 获取客户端最新版本
     */
    public function getNewVersion($type, $brand)
    {
        $url = $this->cloud_platform_host . "/api/admin/client.client/getNewVersion?type=$type&brand=$brand&language=" . checkDetect();
        $result = HttpHelp::postRequest($url);
        if (!$result) {
            return [];
        }
        $result = json_decode($result ?? '{}', true);
        if (($result['code'] ?? 0) != 1) {
            return [];
        }
        $data = $result['data'] ?: [];
        //
        if (isset($data['download_url']) && (strstr($data['download_url'], 'http://nginx/') || strstr($data['download_url'], 'https://nginx/'))) {
            $data['download_url'] = str_replace('http://nginx/', base_url(), $data['download_url']);
            $data['download_url'] = str_replace('https://nginx/', base_url(), $data['download_url']);
        }
        //
        return $data;
    }
}
