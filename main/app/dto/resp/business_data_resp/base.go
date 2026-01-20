package business_data_resp

import "ttpos-server-go/app/dto"

// 商品
type Product struct {
	Name     string  `json:"name"`      // 商品名称
	SalesNum float64 `json:"sales_num"` // 销售数量
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
	FreeOrderTimes        int `json:"free_order_times"`         // 免单次数
	ProductFreeTimes      int `json:"product_free_times"`       // 赠菜次数
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
	GiftPoints     float64 `json:"gift_points"`     // 赠送积分
	UserCount      int     `json:"user_count"`      // 会员数量
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
	SalesNum float64 `json:"sales_num"` // 销售数量
	Prices   float64 `json:"prices"`    // 销售金额
}

// 税收百分比对象
type Percentage struct {
	TaxRate        float64 `json:"tax_rate"`        // 税率
	ConsumptionTax float64 `json:"consumption_tax"` // 消费税
	TotalPrice     float64 `json:"total_price"`     // 合计
}

// 营业数据 - 全部
type BusinessDataAll struct {
	TotalSales                 float64 `json:"total_sales"`                   // 总销售额
	TotalReceivedPrice         float64 `json:"total_received_price"`          // 总实收金额
	TotalPayPrice              float64 `json:"total_pay_price"`               // 总支付金额
	TotalProductPrice          float64 `json:"total_product_price"`           // 总原商品金额 (未含税的总商品金额)
	TotalPayFeeMoney           float64 `json:"total_pay_fee_money"`           // 总支付手续费
	TotalServiceMoney          float64 `json:"total_service_money"`           // 总服务费
	TotalTaxMoney              float64 `json:"total_tax_money"`               // 总税费
	TotalUserDiscountMoney     float64 `json:"total_user_discount_money"`     // 总会员折扣
	TotalDiscountMoney         float64 `json:"total_discount_money"`          // 总优惠折扣
	TotalDiscountRatio         float64 `json:"total_discount_ratio"`          // 总优惠占比
	TotalFreeOrderPrice        float64 `json:"total_free_order_price"`        // 总免单金额
	TotalRefundMoney           float64 `json:"total_refund_money"`            // 总退款金额
	TotalGiveProductPrice      float64 `json:"total_give_product_price"`      // 总赠菜金额
	TotalFreeOrderNum          int     `json:"total_free_order_num"`          // 总免单数量
	TotalGiveProductNum        float64 `json:"total_give_product_num"`        // 总赠菜数量
	TotalOrderNum              int     `json:"total_order_num"`               // 总订单数
	TotalPeopleNum             int     `json:"total_people_num"`              // 总人数
	TotalProductNum            float64 `json:"total_product_num"`             // 总商品数
	TotalTakeoutSaleAmount     float64 `json:"total_takeout_sale_amount"`     // 总外送销售
	TotalTakeoutBusinessAmount float64 `json:"total_takeout_business_amount"` // 总外送营收
	TotalTakeoutRefundAmount   float64 `json:"total_takeout_refund_amount"`   // 总外送退款金额
	TotalTakeoutDeliveryFee    float64 `json:"total_takeout_delivery_fee"`    // 总外送配送费
	TotalTableNum              int     `json:"total_table_num"`               // 总桌数
	TotalCancelOrderNum        int     `json:"total_cancel_order_num"`        // 总取消订单数
	TotalCancelOrderAmount     float64 `json:"total_cancel_order_amount"`     // 总取消订单金额
	AvgOrderPrice              float64 `json:"avg_order_price"`               // 平均订单金额
	MinOrderPrice              float64 `json:"min_order_price"`               // 最小订单金额
	MaxOrderPrice              float64 `json:"max_order_price"`               // 最大订单金额
	// 桌台方式
	AllTableOrderNum      int     `json:"all_table_order_num"`       // 总桌数订单数
	AllTablePeopleNum     int     `json:"all_table_people_num"`      // 总桌数人数
	AllTableAvgOrderPrice float64 `json:"all_table_avg_order_price"` // 总桌数平均订单金额
	AllTableMinOrderPrice float64 `json:"all_table_min_order_price"` // 总桌数最小订单金额
	AllTableMaxOrderPrice float64 `json:"all_table_max_order_price"` // 总桌数最大订单金额
	AllTablePeopleAvg     float64 `json:"all_table_people_avg"`      // 总桌数人均
	// 收银方式 - 店内 (不包含外送和外卖)
	AllCashierOrderNum      int     `json:"all_cashier_order_num"`       // 收银方式订单数
	AllCashierMinOrderPrice float64 `json:"all_cashier_min_order_price"` // 收银方式最小订单金额
	AllCashierMaxOrderPrice float64 `json:"all_cashier_max_order_price"` // 收银方式最大订单金额
	AllCashierAvgOrderPrice float64 `json:"all_cashier_avg_order_price"` // 收银方式平均订单金额
	// 外送数据
	AllTakeoutOrderNum      int     `json:"all_takeout_order_num"`       // 总外送订单数
	AllTakeoutMinOrderPrice float64 `json:"all_takeout_min_order_price"` // 总外送最小订单金额
	AllTakeoutMaxOrderPrice float64 `json:"all_takeout_max_order_price"` // 总外送最大订单金额
	AllTakeoutAvgOrderPrice float64 `json:"all_takeout_avg_order_price"` // 总外送平均订单金额
	// 外卖数据
	AllTakeawayOrderNum      int     `json:"all_takeaway_order_num"`       // 点餐方式-外卖订单数
	AllTakeawayMinOrderPrice float64 `json:"all_takeaway_min_order_price"` // 点餐方式-外卖最小订单金额
	AllTakeawayMaxOrderPrice float64 `json:"all_takeaway_max_order_price"` // 点餐方式-外卖最大订单金额
	AllTakeawayAvgOrderPrice float64 `json:"all_takeaway_avg_order_price"` // 点餐方式-外卖平均订单金额
	// 未结账数据
	UnclosedTotalOrderNum int     `json:"unclosed_total_order_num"` // 未结账数据 - 总订单数
	UnclosedTotalPrice    float64 `json:"unclosed_total_price"`     // 未结账数据 - 总金额
	// 支付方式
	PaymentMethodIncomes []PaymentMethodIncome `json:"payment_method_incomes"` // 支付方式
	AbnormalData         AbnormalData          `json:"abnormal_data"`          // 异常数据
	MemberData           MemberData            `json:"member_data"`            // 会员数据
	PeakHourList         []PeakHour            `json:"peak_hour_list"`         // 高峰时间
	CategoryList         []Category            `json:"category_list"`          // 分类列表
	PercentageList       []Percentage          `json:"percentage_list"`        // 税收百分比对象列表
	OpeningHours         string                `json:"opening_hours"`          // 营业时间
}

// 营业数据 - 按支付方式
type BusinessDataPaymentMethod struct {
	TotalReceivedPrice   float64               `json:"total_received_price"`   // 总实收金额
	PaymentMethodIncomes []PaymentMethodIncome `json:"payment_method_incomes"` // 支付方式
	OpeningHours         string                `json:"opening_hours"`          // 营业时间
}

// 营业数据 - 按商品分类
type BusinessDataProductCategory struct {
	SalesNum             int                   `json:"sales_num"`              // 销售笔数
	TotalRefundMoney     float64               `json:"total_refund_money"`     // 总退款金额
	TotalReceivedPrice   float64               `json:"total_received_price"`   // 总实收金额
	CategoryList         []Category            `json:"category_list"`          // 分类列表
	PaymentMethodIncomes []PaymentMethodIncome `json:"payment_method_incomes"` // 支付方式
	OpeningHours         string                `json:"opening_hours"`          // 营业时间
}

// 营业数据 - 按商品
type BusinessDataProduct struct {
	Products     []Product `json:"products"`      // 商品列表
	OpeningHours string    `json:"opening_hours"` // 营业时间
	BatchRange   string    `json:"batch_range"`   // 批次号
}

// 营业数据 - 按区域
type BusinessDataArea struct {
	Areas []Area `json:"areas"` // 区域列表
}

// 区域
type Area struct {
	Name               string  `json:"name"`                 // 区域名称
	TotalSales         float64 `json:"total_sales"`          // 总销售额
	TotalReceivedPrice float64 `json:"total_received_price"` // 总实收金额
	TotalProductNum    float64 `json:"total_product_num"`    // 总商品数
}

// 营业数据 - 商品排行
type BusinessDataProductRank struct {
	Ranks []ProductRank `json:"ranks"` // 商品排行
}

// 商品排行
type ProductRank struct {
	ProductName string  `json:"product_name"` // 商品名称
	SalesNum    float64 `json:"sales_num"`    // 销售数量
	SalesPrice  float64 `json:"sales_price"`  // 销售金额
}

// 营业数据 - 商品销售列表统计
type BusinessDataCountProductSalesPagination struct {
	List []BusinessDataCountProductSalesItem `json:"list"` // 商品列表
	Meta dto.PageResponse                    `json:"meta"` // 元数据
}

// BusinessDataCountProductItem 营业数据商品销售统计列表
type BusinessDataCountProductSalesItem struct {
	ProductName        string  `json:"product_name"`         // 商品名称
	CategoryName       string  `json:"category_name"`        // 分类名称
	OriginalSalesPrice float64 `json:"original_sales_price"` // 原价销售额
	SalesPrice         float64 `json:"sales_price"`          // 实际销售金额
	TotalPayPrice      float64 `json:"total_pay_price"`      // 营业收入
	SalesNum           float64 `json:"sales_num"`            // 销售数量
	GiveProductNum     float64 `json:"give_product_num"`     // 赠菜数量
}

// 营业数据 - 7天
type BusinessDataCount7Days struct {
	Days []string                     `json:"days"` // 7天
	Data []BusinessDataCount7DaysItem `json:"data"` // 7天
}

// 营业数据 - 7天 - 单天
type BusinessDataCount7DaysItem struct {
	Day        string  `json:"day"`         // 日期
	TotalNum   int64   `json:"total_num"`   // 总订单数量
	TotalMoney float64 `json:"total_money"` // 总实收金额
}

// 营业数据 - 导出
type BusinessDataExport struct {
	Days []string                 `json:"days"` // 日期
	Data []BusinessDataExportItem `json:"data"` // 数据
}

// 营业数据 - 导出 - 单天
type BusinessDataExportItem struct {
	Day                        string                      `json:"day"`                           // 日期
	TotalSaleAmount            float64                     `json:"total_sale_amount"`             // 总销售额
	TotalReceivedAmount        float64                     `json:"total_received_amount"`         // 实收金额
	TotalProductNum            float64                     `json:"total_product_num"`             // 商品数量
	TotalMemberNum             int64                       `json:"total_member_num"`              // 新增会员数
	TotalDiscountMember        float64                     `json:"total_discount_member"`         // 会员折扣
	TotalBusinessAmount        float64                     `json:"total_business_amount"`         // 营业收入
	TotalServiceFee            float64                     `json:"total_service_fee"`             // 服务费
	TotalPaymentFee            float64                     `json:"total_payment_fee"`             // 支付手续费
	TotalTax                   float64                     `json:"total_tax"`                     // 税费
	TotalRefundAmount          float64                     `json:"total_refund_amount"`           // 退款金额
	TotalDiscount              float64                     `json:"total_discount"`                // 优惠折扣
	TotalDiscountRatio         float64                     `json:"total_discount_ratio"`          // 优惠折扣占比
	TotalGiftAmount            float64                     `json:"total_gift_amount"`             // 赠菜总额
	TotalGiftNum               float64                     `json:"total_gift_num"`                // 赠菜数量
	TotalFreeAmount            float64                     `json:"total_free_amount"`             // 免单总额
	TotalFreeNum               float64                     `json:"total_free_num"`                // 免单数量
	TotalRechargeAmount        float64                     `json:"total_recharge_amount"`         // 充值金额
	TotalTakeoutSaleAmount     float64                     `json:"total_takeout_sale_amount"`     // 外送销售
	TotalTakeoutBusinessAmount float64                     `json:"total_takeout_business_amount"` // 外送营收
	TotalTakeoutRefundAmount   float64                     `json:"total_takeout_refund_amount"`   // 外送退款
	TotalTakeoutDeliveryFee    float64                     `json:"total_takeout_delivery_fee"`    // 外送配送费
	TotalOrderNum              int64                       `json:"total_order_num"`               // 合计订单数
	MinOrderAmount             float64                     `json:"min_order_amount"`              // 最小订单金额
	MaxOrderAmount             float64                     `json:"max_order_amount"`              // 最大订单金额
	AvgOrderAmount             float64                     `json:"avg_order_amount"`              // 平均订单金额
	TotalDeskNum               int64                       `json:"total_desk_num"`                // 总桌台数
	TotalMealNum               int64                       `json:"total_meal_num"`                // 桌台方式-人数
	MinDeskOrderAmount         float64                     `json:"min_desk_order_amount"`         // 桌台方式-最小订单金额
	MaxDeskOrderAmount         float64                     `json:"max_desk_order_amount"`         // 桌台方式-最大订单金额
	AvgDeskOrderAmount         float64                     `json:"avg_desk_order_amount"`         // 桌台方式-平均订单金额
	TotalInstantOrderNum       int64                       `json:"total_instant_order_num"`       // 点餐方式-订单数
	MinInstantOrderAmount      float64                     `json:"min_instant_order_amount"`      // 点餐方式-最小订单金额
	MaxInstantOrderAmount      float64                     `json:"max_instant_order_amount"`      // 点餐方式-最大订单金额
	AvgInstantOrderAmount      float64                     `json:"avg_instant_order_amount"`      // 点餐方式-平均订单金额
	TotalTakeoutOrderNum       int64                       `json:"total_takeout_order_num"`       // 外送点餐-订单数
	MinTakeoutOrderAmount      float64                     `json:"min_takeout_order_amount"`      // 外送点餐-最小订单金额
	MaxTakeoutOrderAmount      float64                     `json:"max_takeout_order_amount"`      // 外送点餐-最大订单金额
	AvgTakeoutOrderAmount      float64                     `json:"avg_takeout_order_amount"`      // 外送点餐-平均订单金额
	TotalGrabOrderNum          int64                       `json:"total_grab_order_num"`          // Grab订单数
	MinGrabOrderAmount         float64                     `json:"min_grab_order_amount"`         // Grab最小订单金额
	MaxGrabOrderAmount         float64                     `json:"max_grab_order_amount"`         // Grab最大订单金额
	AvgGrabOrderAmount         float64                     `json:"avg_grab_order_amount"`         // Grab平均订单金额
	AreaList                   []BusinessDataExportArea    `json:"area_list"`                     // 区域列表
	PaymentList                []BusinessDataExportPayment `json:"payment_list"`                  // 支付列表
}

// 营业数据 - 导出 - 区域
type BusinessDataExportArea struct {
	AreaID             int64   `json:"area_id"`              // 区域id
	AreaName           string  `json:"area_name"`            // 区域名称
	AreaSaleAmount     float64 `json:"area_sale_amount"`     // 区域销售额
	AreaBusinessAmount float64 `json:"area_business_amount"` // 区域营业收入
	AreaProductNum     float64 `json:"area_product_num"`     // 区域商品数量
}

// 营业数据 - 导出 - 支付
type BusinessDataExportPayment struct {
	ID                 uint64  `json:"id"`                   // 支付方式ID
	Sort               int     `json:"sort"`                 // 支付方式排序
	CreateTime         int64   `json:"create_time"`          // 支付方式创建时间
	PaymentName        string  `json:"payment_name"`         // 支付名称
	PaymentCode        int     `json:"payment_code"`         // 支付编码
	TotalOrderNum      int64   `json:"total_order_num"`      // 总订单数
	TotalPaymentAmount float64 `json:"total_payment_amount"` // 总支付金额
}

// 营业数据 - 班次退款金额
type BusinessDataShiftRefundAmount struct {
	RefundAmount float64 `json:"refund_amount"` // 退款金额
}

// 营业数据 - 全部
type BusinessDataHome struct {
	TotalReceivedPrice     float64    `json:"total_received_price"`      // 总实收金额
	TotalUserDiscountMoney float64    `json:"total_user_discount_money"` // 总会员折扣
	TotalDiscountMoney     float64    `json:"total_discount_money"`      // 总优惠折扣
	TotalRefundMoney       float64    `json:"total_refund_money"`        // 总退款金额
	TotalOrderNum          int        `json:"total_order_num"`           // 总订单数
	MemberData             MemberData `json:"member_data"`               // 会员数据
}

// 营业数据 - 后厨效率分析
type BusinessDataKitchenEfficiencyAnalysis struct {
	Meta dto.PageResponse                `json:"meta"` // 元数据
	List []KitchenEfficiencyAnalysisItem `json:"list"` // 后厨效率分析
}

// 营业数据 - 后厨效率分析 - 单条
type KitchenEfficiencyAnalysisItem struct {
	ProductPackageUuid uint64             `json:"product_package_uuid"` // 商品包UUID
	ProductName        dto.LocaleResponse `json:"product_name"`         // 商品名称
	CategoryName       dto.LocaleResponse `json:"category_name"`        // 分类名称
	Min                float64            `json:"min"`                  // 最短出品时长
	Max                float64            `json:"max"`                  // 最长出品时长
	Avg                float64            `json:"avg"`                  // 平均出品时长

	exist bool `json:"exist"` // 是否存在
	index int  `json:"index"` // 排序
}

func (item *KitchenEfficiencyAnalysisItem) SetIndex(index int) {
	item.index = index
}

func (item *KitchenEfficiencyAnalysisItem) GetIndex() int {
	return item.index
}

func (item *KitchenEfficiencyAnalysisItem) SetExist(exist bool) {
	item.exist = exist
}

func (item *KitchenEfficiencyAnalysisItem) GetExist() bool {
	return item.exist
}

type BusinessDataKitchenEfficiencyAnalysisAvg struct {
	Avg float64 `json:"avg"` // 平均出品时长,单位:秒
}

type KitchenProductionDetail struct {
	Meta dto.PageResponse              `json:"meta"` // 元数据
	List []KitchenProductionDetailItem `json:"list"` // 后厨菜品出品明细
}

// 后厨菜品出品明细 - 单条
type KitchenProductionDetailItem struct {
	ProductName    dto.LocaleResponse `json:"product_name"`     // 商品名称
	FlavorName     dto.LocaleResponse `json:"flavor_name"`      // 规格名称
	CategoryName   dto.LocaleResponse `json:"category_name"`    // 分类名称
	Number         float64            `json:"number"`           // 完成数量
	CreateTime     int64              `json:"create_time"`      // 下单时间
	MakeFinishTime int64              `json:"make_finish_time"` // 制作完成时间
	MakeDuration   int64              `json:"make_duration"`    // 制作时长,单位:秒
	SendFinishTime int64              `json:"send_finish_time"` // 传菜完成时间
	SendDuration   int64              `json:"send_duration"`    // 传菜时长,单位:秒
	FinishTime     int64              `json:"finish_time"`      // 完成时间
	AllDuration    int64              `json:"all_duration"`     // 总时长,单位:秒
}

// 营业数据 - 时段营业统计
type BusinessTimePeriod struct {
	Meta dto.PageResponse         `json:"meta"` // 元数据
	List []BusinessTimePeriodItem `json:"list"` // 时段营业统计
}

// 营业数据 - 时段营业统计 - 单条
type BusinessTimePeriodItem struct {
	TimePeriod         string  `json:"time_period"`           // 时段, 00:00-01:00
	OrderAmount        float64 `json:"order_amount"`          // 订单金额
	OrderAmountMealAvg float64 `json:"order_meal_avg_amount"` // 订单金额人均
	PayAmount          float64 `json:"pay_amount"`            // 实付金额
	PayAmountMealAvg   float64 `json:"pay_amount_meal_avg"`   // 实付金额人均
	OrderNum           int64   `json:"order_num"`             // 订单数量
	MealNum            int64   `json:"meal_num"`              // 用餐人数
}

// 营业数据 - 综合运营统计
type StatisticsSummary struct {
	Meta dto.PageResponse        `json:"meta"` // 元数据
	List []StatisticsSummaryItem `json:"list"` // 综合运营统计
}

// 营业数据 - 综合运营统计 - 单条
type StatisticsSummaryItem struct {
	Date               string  `json:"date"`                  // 日期
	OrderAmount        float64 `json:"order_amount"`          // 订单金额
	PayAmount          float64 `json:"pay_amount"`            // 实付金额
	OrderNum           int64   `json:"order_num"`             // 订单数量
	MealNum            int64   `json:"meal_num"`              // 用餐人数
	DeskNum            int64   `json:"desk_num"`              // 桌台数
	OrderAmountMealAvg float64 `json:"order_meal_avg_amount"` // 订单金额人均
	PayAmountMealAvg   float64 `json:"pay_amount_meal_avg"`   // 实付金额人均
	OrderAmountAvg     float64 `json:"order_amount_avg"`      // 订单金额平均
	PayAmountAvg       float64 `json:"pay_amount_avg"`        // 实付金额平均
	InstantOrderAmount float64 `json:"instant_order_amount"`  // 点餐订单金额
	DeskOrderAmount    float64 `json:"desk_order_amount"`     // 桌台订单金额
	TakeoutOrderAmount float64 `json:"takeout_order_amount"`  // 外送订单金额
}

// 营业数据 - 支付方式统计
type StatisticsPaymentMethod struct {
	Meta dto.PageResponse              `json:"meta"` // 元数据
	List []StatisticsPaymentMethodItem `json:"list"` // 支付方式统计
}

// 营业数据 - 支付方式统计 - 单条
type StatisticsPaymentMethodItem struct {
	Date          string  `json:"date"`           // 日期
	PaymentName   string  `json:"payment_name"`   // 支付方式名称
	PaymentNum    int     `json:"payment_num"`    // 支付次数
	PaymentAmount float64 `json:"payment_amount"` // 支付金额
}
