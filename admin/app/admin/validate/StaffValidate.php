<?php

namespace app\admin\validate;

use app\admin\model\admin\Staff;
use app\common\validate\BaseValidate;

class StaffValidate extends BaseValidate
{
    //定义验证规则
    protected $rule = [
        'uuid' => 'require|checkIdExist',
        'email|邮箱' => 'require|max:255|checkEmailExist|email',
        'phone|手机号' => 'max:20|checkPhoneExist',
        'real_name|姓名' => 'max:255',
        'password|密码' => 'require|checkPassword',
        'confirm_password|确认密码' => 'requireWith:password|confirm:password',
        'company_uuid|门店UUID' => 'require|integer',
        'role_uuids|角色UUID列表' => 'array',
        'company_list|门店列表' => 'array|checkCompanyList',
    ];

    protected $message = [
        'uuid.require' => '参数错误',
        'uuid.checkIdExist' => '账号不存在',
        'email.require' => '邮箱不能为空',
        'email.checkEmailExist' => '该邮箱已在平台注册',
        'email.max' => '邮箱长度不能超过255个字符',
        'email.email' => '请确认邮箱格式',
        'phone.max' => '手机号长度不能超过20个字符',
        'phone.checkPhoneExist' => '该手机号已在平台注册',
        'real_name.max' => '姓名长度不能超过255个字符',
        'password.require' => '密码不能为空',
        'confirm_password.requireWith' => '确认密码不能为空',
        'confirm_password.confirm' => '确认密码与密码不一致',
        'company_uuid.require' => '请选择门店',
        'company_uuid.integer' => '门店UUID格式错误',
        'role_uuids.array' => '角色UUID列表格式错误',
        'company_list.array' => '门店列表格式错误',
        'company_list.checkCompanyList' => '门店列表格式错误',
    ];

    protected $scene = [
        'add' => [
            'email',
            'phone',
            'real_name',
            'password',
            'confirm_password',
            'company_uuid',
            'role_uuids',
        ],
        'edit' => [
            'uuid',
            'email',
            'phone',
            'real_name',
            'password',
            'confirm_password',
            'company_list',
        ],
        'uuid' => [
            'uuid'
        ]
    ];

    /**
     * edit 验证场景额外定义
     */
    public function sceneEdit()
    {
        return $this->only($this->scene[$this->currentScene])->remove('password', 'require');
    }

    /**
     * 验证邮箱是否存在
     */
    protected function checkEmailExist($value, $rule, $data = [])
    {
        $uuid = $data['uuid'] ?? 0;
        $user = Staff::where('email', $value)
            ->where('delete_time', 0)
            ->when($uuid, function ($q) use ($uuid) {
                $q->where('uuid', '<>', $uuid);
            })
            ->find();
        if ($user) {
            return false;
        } else {
            return true;
        }
    }

    /**
     * 验证手机号是否存在（排除空字符串）
     */
    protected function checkPhoneExist($value, $rule, $data = [])
    {
        // 空字符串不验证
        if (empty($value)) {
            return true;
        }
        
        $uuid = $data['uuid'] ?? 0;
        $user = Staff::where('phone', $value)
            ->where('phone', '<>', '')
            ->where('delete_time', 0)
            ->when($uuid, function ($q) use ($uuid) {
                $q->where('uuid', '<>', $uuid);
            })
            ->find();
        if ($user) {
            return false;
        } else {
            return true;
        }
    }

    /**
     * 验证ID是否存在
     */
    protected function checkIdExist($value, $rule, $data = [])
    {
        $user = Staff::detail($value);
        if (!$user) {
            return false;
        } else {
            return true;
        }
    }

    /**
     * 验证门店列表格式
     */
    protected function checkCompanyList($value, $rule, $data = [])
    {
        if (!is_array($value) || empty($value)) {
            return true; // 空数组允许，表示不更新门店列表
        }

        foreach ($value as $item) {
            // 每个元素必须是数组
            if (!is_array($item)) {
                return false;
            }

            // 必须包含 company_uuid
            if (!isset($item['company_uuid']) || !is_numeric($item['company_uuid'])) {
                return false;
            }

            // role_uuids 可选，但如果存在必须是数组
            if (isset($item['role_uuids']) && !is_array($item['role_uuids'])) {
                return false;
            }
        }

        return true;
    }

    /**
     * 验证密码强度
     * 密码不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种
     */
    protected function checkPassword($value, $rule, $data = [])
    {
        if (strpos($value, ' ') !== false) {
            return '密码不能包含空格';
        }
        if (strlen($value) < 8 || strlen($value) > 16) {
            return '密码长度为8-16个字符';
        }

        $count = 0;
        if (preg_match('/[a-zA-Z]/', $value)) {
            $count++;
        }
        if (preg_match('/\d/', $value)) {
            $count++;
        }
        if (preg_match('/[^a-zA-Z0-9]/', $value)) {
            $count++;
        }
        if ($count < 2) {
            return '密码必须包含字母、数字、符号中至少2种';
        }
        return true;
    }
}
