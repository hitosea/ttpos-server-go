<?php

namespace app\common\model\admin;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 * 超管后台用户模型
 */
class User extends BaseModel
{
    use SoftDelete;

    protected $name = 'admin_user';
    protected $pk = 'admin_user_id';
    protected $deleteTime = 'delete_time';

    /**
     * 关联角色表
     */
    public function role()
    {
        return $this->belongsToMany('app\\common\\model\\admin\\Role', 'app\\common\\model\\admin\\UserRole');
    }

    /**
     * 关联用户角色表
     */
    public function userRole()
    {
        return $this->hasMany('app\\common\\model\\admin\\UserRole', 'admin_user_id', 'admin_user_id');
    }

    /**
     * 启用禁用
     * @return bool
     */
    public function updateStatus()
    {
        return $this->allowField(['status'])->save([
            'status' => $this['status'] == 1 ? 0 : 1,
        ]);
    }
}
