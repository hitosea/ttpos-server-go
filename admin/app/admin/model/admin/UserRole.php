<?php

namespace app\admin\model\admin;

use app\common\model\admin\UserRole as UserRoleModel;


/**
 * 角色模型
 */
class UserRole extends UserRoleModel
{

    /**
     * 根据条件获取角色id
     * @param array $where
     * @return array
     */
    public function getUserRole($where)
    {
        return $this->where($where)->column('role_id');
    }

    /**
     * 获取角色下的用户
     */
    public static  function getUserRoleCount($role_id)
    {
        $model = new static();
        return $model->alias('userRole')
            ->join('admin_user', 'userRole.admin_user_id = admin_user.admin_user_id', 'left')
            ->where('userRole.role_id', '=', $role_id)
            ->count();
    }
}
