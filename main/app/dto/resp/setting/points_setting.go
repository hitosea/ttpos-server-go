package setting

import (
	"math"
	"strconv"

	"github.com/duke-git/lancet/v2/slice"

	"github.com/shopspring/decimal"
)

// Points 积分设置
type Points struct {
	DeductionOrder     string             `json:"deduction_order"`      // 扣款顺序 1-先主账户后赠送账户 2-先赠送账户后主账户 3-按比例
	DeductRatioMain    string             `json:"deduct_ratio_main"`    // 主账户扣款比例0-100
	DeductRatioGift    string             `json:"deduct_ratio_gift"`    // 赠送账户扣款比例0-100
	PointsName         string             `json:"points_name"`          // 积分名称自定义
	IsShoppingGift     string             `json:"is_shopping_gift"`     // 是否开启购物送积分
	GiftRatio          string             `json:"gift_ratio"`           // 积分赠送比例
	IsShoppingDiscount string             `json:"is_shopping_discount"` // 是否允许下单使用积分抵扣
	Discount           DiscountItem       `json:"discount"`             // 积分抵扣
	Describe           string             `json:"describe"`             // 充值说明
	DeductOrder        string             `json:"deduct_order"`
	OpenPointsExchange string             `json:"open_points_exchange"` // 是否开启积分抵扣，“1”开启，“0”关闭
	PointsExchangeRate string             `json:"points_exchange_rate"` // 积分汇率，每积分抵扣的金额，可输入大于0的两位小数
	AutoPointsExchange string             `json:"auto_points_exchange"` // 是否开启自动抵扣，“1”开启，“0”关闭
	ShoppingGiftRules  []ShoppingGiftRule `json:"shopping_gift_rules"`  // 购物赠送积分规则
	Exchange           PointsExchange     `json:"exchange"`             // 积分抵扣消费
}

const RuleTypePaymentAmount = "payment_amount" // 按付款金额比例赠送
const RuleTypeDesk = "desk"                    // 按桌台人数赠送

type ShoppingGiftRule struct {
	Type                     string            `json:"type"`                       // 类型: payment_amount - 按付款金额比例赠送; desk - 按桌台人数赠送
	IsOpen                   string            `json:"is_open"`                    // 是否开启: "1" - 开启; "0" - 关闭
	IsMemberLevelRelated     string            `json:"is_member_level_related"`    // 是否按会员等级赠送: "1" - 是; "0" - 否
	Value                    string            `json:"value"`                      // 积分比例、积分数量
	PaymentAmountRequirement string            `json:"payment_amount_requirement"` // 付款金额要求。如“12.56”
	MealType                 []string          `json:"meal_type"`                  // 就餐类型: buffet - 自助餐; non-buffet - 非自助餐
	BalancePaymentGetPoints  string            `json:"balance_payment_get_points"` // 会员余额支付是否赠送: "1" - 是; "0" - 否
	RefundReturnPoints       string            `json:"refund_return_points"`       // 退款自动扣积分: "1" - 是; "0" - 否
	MemberLevels             []MemberLevelItem `json:"member_levels"`              // 会员等级
}

// 是否按会员等级赠送; true: 按等级区分； false：所有会员等级相同
func (s *ShoppingGiftRule) getIsMemberLevelRelated() bool {
	if s.IsMemberLevelRelated == "1" {
		return true
	}
	return false
}

func (s *ShoppingGiftRule) getValue() float64 {
	if s.Value == "" {
		return 0
	}
	value, _ := strconv.ParseFloat(s.Value, 64)
	return value
}

func (s *ShoppingGiftRule) getPaymentAmountRequirement() float64 {
	if s.PaymentAmountRequirement == "" {
		return 0
	}
	value, _ := strconv.ParseFloat(s.Value, 64)
	return value
}

// getBalancePaymentGetPoints 获取会员余额支付是否赠送, 默认是赠送
func (s *ShoppingGiftRule) getBalancePaymentGetPoints() bool {
	if s.BalancePaymentGetPoints == "0" {
		return false
	}
	return true
}

// getType 获取规则类型. 0 - 按付款金额比例赠送; 1 - 按桌台人数赠送:
func (s *ShoppingGiftRule) getType() uint8 {
	if s.Type == RuleTypePaymentAmount {
		return 0
	} else if s.Type == RuleTypeDesk {
		return 1
	}
	return 0
}

type PointsRule struct {
	GiftPointsType           uint8   // 0 按付款金额比例赠送；1 按桌台人数赠送
	Value                    float64 // 积分比例、积分数量
	BalancePaymentGetPoints  bool    // 会员余额支付是否赠送
	PaymentAmountRequirement float64 // 付款金额达到多少时，开始赠送积分。默认是0，无论多少都发放积分
}

type MemberLevelItem struct {
	Uuid  uint64 `json:"uuid"`  // 会员等级Uuid
	Name  string `json:"name"`  // 会员等级名称
	Value string `json:"value"` // 积分比例/积分数量，在后台修改时，根据type去设置对应会员等级的points_rate或者points_quantity
}

func (item *MemberLevelItem) getValue() float64 {
	if item.Value == "" {
		return 0
	}
	value, _ := strconv.ParseFloat(item.Value, 64)
	return value
}

type PointsExchange struct {
	OpenPointsExchange string `json:"open_points_exchange"` // 会员积分抵扣订单金额
	PointsExchangeRate string `json:"points_exchange_rate"` // 每积分抵扣应付金额
	AutoPointsExchange string `json:"auto_points_exchange"` // 是否自动抵扣
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

// getOpenPointsExchange 解析是否开启积分抵扣
func (p *Points) GetOpenPointsExchange() bool {
	if p.OpenPointsExchange == "1" {
		return true
	}
	return false
}

// 获取每积分抵扣金额。当未开启积分抵扣时，积分抵扣金额为0
func (p *Points) GetPointsExchangeRate() float64 {
	// 如果未开启积分抵扣时，积分抵扣金额为0
	if !p.GetOpenPointsExchange() {
		return 0
	}
	rate := p.getPointsExchangeRate()
	return rate
}

// getPointsExchangeRate 解析积分汇率, 每积分抵扣的金额，可输入大于0的两位小数
func (p *Points) getPointsExchangeRate() float64 {
	if p.PointsExchangeRate == "" {
		return 0
	}
	rate, err := strconv.ParseFloat(p.PointsExchangeRate, 64)
	if err != nil {
		return 0
	}
	return rate
}

// 是否开启自动抵扣
func (p *Points) IsAutoPointsExchange() bool {
	if p.AutoPointsExchange == "1" {
		return true
	}
	return false
}

// 获取订单积分赠送规则
func (p *Points) GetPointsGiftRule(isBufferOrder bool, memberLevelUuid uint64) PointsRule {
	var targetRule *ShoppingGiftRule
	var buffetRule *ShoppingGiftRule    // 自助餐规格的积分发放规格
	var nonBuffetRule *ShoppingGiftRule // 非自助餐的积分发放规格
	for _, rule := range p.ShoppingGiftRules {
		if slice.Contain(rule.MealType, "buffet") {
			buffetRule = &rule
		}
		if slice.Contain(rule.MealType, "non-buffet") {
			nonBuffetRule = &rule
		}
	}
	if isBufferOrder {
		targetRule = buffetRule
	} else {
		targetRule = nonBuffetRule
	}

	if targetRule == nil {
		return PointsRule{
			Value: 0, // 没有找到积分赠送规格时，0表示不赠送积分
		}
	}

	value := targetRule.getValue() // 默认使用“所有会员等级相同”的值
	// 如果按等级区分
	if targetRule.getIsMemberLevelRelated() {
		match := false // 判断是否找到对应的等级
		for _, level := range targetRule.MemberLevels {
			if level.Uuid == memberLevelUuid {
				value = level.getValue() // 使用对应等级的值
				match = true
				break
			}
		}
		// 如果没有找到对应的等级,则不赠送积分
		if !match {
			value = 0
		}
	}

	return PointsRule{
		GiftPointsType:           targetRule.getType(),                     // 按比例 或 按人数
		Value:                    value,                                    // 比例 或 每人积分数
		BalancePaymentGetPoints:  targetRule.getBalancePaymentGetPoints(),  // 会员余额支付是否赠送积分
		PaymentAmountRequirement: targetRule.getPaymentAmountRequirement(), // 付款金额达到多少时，开始赠送积分。
	}
}

type DiscountItem struct {
	DiscountRatio  string `json:"discount_ratio"`   // 积分抵扣比例
	FullOrderPrice string `json:"full_order_price"` // 订单满[?]元
	MaxMoneyRatio  string `json:"max_money_ratio"`  // 最高可抵扣订单额百分比
}
