<?php

namespace app\common\model\shop;


use app\common\model\BaseModel;

/**
 * 应用模型
 */
class RoleAccess extends BaseModel
{
    protected $name = 'role_access';
    protected $pk = 'id';

    /**
     * 追加属性
     */
    protected $append = ['role_id', 'access_id'];

    /**
     * 兼容字段
     */
    public function getRoleIdAttr()
    {
        return $this->role_uuid ?: 0;
    }

    /**
     * 兼容字段
     */
    public function getAccessIdAttr()
    {
        return $this->access_uuid ?: 0;
    }

    /**
     * 权限
     * @return \think\model\relation\BelongsTo
     */
    public function shopAccess()
    {
        return $this->belongsTo('Access', 'access_uuid', 'uuid');
    }

    /**
     * 获取指定角色的所有权限id
     * @param int|array $role_id 角色id (支持数组)
     */
    public static function getAccessIds($role_id)
    {
        $roleIds = is_array($role_id) ? $role_id : [(int)$role_id];
        return (new self)->where('role_uuid', 'in', $roleIds)->column('access_uuid');
    }
}
