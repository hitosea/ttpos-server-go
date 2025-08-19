// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MemberCouponUseRecord is the golang structure of table ttpos_member_coupon_use_record for DAO operations like Where/Data.
type MemberCouponUseRecord struct {
	g.Meta         `orm:"table:ttpos_member_coupon_use_record, do:true"`
	Id             interface{} //
	Uuid           interface{} // 唯一ID
	MemberUuid     interface{} // 会员uuid
	CouponUuid     interface{} // 优惠券uuid
	UseOrderUuid   interface{} // 优惠券使用订单uuid
	UseOrderAmount interface{} // 优惠券使用订单金额
	CreateTime     interface{} // 创建时间
	UpdateTime     interface{} // 更新时间
	DeleteTime     interface{} // 删除时间
}
