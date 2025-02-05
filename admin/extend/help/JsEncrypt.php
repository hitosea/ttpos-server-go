<?php

namespace help;

use think\facade\Cache;

class JsEncrypt
{
    /**
     * 生成PGP密钥对
     * @param $name
     * @param $email
     * @param string $passphrase
     * @return array
     */
    public static function gnerateKeys(): array
    {
        // 创建密钥对
        $res = openssl_pkey_new([
            "digest_alg" => "sha256",
            "private_key_bits" => 2048,
            "private_key_type" => OPENSSL_KEYTYPE_RSA,
        ]);
        // 生成私钥
        openssl_pkey_export($res, $private_key);
        // 从资源中提取公钥
        $public_key = openssl_pkey_get_details($res)["key"];
        // 
        return compact('private_key', 'public_key');
    }

    /**
     * 解密
     * @param $encryptedText
     * @param $privateKey
     * @param null $passphrase
     * @return string
     */
    public static function decryptApi($encryptedText, $privateKey): string
    {
        if (strlen($privateKey) < 50) {
            $keyCache = StringHelp::json2array(Cache::get("KeyPair::" . $privateKey . 'jsencrypt'));
            $privateKey = $keyCache['private_key'] ?? '';
        }
        $encryptedTexts = explode('||', base64_decode($encryptedText));
        // 解密数据
        $iv = $encryptedTexts[0];
        $encryptedData = $encryptedTexts[1];
        $encryptedKey = $encryptedTexts[2];
        // 
        openssl_private_decrypt(base64_decode($encryptedKey), $symmetricKey, $privateKey);
        // 调用解密函数
        return openssl_decrypt(base64_decode($encryptedData), 'AES-128-CBC', $symmetricKey, OPENSSL_RAW_DATA,  base64_decode($iv));
    }

    /**
     * 加密
     * @param $plaintext
     * @param $publicKey
     * @return string
     */
    public static function encrypt($plaintext, $publicKey): string
    {
        if (strlen($publicKey) < 50) {
            $keyCache = StringHelp::json2array(Cache::get("KeyPair::" . $publicKey . 'jsencrypt'));
            $publicKey = $keyCache['public_key'];
        }
        // 使用对称密钥加密数据
        $symmetricKey = StringHelp::generatePassword(32); // AES-256 需要 32 字节密钥
        $iv = openssl_random_pseudo_bytes(openssl_cipher_iv_length('aes-256-cbc'));
        $encryptedData = openssl_encrypt($plaintext, 'aes-256-cbc', $symmetricKey, OPENSSL_RAW_DATA, $iv);
        // 使用 RSA 加密对称密钥
        openssl_public_encrypt($symmetricKey, $encryptedKey, $publicKey);
        // 将加密后的数据、IV 和加密的对称密钥一起传输
        return base64_encode($iv . $encryptedData) . '||' .  base64_encode( $encryptedKey);
    }


    /**
     * 解析PGP参数
     * @param $string
     * @return string[]
     */
    public static function pgpParseStr($string): array
    {
        $array = [
            'encrypt_type' => '',
            'encrypt_id' => '',
            'client_type' => '',
            'client_key' => '',
        ];
        $string = str_replace(";", "&", $string);
        parse_str($string, $params);
        foreach ($params as $key => $value) {
            $key = strtolower(trim($key));
            if ($key) {
                $array[$key] = trim($value);
            }
        }
        if ($array['client_type'] === 'jsencrypt' && $array['client_key']) {
            $array['client_key'] = self::publicFormat($array['client_key']);
        }
        return $array;
    }

    /**
     * 还原公钥格式
     * @param $key
     * @return string
     */
    public static function publicFormat($key): string
    {
        $key = str_replace(["-", "_", "$"], ["+", "/", "\n"], $key);
        if (!str_contains($key, '-----BEGIN PUBLIC KEY-----')) {
            $key = "-----BEGIN PUBLIC KEY-----\n" . $key . "\n-----END PUBLIC KEY-----";
        }
        return $key;
    }
}
