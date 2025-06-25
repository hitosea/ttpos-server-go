package constant

// 记录类型：1-首次添加、2-调整添加、3-调整扣减、4-活动扣减、5、奖励领取（冻结）、6、核销扣减
const (
	CouponRecordTypeCreate            = iota + 1 // 首次添加
	CouponRecordTypeIncrease                     // 调整添加
	CouponRecordTypeDecrease                     // 调整扣减
	CouponRecordTypeActivityDeduction            // 活动扣减
	CouponRecordTypeBonus                        // 奖励领取（冻结）
	CouponRecordTypeUsed                         // 核销扣减
)

// requirement
const (
	CouponRequirementNone   = "none"      // 任何人可用的优惠券
	CouponRequirementMember = "marketing" // 会员
)
