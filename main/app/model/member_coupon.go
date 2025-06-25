package model

// 会员优惠券 `ttpos_member_coupon`
type MemberCoupon struct {
	BaseModel
	MemberUuid     uint64  `gorm:"column:member_uuid;type:bigint(20);not null;comment:会员uuid" json:"member_uuid"`
	CouponUuid     uint64  `gorm:"column:coupon_uuid;type:bigint(20);not null;comment:优惠券uuid" json:"coupon_uuid"`
	Name           string  `gorm:"column:name;type:varchar(50);not null;comment:优惠券名称" json:"name"`
	DeductionType  string  `gorm:"column:deduction_type;type:varchar(20);not null;comment:抵扣类型: taxed - 税后抵扣" json:"deduction_type"`
	Type           string  `gorm:"column:type;type:varchar(20);not null;comment:优惠券类型: deduction - 抵扣券" json:"type"`
	DayStartTime   string  `gorm:"column:day_start_time;type:varchar(5);not null;comment:每日适用时段开始时间, hh:mm 格式" json:"day_start_time"`
	DayEndTime     string  `gorm:"column:day_end_time;type:varchar(5);not null;comment:每日适用时段结束时间, hh:mm 格式" json:"day_end_time"`
	ValidStartTime int     `gorm:"column:valid_start_time;type:int(11);default:0;comment:优惠券有效开始时间, requirement = none 时有效" json:"valid_start_time"`
	ValidEndTime   int     `gorm:"column:valid_end_time;type:int(11);default:0;comment:优惠券有效结束时间, requirement = none 时有效" json:"valid_end_time"`
	Amount         float64 `gorm:"column:amount;type:decimal(14,2);not null;comment:优惠券面值" json:"amount"`
	Status         int     `gorm:"column:status;type:int(1);not null;comment:优惠券状态 0未使用 1已使用" json:"status"`
	StartTime      int64   `gorm:"column:start_time;type:bigint(20);not null;comment:优惠券开始时间" json:"start_time"`
	EndTime        int64   `gorm:"column:end_time;type:bigint(20);not null;comment:优惠券结束时间" json:"end_time"`
	UseTime        int64   `gorm:"column:use_time;type:bigint(20);not null;comment:优惠券使用时间" json:"use_time"`

	MarketingCoupon *MarketingCoupon `gorm:"foreignKey:CouponUuid;references:Uuid" json:"marketing_coupon"`
}

// 会员优惠券使用记录 `ttpos_member_coupon_use_record`
type MemberCouponUseRecord struct {
	BaseModel
	MemberUuid     uint64  `gorm:"column:member_uuid;type:bigint(20);not null;comment:会员uuid" json:"member_uuid"`
	CouponUuid     uint64  `gorm:"column:coupon_uuid;type:bigint(20);not null;comment:优惠券uuid" json:"coupon_uuid"`
	UseOrderUuid   uint64  `gorm:"column:use_order_uuid;type:bigint(20);not null;comment:优惠券使用订单uuid" json:"use_order_uuid"`
	UseOrderAmount float64 `gorm:"column:use_order_amount;type:decimal(14,2);not null;comment:优惠券使用订单金额" json:"use_order_amount"`
}
