<?php
namespace app\common\model\marketing;

use app\common\model\BaseModel;

/**
 * 邀请有礼奖品模型
 */
class MarketingActivityPrize extends BaseModel
{
    protected $name = 'marketing_activity_prize';

    // 追加属性
    protected $append = [
        'prize_type_text'
    ];
    
    /**
     * 获取奖品类型文字
     */
    public function getPrizeTypeTextAttr($value, $data)
    {
        $types = [
            1 => '优惠券',
            2 => '未知'
        ];
        return $types[$data['prize_type']] ?? '';
    }
    
    /**
     * 关联活动
     */
    public function activity()
    {
        return $this->belongsTo('MarketingActivity', 'activity_uuid', 'uuid');
    }

    /**
     * 关联优惠券
     */
    public function couponName()
    {
        return $this->belongsTo('MarketingCoupon', 'prize_uuid', 'uuid')->field('uuid, name')->bind(['coupon_name' => 'name']);
    }

    /**
     * 关联优惠券
     */
    public function coupon()
    {
        return $this->belongsTo('MarketingCoupon', 'prize_uuid', 'uuid');
    }
} 