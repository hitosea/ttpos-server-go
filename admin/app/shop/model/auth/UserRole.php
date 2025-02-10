<?php

namespace app\shop\model\auth;

use app\common\model\shop\UserRole as UserRoleModel;


/**
 * 角色模型
 */
class UserRole extends UserRoleModel
{

    public function getUserRole($where)
    {
        return $this->where($where)->column('role_uuid');
    }

    /**
     * 获取角色下的用户
     */
    public static  function getUserRoleCount($role_id)
    {
        $model = new static();
        return $model->alias('userRole')
            ->join('staff s', 'userRole.staff_uuid = s.uuid', 'left')
            ->where('userRole.role_uuid', '=', $role_id)
            ->count();
    }
}
