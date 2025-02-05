<?php

namespace app\admin\model\settings;

use think\facade\Cache;
use app\common\enum\settings\SettingEnum;
use app\common\model\settings\Setting as SettingModel;

class Setting extends SettingModel
{

    /**
     * 新增
     */
    public function add($data)
    {
        $service = $this->where(['key' => SettingEnum::SYS_ADMIN_CONFIG])->find();
        if (isset($data['weixin_service'])) {
            $this->writeCertPemFiles($data['weixin_service']['cert_pem'], $data['weixin_service']['key_pem']);
        }
        //
        Cache::set(SettingEnum::CLOUD_BASIC, ['base' => $data]);
        //
        if (!$service) {
            $add['key'] = SettingEnum::SYS_ADMIN_CONFIG;
            $add['describe'] = '系统设置';
            $add['values'] = $data;
            return $this->save($add);
        } else {
            $this->where(['key' => SettingEnum::SYS_ADMIN_CONFIG])->save(['values' => json_encode($data), 'update_time' => time()]);
            return true;
        }
    }

    /**
     * 写入cert证书文件
     */
    private function writeCertPemFiles($cert_pem = '', $key_pem = '')
    {
        if (empty($cert_pem) || empty($key_pem)) {
            return false;
        }
        // 证书目录
        $filePath = root_path() . 'runtime/cert/appwx/10000/';
        // 目录不存在则自动创建
        if (!is_dir($filePath)) {
            mkdir($filePath, 0755, true);
        }
        // 写入cert.pem文件
        if (!empty($cert_pem)) {
            file_put_contents($filePath . 'cert.pem', $cert_pem);
        }
        // 写入key.pem文件
        if (!empty($key_pem)) {
            file_put_contents($filePath . 'key.pem', $key_pem);
        }
        return true;
    }
}
