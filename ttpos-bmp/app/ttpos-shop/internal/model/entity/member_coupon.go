// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MemberCoupon is the golang structure for table member_coupon.
type MemberCoupon struct {
	Id             uint    `json:"id"             orm:"id"               description:""`                                  //
	Uuid           int64   `json:"uuid"           orm:"uuid"             description:"唯一ID"`                              // 唯一ID
	MemberUuid     int64   `json:"memberUuid"     orm:"member_uuid"      description:"会员uuid"`                            // 会员uuid
	CouponUuid     int64   `json:"couponUuid"     orm:"coupon_uuid"      description:"优惠券uuid"`                           // 优惠券uuid
	Name           string  `json:"name"           orm:"name"             description:"优惠券名称"`                             // 优惠券名称
	DeductionType  string  `json:"deductionType"  orm:"deduction_type"   description:"抵扣类型: taxed - 税后抵扣"`                // 抵扣类型: taxed - 税后抵扣
	Type           string  `json:"type"           orm:"type"             description:"优惠券类型: deduction - 抵扣券"`            // 优惠券类型: deduction - 抵扣券
	DayStartTime   string  `json:"dayStartTime"   orm:"day_start_time"   description:"每日适用时段开始时间, hh:mm 格式"`              // 每日适用时段开始时间, hh:mm 格式
	DayEndTime     string  `json:"dayEndTime"     orm:"day_end_time"     description:"每日适用时段结束时间, hh:mm 格式"`              // 每日适用时段结束时间, hh:mm 格式
	ValidStartTime int     `json:"validStartTime" orm:"valid_start_time" description:"优惠券有效开始时间, requirement = none 时有效"` // 优惠券有效开始时间, requirement = none 时有效
	ValidEndTime   int     `json:"validEndTime"   orm:"valid_end_time"   description:"优惠券有效结束时间, requirement = none 时有效"` // 优惠券有效结束时间, requirement = none 时有效
	Amount         float64 `json:"amount"         orm:"amount"           description:"优惠券面值"`                             // 优惠券面值
	Status         int     `json:"status"         orm:"status"           description:"优惠券状态 0未使用 1已使用"`                   // 优惠券状态 0未使用 1已使用
	StartTime      int     `json:"startTime"      orm:"start_time"       description:"优惠券开始时间"`                           // 优惠券开始时间
	EndTime        int     `json:"endTime"        orm:"end_time"         description:"优惠券结束时间"`                           // 优惠券结束时间
	UseTime        int     `json:"useTime"        orm:"use_time"         description:"优惠券使用时间"`                           // 优惠券使用时间
	DeleteTime     int     `json:"deleteTime"     orm:"delete_time"      description:"删除时间"`                              // 删除时间
	CreateTime     int     `json:"createTime"     orm:"create_time"      description:"创建时间"`                              // 创建时间
	UpdateTime     int     `json:"updateTime"     orm:"update_time"      description:"更新时间"`                              // 更新时间
}
