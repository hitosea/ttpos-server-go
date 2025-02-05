<?php

namespace help;

use FFI;
use think\facade\Cache;

class DooHelp
{
    private static $doo;

    /**
     * char转为字符串
     * @param $text
     * @return string
     */
    private static function string($text): string
    {
        return FFI::string($text);
    }

    /**
     * 装载
     * @param $token
     * @param $language
     */
    public static function load()
    {
        $licensePath = "license.so";
        $architecture = php_uname('m');
        if (strpos($architecture, 'arm') !== false || strpos($architecture, 'aarch64') !== false) {
            $licensePath = "license_arm.so";
        }
        self::$doo = FFI::cdef(<<<EOF
                char* PgpGenerateKeyPair(char* name, char* email, char* passphrase);
                char* PgpEncrypt(char* plainText, char* publicKey);
                char* PgpDecrypt(char* cipherText, char* privateKey, char* passphrase);
            EOF, "/var/www/bin/" . $licensePath);
    }

    /**
     * 获取实例
     * @param $token
     * @param $language
     * @return mixed
     */
    public static function doo($token = null, $language = null)
    {
        if (self::$doo == null) {
            self::load($token, $language);
        }
        return self::$doo;
    }

    /**
     * 生成PGP密钥对
     * @param $name
     * @param $email
     * @param string $passphrase
     * @return array
     */
    public static function pgpGenerateKeyPair($name, $email, string $passphrase = ""): array
    {
        return StringHelp::json2array(self::string(self::doo()->PgpGenerateKeyPair($name, $email, $passphrase)));
    }

    /**
     * PGP加密
     * @param $plaintext
     * @param $publicKey
     * @return string
     */
    public static function pgpEncrypt($plaintext, $publicKey): string
    {
        if (strlen($publicKey) < 50) {
            $keyCache = StringHelp::json2array(Cache::get("KeyPair::" . $publicKey . 'pgp'));
            $publicKey = $keyCache['public_key'];
        }
    
        return self::string(self::doo()->PgpEncrypt($plaintext, $publicKey));
    }

    /**
     * PGP解密
     * @param $encryptedText
     * @param $privateKey
     * @param null $passphrase
     * @return string
     */
    public static function pgpDecrypt($encryptedText, $privateKey, $passphrase = null): string
    {
        if (strlen($privateKey) < 50) {
            $keyCache = StringHelp::json2array(Cache::get("KeyPair::" . $privateKey . 'pgp'));
            $privateKey = $keyCache['private_key'] ?? '';
            $passphrase = $keyCache['passphrase'] ?? '';
        }
        return self::string(self::doo()->PgpDecrypt($encryptedText, $privateKey, $passphrase));
    }

    /**
     * PGP加密API
     * @param $plaintext
     * @param $publicKey
     * @return string
     */
    public static function pgpEncryptApi($plaintext, $publicKey): string
    {
        $content = StringHelp::array2json($plaintext);
        $content = self::PgpEncrypt($content, $publicKey);
        return preg_replace("/\s*-----(BEGIN|END) PGP MESSAGE-----\s*/i", "", $content);
    }

    /**
     * PGP解密API
     * @param $encryptedText
     * @param null $privateKey
     * @param null $passphrase
     * @return array
     */
    public static function pgpDecryptApi($encryptedText, $privateKey, $passphrase = null): array
    {
        $content = "-----BEGIN PGP MESSAGE-----\n\n" . $encryptedText . "\n-----END PGP MESSAGE-----";
        $content = self::pgpDecrypt($content, $privateKey, $passphrase);
        return StringHelp::json2array($content);
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
        if ($array['client_type'] === 'pgp' && $array['client_key']) {
            $array['client_key'] = self::pgpPublicFormat($array['client_key']);
        }
        return $array;
    }

    /**
     * 还原公钥格式
     * @param $key
     * @return string
     */
    public static function pgpPublicFormat($key): string
    {
        $key = str_replace(["-", "_", "$"], ["+", "/", "\n"], $key);
        if (!str_contains($key, '-----BEGIN PGP PUBLIC KEY BLOCK-----')) {
            $key = "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\n" . $key . "\n-----END PGP PUBLIC KEY BLOCK-----";
        }
        return $key;
    }
}
