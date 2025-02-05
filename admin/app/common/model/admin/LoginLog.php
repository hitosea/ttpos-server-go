<?php

namespace app\common\model\admin;

use app\common\model\BaseModel;

/**
 * 登录日志模型
 */
class LoginLog extends BaseModel
{
    protected $name = 'admin_user_login_log';
    protected $pk = 'id';

    /**
     * 结果兼容多语言
     */
    public function getResultAttr($value)
    {
        return $value ? __($value) : "";
    }

    /**
     * 日志列表
     */
    public function getList($param)
    {
        $username = $param['username'] ?? '';
        //
        return $this->when($username, function ($q) use ($username) {
            $q->like('username', $username);
        })
            ->order('id', 'desc')
            ->paginate($param);
    }

    /**
     * 新增登录日志
     */
    public static function add($username, $ip, $result, $admin_user_id)
    {
        $model = new self();
        $model->save([
            'username' => $username,
            'ip' => $ip,
            'result' => $result,
            'admin_user_id' => $admin_user_id
        ]);
    }
}
