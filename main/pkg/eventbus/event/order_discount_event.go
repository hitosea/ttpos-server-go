package event

import "ttpos-server-go/pkg/utils"

// DiscountSaleOrderPayload 优惠折扣时间荷载
type DiscountSaleOrderPayload struct {
	BasePayload
	OldPrice        float64             `json:"old_price"`         // 折扣优惠前的总金额
	NewPrice        float64             `json:"new_price"`         // 折扣优惠后的总金额
	DiscountType    int                 `json:"discount_type"`     // 折扣类型。1: 订单改价 2: 折扣 3:抹零
	SpecialDiscount float64             `json:"special_discount"` // 优惠金额。整单打折后的优惠金额=会员折扣后的订单应收金额-订单应收金额
	RoundingRate    float64             `json:"rounding_rate"`     // 整单打折使用，打折率。如八折，则打折率是20； 如30%off，则打折率是30。统一展示格式为"优惠折扣：折扣-80%（￥50）"，无论是百分比打折还是百分比减免，都统一展示为百分比减免。
	RoundingType    int                 `json:"rounding_type"`     // 订单抹零使用，抹零规则 1:抹分 2:抹角 3:四舍五入保留一位小数 4:四舍五入到整数
	IsAuto          bool                `json:"is_auto"`           // 是否自动抹零
	AuthorizedStaff *AuthorizedStaffInfo `json:"authorized_staff"`  // 授权员工信息（如果使用了授权验证）
}

func (payload *DiscountSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}
