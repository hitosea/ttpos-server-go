<?php

namespace app\admin\validate;

use app\admin\model\admin\Role;
use app\common\model\admin\Access;
use app\common\validate\BaseValidate;


class PaymentAppValidate extends  BaseValidate
{
    // 定义验证规则
    protected $rule = [
        'shop_supplier_id' => 'require|integer',
        'll_merchant_id' => 'require|max:255',
        'll_public_key' => 'require',
        'll_merchant_private_key' => 'require',
        'll_token' => 'require|max:255',
    ];

    // 定义错误信息
    protected $message = [
        'shop_supplier_id.require' => '请输入商家ID',
        'shop_supplier_id.integer' => '商家ID必须为整数',
        'll_merchant_id.require' => '请输入商户号',
        'll_merchant_id.max' => '商户号不能超过255个字符',
        'll_public_key.require' => '请输入LianLianpay公钥',
        'll_merchant_private_key.require' => '请输入商户私钥',
        'll_token.require' => '请输入商户Token',
        'll_token.max' => '商户Token不能超过255个字符',
    ];

    // 定义验证场景
    protected $scene = [
        'add' => ['shop_supplier_id', 'll_white_ip', 'll_merchant_id', 'll_public_key', 'll_merchant_private_key', 'll_token'],
        'id' => ['shop_supplier_id']
    ];

    // 验证id是否存在
    protected function checkIdExist($value, $rule, $data = [])
    {
        $user = Role::find($value);
        if (!$user) {
            return false;
        } else {
            return true;
        }
    }

    // 自定义验证规则方法
    protected function checkAccessIdExist($value, $rule, $data = [])
    {
        if (!is_array($value)) {
            return false;
        }
        $count = Access::whereIn('id', $value)->count();
        if ($count < count($value)) {
            return false;
        }
        return true;
    }
}
