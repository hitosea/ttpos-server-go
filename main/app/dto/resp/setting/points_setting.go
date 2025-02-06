package setting

// Points 积分设置
type Points struct {
	DeductionOrder     string       `json:"deduction_order"`      // 扣款顺序 1-先主账户后赠送账户 2-先赠送账户后主账户 3-按比例
	DeductRatioMain    string       `json:"deduct_ratio_main"`    // 主账户扣款比例0-100
	DeductRatioGift    string       `json:"deduct_ratio_gift"`    // 赠送账户扣款比例0-100
	PointsName         string       `json:"points_name"`          // 积分名称自定义
	IsShoppingGift     string       `json:"is_shopping_gift"`     // 是否开启购物送积分
	GiftRatio          string       `json:"gift_ratio"`           // 积分赠送比例
	IsShoppingDiscount string       `json:"is_shopping_discount"` // 是否允许下单使用积分抵扣
	Discount           DiscountItem `json:"discount"`             // 积分抵扣
	Describe           string       `json:"describe"`             // 充值说明
	DeductOrder        string       `json:"deduct_order"`
}

type DiscountItem struct {
	DiscountRatio  string `json:"discount_ratio"`   // 积分抵扣比例
	FullOrderPrice string `json:"full_order_price"` // 订单满[?]元
	MaxMoneyRatio  string `json:"max_money_ratio"`  // 最高可抵扣订单额百分比
}
