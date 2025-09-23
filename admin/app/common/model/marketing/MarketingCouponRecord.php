<?php
namespace app\common\model\marketing;

use app\common\model\BaseModel;
use Carbon\Carbon;

/**
 * 优惠券记录模型
 */
class MarketingCouponRecord extends BaseModel
{
    protected $name = 'marketing_coupon_record';

    // 记录类型：1-首次添加、2-调整添加、3-调整扣减、4-活动扣减、5、奖励领取（冻结）、6、核销扣减、7、删除优惠券
    const RecordTypeCreate = 1; // 首次添加
    const RecordTypeIncrease = 2; // 调整添加
    const RecordTypeDecrease = 3; // 调整扣减
    const RecordTypeActivityDeduction = 4; // 活动扣减
    const RecordTypeBonus = 5; // 奖励领取（冻结）
    const RecordTypeUsed = 6; // 核销扣减
    const RecordTypeDelete = 7; // 删除优惠券

    /**
     * 关联优惠券
     */
    public function coupon()
    {
        return $this->belongsTo('MarketingCoupon', 'coupon_uuid', 'uuid');
    }
} 