<?php

namespace app\common\model\shop;

use app\common\model\BaseModel;

/**
 * 登录日志模型
 */
class LoginLog extends BaseModel
{
    protected $name = 'staff_login_log';
    protected $pk = 'id';

    /**
     * 结果兼容多语言
     */
    public function getResultAttr($value)
    {
        return __($value);
    }

    /**
     * 新增登录日志
     */
    public static function add($username, $ip, $result, $app_id)
    {
        (new self)->save([
            'uuid' => createUuid(),
            'staff_uuid' => $app_id,
            'username' => $username,
            'ip' => $ip,
            'result' => $result
        ]);
    }
}
