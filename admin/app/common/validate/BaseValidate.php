<?php

namespace app\common\validate;

use think\Validate;
use think\facade\Request;
use think\exception\ValidateException;

/**
 * 验证基类
 * Interface BaseValidate
 */
class BaseValidate extends  Validate
{
    /**
     * 验证
     * @param string $scene 场景名称
     * @param array|null $params 参数数组
     * @return array 验证后的参数数组
     * @throws \think\exception\HttpException 验证失败时抛出异常
     */
    public function goCheck($scene = "", $params = null)
    {
        //必须设置contetn-type:application/json
        $params = $params ?: Request::instance()->param();
        foreach (self::scene($scene)->rule as $key => $rule) {
            $key = explode('|', $key)[0];
            $params[$key] = $params[$key] ?? '';
        }

        try {
            if ($scene) {
                $result = self::scene($scene)->check($params);
            } else {
                $result = self::check($params);
            }
            if (!$result) {
                $error = self::geterror();
                $msg = is_array($error) ? implode(';', $error) : $error;
                throw new \think\exception\HttpException(0, $msg);
            }
            return $params;
        } catch (ValidateException $e) {
            // 验证失败 输出错误信息
            throw new \think\exception\HttpException(415, $e->getError());
        }
    }

    /**
     * 获取验证场景
     * @param string $scene 场景名称
     * @return bool
     */
    public function getAllScene($scene = '')
    {
        return isset($this->scene[$scene]) ? true : false;
    }

    /**
     * 正整数验证
     * @param $value
     * @param string $rule
     * @param string $data
     * @param string $field
     * @return bool|string
     */
    protected function isInteger($value, $rule = '', $data = '', $field = '')
    {
        if (is_numeric($value) && is_int($value + 0) && ($value + 0) > 0) {
            return true;
        }
        return $field . '必须是正整数';
    }
}
