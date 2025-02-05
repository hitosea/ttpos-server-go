<?php

namespace help;

use Exception;

class RSAEncryptorHelp
{

    const MAX_ENCRYPT_BLOCK_SIZE = 117;
    const MAX_DECRYPT_BLOCK_SIZE = 128;

    const DEFAULT_CHARSET = "UTF-8";

    /**
     * 从文件或字符铭文创建私钥资源标识符
     * @param $privateKey
     * @return false|OpenSSLAsymmetricKey
     */
    public function getPrivateKey($privateKey)
    {
        if(file_exists($privateKey)){
            $privateKey = file_get_contents($privateKey);
        }

        if (strpos($privateKey,'-----BEGIN PRIVATE KEY-----') === false){
            $privateKey = "-----BEGIN PRIVATE KEY-----\n" .
                wordwrap($privateKey, 64, "\n", true) .
                "\n-----END PRIVATE KEY-----";;
        }

        return openssl_pkey_get_private($privateKey);
    }

    /**
     * 从文件或字符铭文创建公钥资源标识符
     * @param $publicKey
     * @return false|OpenSSLAsymmetricKey
     */
    public function getPublicKey($publicKey)
    {
        if(file_exists($publicKey)){
            $publicKey = file_get_contents($publicKey);
        }

        if (strpos($publicKey,'-----BEGIN PUBLIC KEY-----') === false){
            $publicKey = "-----BEGIN PUBLIC KEY-----\n".wordwrap($publicKey, 64, "\n", true) ."\n-----END PUBLIC KEY-----";
        }

        return openssl_pkey_get_public($publicKey);
    }


    /**
     * RSA公钥分段加密数据后返回base64编码
     * @param $data
     * @param $rsaPublicKey
     * @return string
     * @throws Exception
     */
    public function pubEncrypt($data, $rsaPublicKey) {

        $crypted = "";
        $dataArray = str_split($data, self::MAX_ENCRYPT_BLOCK_SIZE);
        foreach($dataArray as $subData){
            $subCrypted = null;
            if(!openssl_public_encrypt($subData, $subCrypted, $rsaPublicKey)){
                throw new \Exception(openssl_error_string());
            }
            $crypted .= $subCrypted;
        }

        return base64_encode($crypted);
    }

    /**
     * RSA公钥分段解密
     * @param $data
     * @param $rsaPublicKey
     * @return string
     * @throws Exception
     */
    public function pubDecrypt($data, $rsaPublicKey) {
        $encryptstr = base64_decode($data);
        $decrypted = [];
        $dataArray = str_split($encryptstr, self::MAX_DECRYPT_BLOCK_SIZE);

        foreach($dataArray as $subData){
            $subDecrypted = null;
            if(!openssl_public_decrypt($subData, $subDecrypted, $rsaPublicKey)){
                throw new \Exception(openssl_error_string());
            }
            $decrypted[] = $subDecrypted;
        }
        return implode('',$decrypted);
    }

    /**
     * RSA私钥分段加密数据后返回base64编码
     * @param $data
     * @param $rsaPublicKey
     * @return string
     * @throws Exception
     */
    public function privEncrypt($data, $rsaPrivateKey) {
        $crypted = [];
        $dataArray = str_split($data, self::MAX_ENCRYPT_BLOCK_SIZE);
        foreach($dataArray as $subData){
            $subCrypted = null;
            if(!openssl_private_encrypt($subData, $subCrypted, $rsaPrivateKey)){
                throw new \Exception(openssl_error_string());
            }
            $crypted[] = $subCrypted;
        }
        $crypted_string = implode('',$crypted);

        return base64_encode($crypted_string);
    }

    /**
     * RSA私钥分段解密
     * @param $data
     * @param $rsaPrivateKey
     * @return string
     * @throws Exception
     */
    public function privDecrypt($data, $rsaPrivateKey) {
        if (!is_string($data)){
            throw new \Exception('error data');
        }
        $encryptstr = base64_decode($data);

        $decrypted = "";
        $dataArray = str_split($encryptstr, self::MAX_DECRYPT_BLOCK_SIZE);

        foreach($dataArray as $subData){
            $subDecrypted = null;
            if(!openssl_private_decrypt($subData, $subDecrypted, $rsaPrivateKey)){
                throw new \Exception(openssl_error_string());
            }
            $decrypted .= $subDecrypted;
        }
        return $decrypted;

    }

    /**
     * RSA私钥生成签名base64编码
     * @param $content
     * @param $privateKey
     * @return string
     */
    public function sign($content, $privateKey)
    {
        openssl_sign($content, $sign, $privateKey);
        $sign = base64_encode($sign);

        return $sign;
    }


    /**
     * RSA公钥验签
     * @param $data
     * @param $sign
     * @param $rsaPublicKey
     * @return bool
     */
    public function verify($data, $sign, $rsaPublicKey) {
        $result = (openssl_verify($data, base64_decode($sign), $rsaPublicKey)===1);
        return $result;
    }

    /**
     * 由数组拆分拼接成待签名字符串
     * @param $params
     * @return string
     */
    public function getSignContent($params)
    {
        unset($params['sign']);


        // 加签方式一（键值拼接）
        $stringToBeSigned = "";
        ksort($params);
        $i = 0;
        foreach ($params as $k => $v) {
            if (false === $this->checkEmpty($v) && "@" != substr($v, 0, 1)) {
                // 转换成目标字符集
                $v = $this->characet($v, self::DEFAULT_CHARSET);
                if ($i == 0) {
                    $stringToBeSigned .= "$k" . "=" . "$v";
                } else {
                    $stringToBeSigned .= "&" . "$k" . "=" . "$v";
                }
                $i++;
            }
        }
        unset ($k, $v);

        // 加签方式一（对数组递归ksort，然后json_encode）
//        $arr = Base::mulArrksort($params);
//        $stringToBeSigned = json_encode($arr, JSON_UNESCAPED_SLASHES);

        return $stringToBeSigned;
    }

    /**
     * 检查参数空
     * @param $value
     * @return bool
     */
    public function checkEmpty($value)
    {
        if (!isset($value))
            return true;
        if ($value === null)
            return true;
        if (trim($value) === "")
            return true;
        return false;
    }


    /**
     * 设置字符编码
     * @param $data
     * @param $targetCharset
     * @return array|false|mixed|string|string[]|null
     */
    public function characet($data, $targetCharset)
    {
        if (!empty($data)) {
            $fileType = self::DEFAULT_CHARSET;
            if (strcasecmp($fileType, $targetCharset) != 0) {
                $data = mb_convert_encoding($data, $targetCharset, $fileType);
            }
        }
        return $data;
    }

    /**
     * 带签名字符串转数组
     * @param $decrypt_string
     * @return array
     */
    public static function decryptStringToArr($decrypt_string)
    {
        $dataArr = [];

        $arr = explode('&', $decrypt_string);
        foreach ($arr as $key_val) {
            $key_val_arr = explode('=', $key_val);
            $dataArr[$key_val_arr[0]] = $key_val_arr[1];
        }

        return $dataArr;
    }

    // 公钥转成PEM格式
    public function toPemPublicKey($pubKey)
    {
        $pubKey = "-----BEGIN PUBLIC KEY-----\n" .
            wordwrap($pubKey, 64, "\n", true) .
            "\n-----END PUBLIC KEY-----";
        return $pubKey;
    }

    // 公钥转成PEM格式
    public function toPemPrivateKey($privateKey)
    {
        $privateKey = "-----BEGIN PRIVATE KEY-----\n" .
            wordwrap($privateKey, 64, "\n", true) .
            "\n-----END PRIVATE KEY-----";
        return $privateKey;
    }

}
