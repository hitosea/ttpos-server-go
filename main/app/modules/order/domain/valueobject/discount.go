package valueobject

import "ttpos-server-go/app/errors"

// DiscountType 折扣类型
type DiscountType int

const (
	// DiscountTypePercent 百分比折扣
	DiscountTypePercent DiscountType = 1
	// DiscountTypeFixed 固定金额折扣
	DiscountTypeFixed DiscountType = 2
)

// Discount 订单优惠值对象（不可变）
type Discount struct {
	discountType DiscountType // 折扣类型
	value        float64      // 折扣值（百分比或固定金额）
	reason       string       // 优惠原因
}

// NewPercentDiscount 创建百分比折扣
func NewPercentDiscount(percent float64, reason string) (*Discount, error) {
	if percent < 0 || percent > 100 {
		return nil, errors.New("折扣百分比必须在0-100之间")
	}
	return &Discount{
		discountType: DiscountTypePercent,
		value:        percent,
		reason:       reason,
	}, nil
}

// NewFixedDiscount 创建固定金额折扣
func NewFixedDiscount(amount float64, reason string) (*Discount, error) {
	if amount < 0 {
		return nil, errors.New("折扣金额不能为负数")
	}
	return &Discount{
		discountType: DiscountTypeFixed,
		value:        amount,
		reason:       reason,
	}, nil
}

// Type 获取折扣类型
func (d *Discount) Type() DiscountType {
	return d.discountType
}

// Value 获取折扣值
func (d *Discount) Value() float64 {
	return d.value
}

// Reason 获取优惠原因
func (d *Discount) Reason() string {
	return d.reason
}

// IsPercent 是否百分比折扣
func (d *Discount) IsPercent() bool {
	return d.discountType == DiscountTypePercent
}

// IsFixed 是否固定金额折扣
func (d *Discount) IsFixed() bool {
	return d.discountType == DiscountTypeFixed
}

// Calculate 计算折扣金额
func (d *Discount) Calculate(originalAmount float64) float64 {
	if d.discountType == DiscountTypePercent {
		return originalAmount * d.value / 100
	}
	// 固定金额折扣不能超过原价
	if d.value > originalAmount {
		return originalAmount
	}
	return d.value
}
