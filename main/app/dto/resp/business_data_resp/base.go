package business_data_resp

// 商品
type Product struct {
	Name     string  `json:"name"`      // 商品名称
	SalesNum int     `json:"sales_num"` // 销售数量
	Price    float64 `json:"price"`     // 单价
	Subtotal float64 `json:"subtotal"`  // 小计
}

// 支付方式
type PaymentMethodIncome struct {
	Name     string  `json:"name"`      // 支付方式名称
	OrderNum int     `json:"order_num"` // 订单数量
	Amount   float64 `json:"amount"`    // 收入金额
	Code     int     `json:"code"`      // 支付方式代码
}

// 异常数据
type AbnormalData struct {
	RefundProductTimes    int `json:"refund_product_times"`     // 退菜次数
	RefundTimes           int `json:"refund_times"`             // 退款次数
	ReverseSettleTimes    int `json:"reverse_settle_times"`     // 反结账次数
	ProductFreeTimes      int `json:"product_free_times"`       // 赠菜次数
	FreeOrderTimes        int `json:"free_order_times"`         // 免单次数
	ProductMoveTimes      int `json:"product_move_times"`       // 转菜次数
	ChangePriceTimes      int `json:"change_price_times"`       // 单品改价次数
	ChangeOrderPriceTimes int `json:"change_order_price_times"` // 整单改价次数
	DiscountOrderTimes    int `json:"discount_order_times"`     // 整单折扣次数
	RoundOrderTimes       int `json:"round_order_times"`        // 整单抹零次数
}

// 会员数据
type MemberData struct {
	RechargeAmount float64 `json:"recharge_amount"` // 充值金额
	GiftMoney      float64 `json:"gift_money"`      // 赠送金额
	GiftPoints     int     `json:"gift_points"`     // 赠送积分
}

// 高峰时间
type PeakHour struct {
	TimePeriod string  `json:"time_period"` // 时间段
	OrderNum   int     `json:"num"`         // 订单数量
	Amount     float64 `json:"amount"`      // 订单金额
}

// 分类
type Category struct {
	Name     string  `json:"name"`      // 分类名称
	SalesNum int     `json:"sales_num"` // 销售数量
	Prices   float64 `json:"prices"`    // 销售金额
}

// 税收百分比对象
type Percentage struct {
	TaxRate        float64 `json:"tax_rate"`        // 税率
	ConsumptionTax float64 `json:"consumption_tax"` // 消费税
}

// 营业数据 - 全部
type BusinessDataAll struct {
	TotalSales             float64 `json:"total_sales"`               // 总销售额
	TotalReceivedPrice     float64 `json:"total_received_price"`      // 总实收金额
	TotalPayPrice          float64 `json:"total_pay_price"`           // 总支付金额
	TotalProductPrice      float64 `json:"total_product_price"`       // 总原商品金额 (未含税的总商品金额)
	TotalPayFeeMoney       float64 `json:"total_pay_fee_money"`       // 总支付手续费
	TotalServiceMoney      float64 `json:"total_service_money"`       // 总服务费
	TotalTaxMoney          float64 `json:"total_tax_money"`           // 总税费
	TotalUserDiscountMoney float64 `json:"total_user_discount_money"` // 总会员折扣
	TotalDiscountMoney     float64 `json:"total_discount_money"`      // 总优惠折扣
	TotalFreeOrderPrice    float64 `json:"total_free_order_price"`    // 总免单金额
	TotalRefundMoney       float64 `json:"total_refund_money"`        // 总退款金额
	TotalOrderNum          int     `json:"total_order_num"`           // 总订单数
	TotalPeopleNum         int     `json:"total_people_num"`          // 总人数
	TotalProductNum        int     `json:"total_product_num"`         // 总商品数
	TotalTableNum          int     `json:"total_table_num"`           // 总桌数
	AvgOrderPrice          float64 `json:"avg_order_price"`           // 平均订单金额
	MinOrderPrice          float64 `json:"min_order_price"`           // 最小订单金额
	MaxOrderPrice          float64 `json:"max_order_price"`           // 最大订单金额
	// 桌台方式
	AllTableOrderNum      int     `json:"all_table_order_num"`       // 总桌数订单数
	AllTablePeopleNum     int     `json:"all_table_people_num"`      // 总桌数人数
	AllTableAvgOrderPrice float64 `json:"all_table_avg_order_price"` // 总桌数平均订单金额
	AllTableMinOrderPrice float64 `json:"all_table_min_order_price"` // 总桌数最小订单金额
	AllTableMaxOrderPrice float64 `json:"all_table_max_order_price"` // 总桌数最大订单金额
	AllTablePeopleAvg     float64 `json:"all_table_people_avg"`      // 总桌数人均
	// 收银方式
	CashierOrderNum      int     `json:"cashier_order_num"`       // 收银方式订单数
	CashierMinOrderPrice float64 `json:"cashier_min_order_price"` // 收银方式最小订单金额
	CashierMaxOrderPrice float64 `json:"cashier_max_order_price"` // 收银方式最大订单金额
	CashierAvgOrderPrice float64 `json:"cashier_avg_order_price"` // 收银方式平均订单金额
	// 未结账数据
	UnclosedTotalOrderNum int     `json:"unclosed_total_order_num"` // 总订单数
	UnclosedTotalPrice    float64 `json:"unclosed_total_price"`     // 总金额
	// 支付方式
	PaymentMethodIncomes []PaymentMethodIncome `json:"payment_method_incomes"` // 支付方式
	AbnormalData         AbnormalData          `json:"abnormal_data"`          // 异常数据
	MemberData           MemberData            `json:"member_data"`            // 会员数据
	PeakHourList         []PeakHour            `json:"peak_hour_list"`         // 高峰时间
	CategoryList         []Category            `json:"category_list"`          // 分类列表
	PercentageList       []Percentage          `json:"percentage_list"`        // 税收百分比对象列表
}

// 营业数据 - 按支付方式
type BusinessDataPaymentMethod struct {
	TotalReceivedPrice   float64               `json:"total_received_price"`   // 总实收金额
	PaymentMethodIncomes []PaymentMethodIncome `json:"payment_method_incomes"` // 支付方式
}

// 营业数据 - 按商品分类
type BusinessDataProductCategory struct {
	SalesNum             int                   `json:"sales_num"`              // 销售笔数
	RefundAmount         float64               `json:"refund_amount"`          // 退款金额
	TotalReceivedPrice   float64               `json:"total_received_price"`   // 总实收金额
	CategoryList         []Category            `json:"category_list"`          // 分类列表
	PaymentMethodIncomes []PaymentMethodIncome `json:"payment_method_incomes"` // 支付方式
}

// 营业数据 - 按商品
type BusinessDataProduct struct {
	Products []Product `json:"products"` // 商品列表
}
