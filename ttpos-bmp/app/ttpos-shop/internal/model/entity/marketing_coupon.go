// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MarketingCoupon is the golang structure for table marketing_coupon.
type MarketingCoupon struct {
	Id             uint    `json:"id"             orm:"id"               description:""`                                          //
	Uuid           int64   `json:"uuid"           orm:"uuid"             description:"优惠券唯一ID"`                                   // 优惠券唯一ID
	Name           string  `json:"name"           orm:"name"             description:"优惠券名称"`                                     // 优惠券名称
	Sort           int     `json:"sort"           orm:"sort"             description:"排序, 1-99"`                                  // 排序, 1-99
	Type           string  `json:"type"           orm:"type"             description:"优惠券类型: deduction - 抵扣券"`                    // 优惠券类型: deduction - 抵扣券
	DeductionType  string  `json:"deductionType"  orm:"deduction_type"   description:"抵扣类型: taxed - 税后抵扣"`                        // 抵扣类型: taxed - 税后抵扣
	Amount         float64 `json:"amount"         orm:"amount"           description:"优惠券金额"`                                     // 优惠券金额
	Count          int     `json:"count"          orm:"count"            description:"优惠券数量, 最大999999"`                           // 优惠券数量, 最大999999
	DayStartTime   string  `json:"dayStartTime"   orm:"day_start_time"   description:"每日适用时段开始时间, hh:mm 格式"`                      // 每日适用时段开始时间, hh:mm 格式
	DayEndTime     string  `json:"dayEndTime"     orm:"day_end_time"     description:"每日适用时段结束时间, hh:mm 格式"`                      // 每日适用时段结束时间, hh:mm 格式
	Requirement    string  `json:"requirement"    orm:"requirement"      description:"获得优惠券所需条件: none - 都可以获取; marketing - 营销活动"` // 获得优惠券所需条件: none - 都可以获取; marketing - 营销活动
	ValidStartTime int     `json:"validStartTime" orm:"valid_start_time" description:"优惠券有效开始时间, requirement = none 时有效"`         // 优惠券有效开始时间, requirement = none 时有效
	ValidEndTime   int     `json:"validEndTime"   orm:"valid_end_time"   description:"优惠券有效结束时间, requirement = none 时有效"`         // 优惠券有效结束时间, requirement = none 时有效
	ValidDays      int     `json:"validDays"      orm:"valid_days"       description:"领取优惠券后n天内有效, requirement = marketing 时有效"`  // 领取优惠券后n天内有效, requirement = marketing 时有效
	CreateTime     int     `json:"createTime"     orm:"create_time"      description:"创建时间"`                                      // 创建时间
	UpdateTime     int     `json:"updateTime"     orm:"update_time"      description:"更新时间"`                                      // 更新时间
	DeleteTime     int     `json:"deleteTime"     orm:"delete_time"      description:"删除时间"`                                      // 删除时间
}
