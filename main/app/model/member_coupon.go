package model

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/utils"
)

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

// 获取优惠券状态
func (model *MemberCoupon) GetStatus() int {
	if model.IsExpire() {
		return 2
	} else if model.IsUsed() {
		return 1
	}
	return 0
}

// 判断优惠券是否可用
// memberUuid 订单会员uuid
// nowTime 订单点击结账时该商家的时区实时时间，格式：HH:mm
func (model *MemberCoupon) IsAvailable(memberUuid uint64, nowTime string) error {
	if model.IsExpire() {
		return errors.New("优惠券已过期")
	}
	if model.IsUsed() {
		return errors.New("优惠券已使用")
	}
	if model.IsDisabled() {
		return errors.New("优惠券已禁用")
	}
	if model.MemberUuid != memberUuid {
		return errors.New("优惠券不属于该会员")
	}
	if model.IsDelete() {
		return errors.New("优惠券已删除")
	}
	// 不在使用时间区间内
	inRange, err := utils.IsTimeInRange(nowTime, model.MarketingCoupon.DayStartTime, model.MarketingCoupon.DayEndTime)
	if err != nil {
		return err
	}
	if !inRange {
		return errors.New("优惠券不在使用时间区间内")
	}

	return nil
}

// 判断优惠券是否过期
func (model *MemberCoupon) IsExpire() bool {
	nowTimestamp := time.Now().Unix()
	if model.StartTime <= nowTimestamp && nowTimestamp <= model.EndTime {
		return false
	}
	return true
}

// 判断优惠券是否已经使用
func (model *MemberCoupon) IsUsed() bool {
	return model.Status == 1 || model.UseTime != 0
}

// 判断优惠券是否已禁用
func (model *MemberCoupon) IsDisabled() bool {
	return model.MarketingCoupon.Status == 0 // 0禁用 1开启
}

// 会员优惠券使用记录 `ttpos_member_coupon_use_record`
type MemberCouponUseRecord struct {
	BaseModel
	MemberUuid     uint64  `gorm:"column:member_uuid;type:bigint(20);not null;comment:会员uuid" json:"member_uuid"`
	CouponUuid     uint64  `gorm:"column:coupon_uuid;type:bigint(20);not null;comment:优惠券uuid" json:"coupon_uuid"`
	UseOrderUuid   uint64  `gorm:"column:use_order_uuid;type:bigint(20);not null;comment:优惠券使用订单uuid" json:"use_order_uuid"`
	UseOrderAmount float64 `gorm:"column:use_order_amount;type:decimal(14,2);not null;comment:优惠券使用订单金额" json:"use_order_amount"`
}

func (model *MemberCouponUseRecord) SetNil() {

}
