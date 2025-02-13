<?php

namespace app\common\model\user;

use app\common\model\BaseModel;
use app\common\enum\user\pointsLog\PointsLogSceneEnum;

class PointsLog extends BaseModel
{
    protected $name = 'member_point_log';
    protected $pk = 'id';

    /**
     * 追加属性
     */
    protected $append = ['log_id', 'user_id'];

    /**
     * 兼容字段
     */
    public function getLogIdAttr($value, $data)
    {
        return $this->id ?: 0;
    }
    public function getUserIdAttr($value, $data)
    {
        return $this->member_uuid ?: 0;
    }

    /**
     * 获取当前模型属性
     */
    public static function getAttributes()
    {
        return [
            // 充值方式
            'scene' => PointsLogSceneEnum::data(),
        ];
    }

    /**
     * 积分变动场景
     */
    public function getSceneAttr($value)
    {
        try {
            return ['text' => PointsLogSceneEnum::data()[$value]['name'], 'value' => $value];
        } catch (\Exception $e) {
            return ['text' => '-', 'value' => $value];
        }
    }

    /**
     * 关联会员记录表
     */
    public function user()
    {
        $module = self::getCalledModule() ?: 'common';
        return $this->belongsTo("app\\{$module}\\model\\user\\User", 'member_uuid', 'uuid')->field('*, nickname as nickName')->hidden(['password']);
    }

    /**
     * 新增记录
     */
    public static function add($data)
    {
        $static = new static;
        $static->save($data);
    }

    /**
     * 新增记录 (批量)
     */
    public function onBatchAdd($saveData)
    {
        return $this->saveAll($saveData);
    }
}
