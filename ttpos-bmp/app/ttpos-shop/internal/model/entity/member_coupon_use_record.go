// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MemberCouponUseRecord is the golang structure for table member_coupon_use_record.
type MemberCouponUseRecord struct {
	Id             uint    `json:"id"             orm:"id"               description:""`            //
	Uuid           int64   `json:"uuid"           orm:"uuid"             description:"唯一ID"`        // 唯一ID
	MemberUuid     int64   `json:"memberUuid"     orm:"member_uuid"      description:"会员uuid"`      // 会员uuid
	CouponUuid     int64   `json:"couponUuid"     orm:"coupon_uuid"      description:"优惠券uuid"`     // 优惠券uuid
	UseOrderUuid   int64   `json:"useOrderUuid"   orm:"use_order_uuid"   description:"优惠券使用订单uuid"` // 优惠券使用订单uuid
	UseOrderAmount float64 `json:"useOrderAmount" orm:"use_order_amount" description:"优惠券使用订单金额"`   // 优惠券使用订单金额
	CreateTime     int     `json:"createTime"     orm:"create_time"      description:"创建时间"`        // 创建时间
	UpdateTime     int     `json:"updateTime"     orm:"update_time"      description:"更新时间"`        // 更新时间
	DeleteTime     int     `json:"deleteTime"     orm:"delete_time"      description:"删除时间"`        // 删除时间
}
