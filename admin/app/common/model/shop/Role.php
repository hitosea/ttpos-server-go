<?php

namespace app\common\model\shop;


use app\common\model\BaseModel;

/**
 * 应用模型
 */
class Role extends BaseModel
{
    protected $name = 'role';
    protected $pk = 'id';

    /**
     * 追加属性
     */
    protected $append = ['role_id'];

    /**
     * 兼容字段
     */
    public function getRoleIdAttr()
    {
        return $this->uuid ?: 0;
    }

    /**
     * 关联权限
     * @return \think\model\relation\HasMany
     */
    public function access()
    {
        return $this->hasMany('RoleAccess', 'role_uuid', 'uuid');
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
        return static::with(['access'])->field('*, name as role_name')->where('uuid', $role_id)->find();
    }
}
