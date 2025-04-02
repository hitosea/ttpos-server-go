<?php

namespace app\common\service\qrcode;

/**
 * 二维码验证
 */
class AuthService
{
    // 生成Token
    public function generateToken($data)
    {
        // 将数组转换为字符串
        $dataString = json_encode($data);
        $hash = md5($dataString);
        // 将散列值与原始数据字符串拼接
        $token = $hash . '.' . base64_encode($dataString);

        return base64_encode($token);
    }

    // 解码Token
    public function decodeToken($token)
    {
        $token = base64_decode($token);
        // 分割token以获取散列值和数据部分
        $parts = explode('.', $token, 2);
        if (count($parts) !== 2) {
            return null; // 无效的token格式
        }
        list($hash, $data) = $parts;
        // 解码base64编码的数据
        $dataString = base64_decode($data);
        // 重新计算数据的散列值以进行验证
        $newHash = hash('sha256', $dataString);
        // 比较散列值以验证token的完整性
        if (hash_equals($newHash, $hash)) {
            // 解码JSON字符串为数组
            return json_decode($dataString, true);
        } else {
            return null; // 散列值不匹配，token可能已被篡改
        }
    }
}
