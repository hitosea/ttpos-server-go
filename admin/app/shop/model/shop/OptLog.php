<?php

namespace app\shop\model\shop;

use app\common\model\shop\OptLog as OptLogModel;

/**
 * 后台管理员登录日志模型
 */
class OptLog extends OptLogModel
{
    /**
     * 获取列表数据
     */
    public function getList($params)
    {
        $model = $this;
        //
        if (isset($params['username']) && $params['username'] != '') {
            $model = $model->like('user.username|user.real_name', $params['username']);
        }
        // 查询列表数据
        return $model->alias('log')->field(['log.*', 'log.source as browser', 'user.username as user_name', 'IF(user.real_name = "", user.username, user.real_name) as real_name'])
            ->join('staff user', 'user.uuid = log.staff_uuid', 'left')
            ->order(['log.create_time' => 'desc'])
            ->paginate($params);
    }
}
