package model

type MemberCoupon struct {
	BaseModel
	MemberUuid uint64  `gorm:"column:member_uuid;type:bigint(20);not null;comment:会员uuid" json:"member_uuid"`
	CouponUuid uint64  `gorm:"column:coupon_uuid;type:bigint(20);not null;comment:优惠券uuid" json:"coupon_uuid"`
	Amount     float64 `gorm:"column:amount;type:decimal(14,2);not null;comment:优惠券面值" json:"amount"`
	Type       int     `gorm:"column:type;type:int(1);not null;comment:优惠券类型 0未知 1优惠券" json:"type"`
	Status     int     `gorm:"column:status;type:int(1);not null;comment:优惠券状态 0未使用 1已使用" json:"status"`
	StartTime  int64   `gorm:"column:start_time;type:bigint(20);not null;comment:优惠券开始时间" json:"start_time"`
	EndTime    int64   `gorm:"column:end_time;type:bigint(20);not null;comment:优惠券结束时间" json:"end_time"`
	UseTime    int64   `gorm:"column:use_time;type:bigint(20);not null;comment:优惠券使用时间" json:"use_time"`
}

type MemberCouponUseRecord struct {
	BaseModel
	MemberUuid     uint64  `gorm:"column:member_uuid;type:bigint(20);not null;comment:会员uuid" json:"member_uuid"`
	CouponUuid     uint64  `gorm:"column:coupon_uuid;type:bigint(20);not null;comment:优惠券uuid" json:"coupon_uuid"`
	UseOrderUuid   uint64  `gorm:"column:use_order_uuid;type:bigint(20);not null;comment:优惠券使用订单uuid" json:"use_order_uuid"`
	UseOrderAmount float64 `gorm:"column:use_order_amount;type:decimal(14,2);not null;comment:优惠券使用订单金额" json:"use_order_amount"`
}
