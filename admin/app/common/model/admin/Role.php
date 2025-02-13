<?php

namespace app\common\model\admin;


use app\common\model\BaseModel;

/**
 * 应用模型
 */
class Role extends BaseModel
{
    protected $name = 'admin_role';
    protected $pk = 'id';

    /**
     * 关联权限
     * @return \think\model\relation\HasMany
     */
    public function access()
    {
        return $this->hasMany('RoleAccess', 'role_id', 'id');
    }

    /**
     * 获取详情
     * @param $where
     * @return array|\think\Model|null
     * @throws \think\db\exception\DataNotFoundException
     * @throws \think\db\exception\DbException
     * @throws \think\db\exception\ModelNotFoundException
     */
    public static function detail($role_id)
    {
        return static::with(['access'])->where('id', $role_id)->find();
    }
}
