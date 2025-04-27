<?php

namespace app\common\model_old\admin;

use app\common\model_old\BaseModel;
use hg\apidoc\annotation as Apidoc;

/**
 * 登录日志模型
 */
class OptLog extends BaseModel
{
    protected $name = 'admin_user_opt_log';
    protected $pk = 'id';

    /**
     * 关联用户表
     */
    public function user()
    {
        return $this->belongsTo('app\\common\\model\\admin\\User', 'admin_user_id', 'admin_user_id');
    }

    /**
     * 日志列表
     * @Apidoc\AddField("user_name",type="string",default="",desc="用户名")
     * @Apidoc\AddField("real_name",type="string",default="",desc="姓名")
     */
    public function getList($param)
    {
        $username = $param['username'] ?? '';
        //
        return $this->alias('log')
            ->leftJoin('admin_user user', 'user.admin_user_id = log.admin_user_id')
            ->field('log.*')
            ->field('user.user_name, IF(user.real_name = "", user.user_name, user.real_name) as real_name')
            ->when($username, function ($q) use ($username) {
                $q->like('user.user_name', $username);
                $q->orLike('user.real_name', $username);
            })
            ->order('log.id', 'desc')
            ->paginate($param);
    }
}
