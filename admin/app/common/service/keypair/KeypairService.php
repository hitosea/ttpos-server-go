<?php

namespace app\common\service\keypair;

use help\DooHelp;
use help\JsEncrypt;
use help\StringHelp;
use think\facade\Cache;

/**
 * 客户端KEY
 */
class KeypairService
{
    /**
     * 客户端KEY
     */
    public static function getKey($client_id, $type = '')
    {
        $type = $type ?: request()->param('type', 'pgp');
        // 
        $cacheKey = "KeyPair::" . $client_id . $type;
        if (Cache::has($cacheKey)) {
            $cacheData = Cache::get($cacheKey);
            if ($cacheData['private_key'] ?? '') {
                return [
                    'type' => $type,
                    'id' => $client_id,
                    'key' => $cacheData['public_key'],
                ];
            }
        }
        if ($type == 'jsencrypt') {
            $data = JsEncrypt::gnerateKeys();
        } else {
            $name = StringHelp::generatePassword(6);
            $email = 'aa@bb.cc';
            $data = DooHelp::pgpGenerateKeyPair($name, $email, StringHelp::generatePassword());
        }
        // 
        Cache::set($cacheKey, $data, 60 * 60 * 24 * 30);
        //
        return [
            'type' => $type,
            'id' => $client_id,
            'key' => $data['public_key'],
        ];
    }
}
