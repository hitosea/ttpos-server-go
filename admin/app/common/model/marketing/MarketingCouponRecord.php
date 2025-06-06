<?php
namespace app\common\model\marketing;

use app\common\model\BaseModel;
use Carbon\Carbon;

/**
 * 优惠券模型
 */
class MarketingCouponRecord extends BaseModel
{
    protected $name = 'marketing_coupon_record';

    /**
     * 关联活动
     */
    public function coupon()
    {
        return $this->belongsTo('MarketingCoupon', 'coupon_uuid', 'uuid');
    }
} 