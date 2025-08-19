// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MemberCoupon is the golang structure of table ttpos_member_coupon for DAO operations like Where/Data.
type MemberCoupon struct {
	g.Meta         `orm:"table:ttpos_member_coupon, do:true"`
	Id             interface{} //
	Uuid           interface{} // 唯一ID
	MemberUuid     interface{} // 会员uuid
	CouponUuid     interface{} // 优惠券uuid
	Name           interface{} // 优惠券名称
	DeductionType  interface{} // 抵扣类型: taxed - 税后抵扣
	Type           interface{} // 优惠券类型: deduction - 抵扣券
	DayStartTime   interface{} // 每日适用时段开始时间, hh:mm 格式
	DayEndTime     interface{} // 每日适用时段结束时间, hh:mm 格式
	ValidStartTime interface{} // 优惠券有效开始时间, requirement = none 时有效
	ValidEndTime   interface{} // 优惠券有效结束时间, requirement = none 时有效
	Amount         interface{} // 优惠券面值
	Status         interface{} // 优惠券状态 0未使用 1已使用
	StartTime      interface{} // 优惠券开始时间
	EndTime        interface{} // 优惠券结束时间
	UseTime        interface{} // 优惠券使用时间
	DeleteTime     interface{} // 删除时间
	CreateTime     interface{} // 创建时间
	UpdateTime     interface{} // 更新时间
}
