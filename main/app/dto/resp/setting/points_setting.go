package setting

import (
	"math"
	"strconv"

	"github.com/shopspring/decimal"
)

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

// 解析扣款顺序
func (p *Points) getDeductOrder() string {
	if p.DeductOrder == "1" {
		return "1"
	}
	if p.DeductOrder == "2" {
		return "2"
	}
	if p.DeductOrder == "3" {
		return "3"
	}
	// 默认扣款顺序为1
	return "1"
}

// 解析主账户扣款比例
func (p *Points) getDeductRatioMain() float64 {
	if p.DeductRatioMain == "" {
		return 0
	}
	ratio, err := strconv.ParseFloat(p.DeductRatioMain, 64)
	if err != nil {
		return 0
	}
	// 主账户扣款比例取值范围0-100, 转换为0-1
	ratio = decimal.NewFromFloat(ratio).Div(decimal.NewFromInt(100)).Truncate(2).InexactFloat64()
	// 主账户扣款比例取值范围0-1
	ratio = math.Min(ratio, 1)
	ratio = math.Max(ratio, 0)
	return ratio
}

// 解析赠送账户扣款比例
func (p *Points) getDeductRatioGift() float64 {
	if p.DeductRatioGift == "" {
		return 0
	}
	ratio, err := strconv.ParseFloat(p.DeductRatioGift, 64)
	if err != nil {
		return 0
	}
	// 赠送账户扣款比例取值范围0-100, 转换为0-1
	ratio = decimal.NewFromFloat(ratio).Div(decimal.NewFromInt(100)).Truncate(2).InexactFloat64()
	// 赠送账户扣款比例取值范围0-1
	ratio = math.Min(ratio, 1)
	ratio = math.Max(ratio, 0)
	return ratio
}

// getIsShoppingGift 解析是否开启购物送积分
func (p *Points) getIsShoppingGift() bool {
	if p.IsShoppingGift == "0" {
		return false
	}
	// 只有是1时，才开启购物送积分
	if p.IsShoppingGift == "1" {
		return true
	}
	return false
}

// getGiftRatio 解析积分赠送比例
func (p *Points) getGiftRatio() float64 {
	if p.GiftRatio == "" {
		return 0
	}
	ratio, err := strconv.ParseFloat(p.GiftRatio, 64)
	if err != nil {
		return 0
	}
	// 积分赠送比例取值范围0-100, 转换为0-1
	ratio = decimal.NewFromFloat(ratio).Div(decimal.NewFromInt(100)).InexactFloat64()
	// 积分赠送比例取值范围0-1
	ratio = math.Min(ratio, 1)
	ratio = math.Max(ratio, 0)
	return ratio
}

// GetDeductRatioMainAndGift 解析主账户扣款比例和赠送账户扣款比例
// 1-先主账户后赠送账户 2-先赠送账户后主账户 3-按比例
func (p *Points) GetDeductRatioMainAndGift() (float64, float64) {
	// 先主账户后赠送账户
	// 扣款顺序为1时，主账户扣款比例为1，赠送账户扣款比例为0
	if p.getDeductOrder() == "1" {
		return 1, 0
	}
	// 先赠送账户后主账户
	// 扣款顺序为2时，主账户扣款比例为0，赠送账户扣款比例为1
	if p.getDeductOrder() == "2" {
		return 0, 1
	}
	// 按比例
	// 扣款顺序为3时，主账户扣款比例和赠送账户扣款比例都为0
	if p.getDeductOrder() == "3" {
		return p.getDeductRatioMain(), p.getDeductRatioGift()
	}
	// 默认扣款顺序为1
	return 1, 0
}

// GetGiftRatio 积分赠送比例
func (p *Points) GetGiftRatio() float64 {
	// 如果没有开启购物送积分，则积分赠送比例为0
	if !p.getIsShoppingGift() {
		return 0
	}
	ratio := p.getGiftRatio()
	return ratio
}

type DiscountItem struct {
	DiscountRatio  string `json:"discount_ratio"`   // 积分抵扣比例
	FullOrderPrice string `json:"full_order_price"` // 订单满[?]元
	MaxMoneyRatio  string `json:"max_money_ratio"`  // 最高可抵扣订单额百分比
}
