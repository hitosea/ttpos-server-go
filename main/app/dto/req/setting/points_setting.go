package setting

type Points struct {
	DeductionOrder     string       `json:"deduction_order"`
	DeductRatioMain    string       `json:"deduct_ratio_main"`
	DeductRatioGift    string       `json:"deduct_ratio_gift"`
	PointsName         string       `json:"points_name"`
	IsShoppingGift     string       `json:"is_shopping_gift"`
	GiftRatio          string       `json:"gift_ratio"`
	IsShoppingDiscount string       `json:"is_shopping_discount"`
	Discount           DiscountItem `json:"discount"`
	Describe           string       `json:"describe"`
	DeductOrder        string       `json:"deduct_order"`
}

type DiscountItem struct {
	DiscountRatio  string `json:"discount_ratio"`
	FullOrderPrice string `json:"full_order_price"`
	MaxMoneyRatio  string `json:"max_money_ratio"`
}
