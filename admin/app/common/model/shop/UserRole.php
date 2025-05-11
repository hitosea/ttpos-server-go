<?php

namespace app\common\model\shop;


use app\common\model\BaseModel;

/**
 * 应用模型
 */
class UserRole extends BaseModel
{
    protected $name = 'staff_role';
    protected $pk = 'id';

    /**
     * 关联角色
     * @return \think\model\relation\BelongsTo
     */
    public function role()
    {
        return $this->belongsTo('Role', 'role_uuid', 'uuid')->field('*, name as role_name');
    }

    /**
     * 获取指定管理员的所有角色id
     * @param $shop_user_id
     * @return array
     */
    public static function getRoleIds($shop_user_id)
    {
        return (new self)->where("staff_uuid","=", $shop_user_id)->column('role_uuid');
    }
}
