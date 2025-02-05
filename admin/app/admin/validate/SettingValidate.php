<?php

namespace app\admin\validate;

use app\common\validate\BaseValidate;


class SettingValidate extends  BaseValidate
{
    //定义验证规则
    protected $rule =   [
        'brand_name|品牌名称' => 'require|max:50',
        'brand_logo|品牌LOGO 正方型' => 'require|max:5000',
        'brand_logo_long|品牌LOGO 长方型' => 'require|max:5000',
        'browser_logo|浏览器LOGO' => 'require|max:5000',
        'browser_title|浏览器标题' => 'require|max:50',
        'expiration_reminder|到期提醒' => 'require|integer',
        'member_default_avatar|会员默认头像' => 'require|string',
        'auth_code_bind_validity_period|授权码绑定有效期' => 'integer',
    ];

    protected $message  = [];

    protected $scene = [
        'edit' => [
            'brand_name',
            'brand_logo',
            'brand_logo_long',
            'browser_logo',
            'browser_title',
            'auth_code_bind_validity_period',
            // 'expiration_reminder',
            // 'member_default_avatar',
        ]
    ];
}
