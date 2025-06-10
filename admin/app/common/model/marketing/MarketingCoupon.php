<?php
namespace app\common\model\marketing;

use app\common\model\BaseModel;
use Carbon\Carbon;

/**
 * 优惠券模型
 */
class MarketingCoupon extends BaseModel
{
    protected $name = 'marketing_coupon';

    // 追加属性
    protected $append = [
        'valid_date',
        'valid_day_time_range',
    ];


    // 优惠券有效日期
    public function getValidDateAttr($value, $data)
    {
        if ($data['requirement'] == "none") {
            return date("Y-m-d", $data['valid_start_time']).'~'.date("Y-m-d", $data['valid_end_time']);
        } else {
            return sprintf(__("活动奖励后%d个自然日内有效"), $data['valid_days']);
        }
    }

    // 优惠券有效日期开始时间
    public function getValidStartTimeAttr($value, $data)
    {
        if ($data['requirement'] == "none") {
            return date("Y-m-d", $data['valid_start_time']);
        } else {
            return "";
        }
    }

    // 优惠券有效日期结束时间
    public function getValidEndTimeAttr($value, $data)
    {
        if ($data['requirement'] == "none") {
            return date("Y-m-d", $data['valid_end_time']);
        } else {
            return "";
        }
    }


    // 优惠券适用时间
    public function getValidDayTimeRangeAttr($value, $data)
    {
        return sprintf("%s~%s", $data['day_start_time'], $data['day_end_time']);
    }
} 