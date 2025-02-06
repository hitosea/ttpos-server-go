<?php

namespace app\admin\validate;

use app\admin\model\admin\User;
use app\common\validate\BaseValidate;


class AdminUserValidate extends  BaseValidate
{
    //定义验证规则
    protected $rule =   [
        'admin_user_id' => 'require|checkIdExist',
        'user_name|邮箱' => 'require|max:64|checkUserNameExist|email',
        'phone|手机号' => 'require|max:20|checkPhoneExist',
        'password|密码' =>  'require|checkPassword',
        'confirm_password|确认密码' => 'requireWith:password|confirm:password',
        'real_name|姓名' => 'require|string|max:50',
        'role_id|角色' => 'require|integerArray',
        //
        'pass|新密码' =>  'require|checkPassword',
        'oldPass|原密码' =>  'require',
        'checkPass|确认密码' =>  'require|confirm:pass',
    ];

    protected $message = [
        'admin_user_id.require' => '参数错误',
        'admin_user_id.checkIdExist' => '用户不存在',
        'user_name.checkUserNameExist' => '用户名已存在',
        'user_name.max' => '邮箱长度不能超过64个字符',
        'user_name.email' => '请确认邮箱格式',
        'phone.require' => '请输入手机号',
        'phone.max' => '手机号长度不能超过20个字符',
        'phone.checkPhoneExist' => '手机号已存在',
        'password.require' => '密码不能为空',
        'confirm_password.requireWith' => '确认密码不能为空',
        'confirm_password.confirm' => '确认密码与密码不一致',
        //
        'checkPass.require' => '确认密码不能为空',
        'checkPass.confirm' => '确认密码与密码不一致',
    ];

    protected $scene = [
        'add' => [
            'user_name',
            'phone',
            'password',
            'confirm_password',
            'real_name',
            'role_id',
        ],
        'edit' => [
            'admin_user_id',
            'user_name',
            'phone',
            'password',
            'confirm_password',
            'real_name',
            'role_id',
        ],
        'pass' => [
            'pass',
            'checkPass',
            'oldPass',
        ],
        'id' => [
            'admin_user_id'
        ]
    ];

    /**
     * edit 验证场景额外定义
     *
     */
    public function sceneEdit()
    {
        return $this->only($this->scene[$this->currentScene])->remove('password', 'require');
    }

    /**
     * 验证username是否存在
     */
    protected function checkUserNameExist($value, $rule, $data = [])
    {
        $id = $data['admin_user_id'] ?? 0;
        $user = User::where('username', $value)
            ->when($id, function ($q) use ($id) {
                $q->where('admin_user_id', '<>', $id);
            })
            ->find();
        if ($user) {
            return false;
        } else {
            return true;
        }
    }

    /**
     *  验证手机号是否存在
     */
    protected function checkPhoneExist($value, $rule, $data = [])
    {
        $id = $data['admin_user_id'] ?? 0;
        $user = User::where('phone', $value)
            ->when($id, function ($q) use ($id) {
                $q->where('admin_user_id', '<>', $id);
            })
            ->find();
        if ($user) {
            return false;
        } else {
            return true;
        }
    }

    /**
     * 验证id是否存在
     *
     */
    protected function checkIdExist($value, $rule, $data = [])
    {
        $user = User::find($value);
        if (!$user) {
            return false;
        } else {
            return true;
        }
    }
}
