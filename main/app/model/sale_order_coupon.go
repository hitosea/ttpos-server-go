package model

// SaleOrderCoupon 销售订单优惠券 `ttpos_sale_order_coupon`
type SaleOrderCoupon struct {
	BaseModel
	CouponAmount        float64 `gorm:"column:coupon_amount;type:decimal(12,2);default:0;comment:优惠券抵扣金额，实际抵扣金额" json:"coupon_amount"`
	CouponOriginAmount  float64 `gorm:"column:coupon_origin_amount;type:decimal(12,2);default:0;comment:优惠券原始金额(面值)" json:"coupon_origin_amount"`
	CouponRequirement   string  `gorm:"column:coupon_requirement;type:varchar(255);default:'';comment:优惠券的类型，none-所有人可用，但一个saleBill只能用一张优惠券;marketing-会员通过营销活动获的优惠券；" json:"coupon_requirement"`
	MemberCouponUuid    uint64  `gorm:"column:member_coupon_uuid;type:bigint(20);default:0;comment:会员的优惠券uuid,表示该订单使用会员的哪个优惠券。none时有值" json:"member_coupon_uuid"`
	MarketingCouponUuid uint64  `gorm:"column:marketing_coupon_uuid;type:bigint(20);default:0;comment:营销优惠券uuid,表示该订单使用营销的哪个优惠券。marketing时有值" json:"marketing_coupon_uuid"`
	SaleOrderUuid       uint64  `gorm:"column:sale_order_uuid;type:bigint(20);default:0;comment:销售订单ID" json:"sale_order_uuid"`
}
