// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MarketingCoupon is the golang structure of table ttpos_marketing_coupon for DAO operations like Where/Data.
type MarketingCoupon struct {
	g.Meta         `orm:"table:ttpos_marketing_coupon, do:true"`
	Id             interface{} //
	Uuid           interface{} // 优惠券唯一ID
	Name           interface{} // 优惠券名称
	Sort           interface{} // 排序, 1-99
	Type           interface{} // 优惠券类型: deduction - 抵扣券
	DeductionType  interface{} // 抵扣类型: taxed - 税后抵扣
	Amount         interface{} // 优惠券金额
	Count          interface{} // 优惠券数量, 最大999999
	DayStartTime   interface{} // 每日适用时段开始时间, hh:mm 格式
	DayEndTime     interface{} // 每日适用时段结束时间, hh:mm 格式
	Requirement    interface{} // 获得优惠券所需条件: none - 都可以获取; marketing - 营销活动
	ValidStartTime interface{} // 优惠券有效开始时间, requirement = none 时有效
	ValidEndTime   interface{} // 优惠券有效结束时间, requirement = none 时有效
	ValidDays      interface{} // 领取优惠券后n天内有效, requirement = marketing 时有效
	CreateTime     interface{} // 创建时间
	UpdateTime     interface{} // 更新时间
	DeleteTime     interface{} // 删除时间
}
