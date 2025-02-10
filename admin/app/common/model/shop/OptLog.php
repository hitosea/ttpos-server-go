<?php

namespace app\common\model\shop;

use app\common\model\BaseModel;

/**
 * 登录日志模型
 */
class OptLog extends BaseModel
{
    protected $name = 'staff_operation_log';
    protected $pk = 'id';

    /**
     * 追加属性
     */
    protected $append = ['opt_log_id'];

    /**
     * 兼容ID字段
     */
    public function getOptLogIdAttr()
    {
        return $this->id ?? 0;
    }

    /**
     * 关联用户表
     */
    public function user()
    {
        return $this->belongsTo('app\\common\\model\\shop\\User', 'staff_uuid', 'uuid');
    }
}