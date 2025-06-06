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
            return Carbon::create($data['valid_start_time'])->format("Y-m-d").'~'.Carbon::create($data['valid_end_time'])->format("Y-m-d");
        } else {
            return sprintf("活动奖励后%d个自然日内有效", $data['valid_days']);
        }
    }

    // 优惠券适用时间
    public function getValidDayTimeRangeAttr($value, $data)
    {
        return sprintf("%s~%s", $data['day_start_time'], $data['day_end_time']);
    }
} 