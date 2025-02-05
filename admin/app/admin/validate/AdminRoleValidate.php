<?php

namespace app\admin\validate;

use app\admin\model\admin\Role;
use app\common\model\admin\Access;
use app\common\validate\BaseValidate;


class AdminRoleValidate extends  BaseValidate
{
    //定义验证规则
    protected $rule =   [
        'id' => 'require|checkIdExist',
        'role_name|角色名称' => 'require|max:50|checkRoleNameExist',
        'access_id|权限' => 'require|integerArray|checkAccessIdExist',
        'sort|排序' => 'integer',
    ];

    protected $message = [
        'id.require' => 'id参数错误',
        'id.checkIdExist' => '角色不存在',
        'role_name.require' => '请输入名称',
        'role_name.max' => '角色名称最大可输入50个字符',
        'role_name.checkRoleNameExist' => '角色名称已存在',
        'access_id.require' => '请选择权限',
        'access_id.integerArray' => '权限数值必须为整数',
        'access_id.checkAccessIdExist' => '权限ID不存在',
    ];

    protected $scene = [
        'add' => ['role_name', 'access_id', 'sort'],
        'edit' => ['id', 'role_name', 'access_id', 'sort'],
        'id' => ['id']
    ];

    // 验证角色名是否存在
    protected function checkRoleNameExist($value, $rule, $data = [])
    {
        if (Role::where('role_name', $value,)->where('id', '<>', $data['id'] ?? 0)->count()) {
            return false;
        } else {
            return true;
        }
    }

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
