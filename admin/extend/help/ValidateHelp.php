<?php

namespace help;


// 验证帮助
class ValidateHelp
{
    /**
     * 验证手机号
     * @param string $phone 手机号
     * @return bool|string 成功返回 true，失败返回错误信息
     */
    public static function validatePhone($phone)
    {
        // 检查是否为空
        if (empty($phone)) {
            return '请输入手机号';
        }
        // 检查邮箱长度
        if (strlen($phone) > 20) {
            return '手机号长度不能超过20位';
        }
        return true;
    }

    /**
     * 验证邮箱
     *
     * @param string $email 邮箱
     * @return bool|string 成功返回 true，失败返回错误信息
     */
    public static function validateEmail($email)
    {
        // 检查是否为空
        if (empty($email)) {
            return '请输入邮箱';
        }
        // 检查邮箱格式
        if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
            return '请确认邮箱格式';
        }
        // 检查邮箱长度
        if (strlen($email) > 64) {
            return '邮箱长度不能超过64位';
        }
        return true;
    }

    /**
     * 函数用于检查数组或 JSON 字符串中是否存在空值
     *
     * @param array|string $input
     * @return bool
     */
    public static function hasEmptyValue($input): bool
    {
        if (is_string($input)) {
            $input = json_decode($input, true);
            if ($input === null) {
                return true;
            }
        }
        if (!is_array($input)) {
            return true;
        }
        // 检查数组中是否存在空值
        return in_array("", array_map('trim', $input), true);
    }

    /**
     * 函数用于检查数组或 JSON 字符串中是否超过指定长度
     *
     * @param array|string $input
     * @param int $length
     * @return array|bool
     */
    public static function hasExceedLength($input, int $length): array|bool
    {
        $result = [];
        if (is_string($input)) {
            $input = json_decode($input, true);
            if (json_last_error() !== JSON_ERROR_NONE) {
                return ['error' => 'Invalid JSON string'];
            }
        }
        if (!is_array($input)) {
            return ['error' => 'Input is not an array'];
        }
        foreach ($input as $key => $value) {
            $result[$key] = mb_strlen($value) > $length;
        }
        $hasError = in_array(true, $result, true);
        return $hasError ? [true, $result] : false;
    }

    /**
     * 验证是否为4-16位纯数字
     *
     * @param string $str
     * @return bool
     */
    public static function validateNumber($str)
    {
        if ($str === '' || $str === null) {
            return false;
        }
        return preg_match('/^\d{4,16}$/', $str) === 1;
    }

    /**
     * 验证输入的字符串是否只包含英文字符和数字，并且长度不超过50
     *
     * @param string $input
     * @return bool
     */
    public static function validateAlphanumeric($input)
    {
        // 使用正则表达式检查输入的字符串是否只包含英文字符和数字，并且长度不超过50
        return preg_match('/^[a-zA-Z0-9]{1,50}$/', $input) === 1;
    }

    /**
     * 密码必须是 不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种
     */
    public static function validateAlphaPassword($input)
    {
        // 正则表达式解释：
        // ^(?:(?=.*\d)(?=.*[a-zA-Z])|(?=.*\d)(?=.*[\W_])|(?=.*[a-zA-Z])(?=.*[\W_]))(?!.*\s).{8,16}$
        // ^ - 字符串开头
        // (?=.*[a-zA-Z]) - 必须包含至少一个字母
        // (?=.*\d|.*[-+_!@#$%^&*.,?]) - 必须包含至少一个数字或符号
        // (?=.*[-+_!@#$%^&*.,?]|.*\d) - 必须包含至少一个符号或数字
        // (?!.*\s) - 不能包含空格
        // .{8,16}$ - 长度为 8 到 16 个字符
        return preg_match('/^(?:(?=.*\d)(?=.*[a-zA-Z])|(?=.*\d)(?=.*[\W_])|(?=.*[a-zA-Z])(?=.*[\W_]))(?!.*\s).{8,16}$/', $input) === 1;
    }

    /**
     * 验证会员卡号，1~48位字符，允许输入字母和数字，不允许输入特殊字符
     *
     * @param string $input
     * @return bool
     */
    public static function validateCardNumber($input)
    {
        return preg_match('/^[a-zA-Z0-9]{1,48}$/', $input) === 1;
    }

    /**
     * 验证权限密码，必须为 4-8 位数字
     *
     * @param string $input
     * @return bool
     */
    public static function validatePermissionPassword($input)
    {
        return preg_match('/^\d{4,8}$/', $input) === 1;
    }
}
