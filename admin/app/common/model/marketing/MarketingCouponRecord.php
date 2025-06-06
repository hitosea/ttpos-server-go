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

    // 记录类型：1-首次添加、2-调整添加、3-调整扣减
    const RecordTypeCreate = 1; // 首次添加
    const RecordTypeIncrease = 2; // 调整添加
    const RecordTypeDecrease = 3; // 调整扣减

    /**
     * 关联优惠券
     */
    public function coupon()
    {
        return $this->belongsTo('MarketingCoupon', 'coupon_uuid', 'uuid');
    }
} 