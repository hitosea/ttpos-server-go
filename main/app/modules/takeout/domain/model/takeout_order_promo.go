package model

// TakeoutOrderPromo 外卖订单促销表（多平台）
type TakeoutOrderPromo struct {
	BaseModel

	// 关联外卖订单
	TakeoutOrderUuid string `gorm:"column:takeout_order_uuid" json:"takeout_order_uuid"`

	// 平台信息
	Platform string `gorm:"column:platform" json:"platform"`

	// 促销信息
	PromoCode        string `gorm:"column:promo_code" json:"promo_code"`               // 促销代码
	PromoName        string `gorm:"column:promo_name" json:"promo_name"`               // 促销名称
	PromoDescription string `gorm:"column:promo_description" json:"promo_description"` // 促销描述

	// 金额信息（单位：分）
	PromoAmount      int64 `gorm:"column:promo_amount" json:"promo_amount"`               // 促销金额
	MexFundedRatio   int   `gorm:"column:mex_funded_ratio" json:"mex_funded_ratio"`       // 商户承担比例（百分比）
	MexFundedAmount  int64 `gorm:"column:mex_funded_amount" json:"mex_funded_amount"`     // 商户承担金额
	TargetedPrice    int64 `gorm:"column:targeted_price" json:"targeted_price"`           // 目标价格（订单小计）
	PromoAmountInMin int64 `gorm:"column:promo_amount_in_min" json:"promo_amount_in_min"` // 促销金额（最小单位）
}

func (*TakeoutOrderPromo) TableName() string {
	return "ttpos_takeout_order_promo"
}
