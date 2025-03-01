<?php

namespace app\common\model\user;

use app\common\model\BaseModel;
use app\common\library\helper;

/**
 * 用户等级模型
 */
class Grade extends BaseModel
{
    protected $name = 'member_level';
    protected $pk = 'id';

    /**
     * 追加属性
     */
    protected $append = ['grade_id', 'weight', 'equity', 'open_points', 'upgrade_points'];

    /**
     * 兼容字段
     */
    public function getGradeIdAttr()
    {
        return $this->uuid ?: 0;
    }
    public function getWeightAttr()
    {
        return $this->priority ?: 0;
    }
    public function getEquityAttr()
    {
        return $this->discount ?: '';
    }
    public function getOpenPointsAttr()
    {
        return $this->open_point ?: 0;
    }
    public function getUpgradePointsAttr()
    {
        return $this->upgrade_point ?: 0;
    }

    /**
     * 用户等级模型初始化
     */
    public static function init()
    {
        parent::init();
    }

    /**
     * 备注信息翻译
     */
    public function getRemarkAttr($value, $data)
    {
        if ($data['is_default'] == 1) {
            return __($value);
        }
        //
        $remark = '';
        if ($data['open_money'] == 1) {
            $money = Helper::amountPermillage($data['upgrade_money']);
            $remark .= __("会员消费满") . " {$money} " . __("可升级到此等级");
        }
        if ($data['open_point'] == 1) {
            if (!empty($remark)) {
                $remark .= '\r\n';
            }
            $remark .= __("会员积分满") . " {$data['upgrade_point']} " . __("可升级到此等级");
        }
        return $remark;
    }

    /**
     * 设置折扣率
     */
    public function setDiscountAttr($value)
    {
        return helper::bcdiv($value, 100, 4);
    }

    /**
     * 获取折扣率
     */
    public function getDiscountAttr($value, $data)
    {
        return floatval(helper::bcmul($value, 100));
    }

    /**
     * 获取详情
     */
    public static function detail($grade_id)
    {
        return self::where('uuid', $grade_id)->find();
    }

    /**
     * 获取列表记录
     */
    public function getLists()
    {
        return $this->field('uuid, name, priority, create_time')->order(['priority' => 'asc', 'create_time' => 'asc'])->select();
    }

    /**
     * 获取可用的会员等级列表
     */
    public static function getUsableList($appId = null)
    {
        $model = new static;
        $appId = $appId ? $appId : $model::$app_id;
        return $model->order(['priority' => 'asc', 'create_time' => 'asc'])->select();
    }

    /**
     * 获取可用的会员等级列表(升级使用)
     */
    public static function getUsable($appId = null)
    {
        $model = new static;
        $appId = $appId ? $appId : $model::$app_id;
        return $model->order(['priority' => 'desc'])->select();
    }

    /**
     * 获取默认等级id
     */
    public static function getDefaultGradeId()
    {
        $grade = self::where('is_default', '=', 1)->find();
        return $grade['grade_id'];
    }
}
