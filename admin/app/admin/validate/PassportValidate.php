<?php

namespace app\admin\validate;

use app\common\validate\BaseValidate;


class PassportValidate extends  BaseValidate
{
    //定义验证规则
    protected $rule =   [
        'password|密码' =>  'require',
        'username|用户名' => 'require|max:20',
        'code|验证码' =>  'require',
    ];

    protected $message = [];

    protected $scene = [
        'login' => [
            'username',
            'password',
            'code',
        ]
    ];
}
