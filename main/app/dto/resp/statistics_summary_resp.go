package resp

import "ttpos-server-go/app/dto"

// CompanySummaryItem 门店汇总统计门店项
type CompanySummaryItem struct {
	CompanyUuid uint64 `json:"company_uuid"` // 门店UUID
	CompanyName string `json:"company_name"` // 门店名称
	StoreCode   string `json:"store_code"`   // 店铺编号（用于排序和格式化）
}

// CompanySummaryListResp 门店列表响应
type CompanySummaryListResp struct {
	List []*CompanySummaryItem `json:"list"` // 门店列表
}

// CompanyBusinessSummaryResp 门店营业数据汇总响应
type CompanyBusinessSummaryResp struct {
	Meta       dto.PageResponse             `json:"meta"`        // 分页信息
	List       []CompanyBusinessSummaryItem `json:"list"`        // 明细列表或汇总列表
	SummaryRow CompanyBusinessSummaryItem   `json:"summary_row"` // 汇总行（明细表返回默认值，汇总表返回汇总数据）
}

// CompanyBusinessSummaryItem 门店营业数据汇总项
type CompanyBusinessSummaryItem struct {
	Date               string  `json:"date"`                  // 营业日（格式：YYYY-MM-DD）
	CompanyName        string  `json:"company_name"`          // 门店名称
	OrderAmount        float64 `json:"order_amount"`          // 订单金额（含优惠前，保留2位小数）
	PayAmount          float64 `json:"pay_amount"`            // 实付金额（保留2位小数）
	OrderNum           int64   `json:"order_num"`             // 订单数量
	CashTC             int64   `json:"cash_tc"`               // 现金TC（现金支付订单数）
	CashAmount         float64 `json:"cash_amount"`           // 现金金额（现金支付总金额，保留2位小数）
	CashAC             float64 `json:"cash_ac"`               // 现金AC（现金金额/现金TC，保留2位小数）
	MealNum            int64   `json:"meal_num"`              // 用餐人数
	DeskNum            int64   `json:"desk_num"`              // 消费桌数
	AvgCustomerPrice   float64 `json:"avg_customer_price"`    // 平均客单价（实付金额/用餐人数，保留2位小数）
	OrderAmountMealAvg float64 `json:"order_amount_meal_avg"` // 订单金额人均（订单金额/用餐人数，保留2位小数）
	OrderAmountAvg     float64 `json:"order_amount_avg"`      // 订单金额单均（订单金额/订单数量，保留2位小数）
	PayAmountAvg       float64 `json:"pay_amount_avg"`        // 实付金额单均（实付金额/订单数量，保留2位小数）
	InstantOrderAmount float64 `json:"instant_order_amount"`  // 点餐订单金额（保留2位小数）
	DeskOrderAmount    float64 `json:"desk_order_amount"`     // 桌台订单金额（保留2位小数）
	TakeoutOrderAmount float64 `json:"takeout_order_amount"`  // 外送订单金额（保留2位小数）
}

// CompanyPaymentMethodSummaryResp 门店支付方式汇总响应
type CompanyPaymentMethodSummaryResp struct {
	Meta       dto.PageResponse                  `json:"meta"`        // 分页信息
	List       []CompanyPaymentMethodSummaryItem `json:"list"`        // 明细列表或汇总列表
	SummaryRow []CompanyPaymentMethodSummaryItem `json:"summary_row"` // 汇总行（明细表返回空数组，汇总表返回按支付方式分组的汇总数据）
}

// CompanyPaymentMethodSummaryItem 门店支付方式汇总项
type CompanyPaymentMethodSummaryItem struct {
	Date          string  `json:"date"`           // 营业日（格式：YYYY-MM-DD 或 YYYY-MM）
	CompanyName   string  `json:"company_name"`   // 门店名称
	PaymentName   string  `json:"payment_name"`   // 支付方式名称
	PaymentAmount float64 `json:"payment_amount"` // 支付金额（保留2位小数）
	PaymentNum    int64   `json:"payment_num"`    // 支付笔数
	PaymentRatio  float64 `json:"payment_ratio"`  // 支付占比（保留2位小数，百分比）
}

// CompanyRefundSummaryResp 门店退款金额汇总响应
type CompanyRefundSummaryResp struct {
	Meta       dto.PageResponse           `json:"meta"`        // 分页信息
	List       []CompanyRefundSummaryItem `json:"list"`        // 明细列表或汇总列表
	SummaryRow CompanyRefundSummaryItem   `json:"summary_row"` // 汇总行（明细表返回默认值，汇总表返回汇总数据）
}

// CompanyRefundSummaryItem 门店退款金额汇总项
type CompanyRefundSummaryItem struct {
	Date                string  `json:"date"`                  // 营业日（格式：YYYY-MM-DD 或 YYYY-MM）
	CompanyName         string  `json:"company_name"`          // 门店名称
	RefundAmount        float64 `json:"refund_amount"`         // 退款金额（保留2位小数）
	RefundNum           int64   `json:"refund_num"`            // 退款笔数
	RefundRate          float64 `json:"refund_rate"`           // 退款率（保留2位小数，百分比）
	PartialRefundAmount float64 `json:"partial_refund_amount"` // 部分退款金额（保留2位小数）
	PartialRefundNum    int64   `json:"partial_refund_num"`    // 部分退款笔数
	FullRefundAmount    float64 `json:"full_refund_amount"`    // 整单退款金额（保留2位小数）
	FullRefundNum       int64   `json:"full_refund_num"`       // 整单退款笔数
}

// CompanyPaymentMethodListResp 门店支付方式列表响应
type CompanyPaymentMethodListResp struct {
	List []CompanyPaymentMethodItem `json:"list"` // 支付方式列表（已去重）
}

// CompanyPaymentMethodItem 门店支付方式项
type CompanyPaymentMethodItem struct {
	PaymentName string `json:"payment_name"` // 支付方式名称
}
