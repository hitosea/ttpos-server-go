<?php

namespace app\common\model\admin;


use app\common\model\BaseModel;

/**
 * 应用模型
 */
class UserRole extends BaseModel
{
    protected $name = 'admin_user_role';
    protected $pk = 'id';

    /**
     * 关联角色
     * @return \think\model\relation\BelongsTo
     */
    public function role()
    {
        return $this->belongsTo('Role', 'role_id', 'id');
    }

    /**
     * 获取指定管理员的所有角色id
     * @param $admin_user_id
     * @return array
     */
    public static function getRoleIds($admin_user_id)
    {
        return (new self)->where('admin_user_id', '=', $admin_user_id)->column('role_id');
    }
}
