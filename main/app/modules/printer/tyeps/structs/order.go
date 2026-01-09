package structs

import (
	"ttpos-server-go/app/model"
)

// StatementOrderData 订单数据结构体
type StatementOrderData struct {
	BrandName string                 `json:"brand_name"`
	Store     StatementStoreData     `json:"store"`
	Order     StatementOrderInfoData `json:"order"`
	Invoice   StatementInvoiceData   `json:"invoice"`
}

// StatementStoreData 店铺数据结构体
type StatementStoreData struct {
	Name             string `json:"name"`               // 店铺名称
	StoreCode        string `json:"store_code"`         // 店铺编号
	Address          string `json:"address"`            // 店铺地址
	Phone            string `json:"phone"`              // 店铺电话
	Logo             string `json:"logo"`               // 店铺logo
	Company          string `json:"company"`            // 公司名称
	CompanyAddr      string `json:"company_addr"`       // 公司地址
	CompanyPhone     string `json:"company_phone"`      // 公司电话
	CompanyTaxNumber string `json:"company_tax_number"` // 公司税号
	CashierSn        string `json:"cashier_sn"`         // 收银员编号
	PrinterSn        string `json:"printer_sn"`         // 打印机编号
}

// StatementOrderInfoData 订单信息数据结构体
type StatementOrderInfoData struct {
	Status                 uint                      `json:"status"`                    // 订单状态
	SerialNo               string                    `json:"serial_no"`                 // 桌位编号
	SerialNos              string                    `json:"serial_nos"`                // 桌位编号（带人数）
	OrderNo                string                    `json:"order_no"`                  // 订单号
	MealNum                uint                      `json:"meal_num"`                  // 就餐人数
	Remark                 string                    `json:"remark"`                    // 订单备注
	CashierName            string                    `json:"cashier_name"`              // 收银员名称
	FinishTime             string                    `json:"finish_time"`               // 完成时间
	CreateTime             string                    `json:"create_time"`               // 创建时间
	PayTime                string                    `json:"pay_time"`                  // 支付时间
	UpdateTime             string                    `json:"update_time"`               // 更新时间
	CancelTime             string                    `json:"cancel_time"`               // 取消时间
	RejectReason           string                    `json:"reject_reason"`             // 拒单原因
	Buffets                []StatementBuffetData     `json:"buffets"`                   // 自助餐列表
	Delays                 []StatementDelayData      `json:"delays"`                    // 加钟列表
	Products               []StatementProductData    `json:"products"`                  // 商品列表
	ProductNum             float64                   `json:"product_num"`               // 商品数量
	ProductAmount          string                    `json:"product_amount"`            // 商品金额
	ServiceFee             string                    `json:"service_fee"`               // 服务费
	TaxRate                float64                   `json:"tax_rate"`                  // 税费率
	TaxFeeType             uint                      `json:"tax_fee_type"`              // 税费类型	0-关闭消费税 1-商品未含税 2-商品已含税
	IsContainTax           uint                      `json:"is_contain_tax"`            // 是否包含税
	DiscountFee            string                    `json:"discount_fee"`              // 折扣金额
	DiscountRate           string                    `json:"discount_rate"`             // 折扣率
	MemberDiscountFee      string                    `json:"member_discount_fee"`       // 会员折扣金额
	MemberDiscountRate     float64                   `json:"member_discount_rate"`      // 会员折扣率
	MemberCardDiscountRate float64                   `json:"member_card_discount_rate"` // 会员卡折扣率
	MemberPointsDiscount   float64                   `json:"member_points_discount"`    // 会员积分抵扣金额
	CouponExchangeAmount   float64                   `json:"coupon_exchange_amount"`    // 优惠券抵扣金额
	ActivityAmount         float64                   `json:"activity_amount"`           // 活动抵扣金额
	CheckOutZeroFee        string                    `json:"check_out_zero_fee"`        // 结账抹零金额
	ReturnAmount           string                    `json:"return_amount"`             // 退款金额
	PaymentCommissionFee   float64                   `json:"payment_commission_fee"`    // 支付手续费
	FreeAmount             string                    `json:"free_amount"`               // 免单金额
	ActualReceivePrice     string                    `json:"actual_receive_price"`      // 实际收款金额
	PaymentMethods         []StatementPaymentMethod  `json:"payment_methods"`           // 支付方式列表
	PercentageLists        []StatementPercentageData `json:"percentage_lists"`          // 税费列表
	IsFree                 bool                      `json:"is_free"`                   // 是否免单
	IsMember               bool                      `json:"is_member"`                 // 是否会员
	MemberRemainingBalance string                    `json:"member_remaining_balance"`  // 会员剩余余额
	MemberPoints           float64                   `json:"member_points"`             // 会员积分
	PaymentName            string                    `json:"payment_name"`              // 支付方式名称
	PaymentQrcode          string                    `json:"payment_qrcode"`            // 支付方式二维码
	Barcode                string                    `json:"barcode"`                   // 条形码
	IsOrderSourceTakeout   bool                      `json:"is_order_source_takeout"`   // 是否订单来源为外卖
	// 外卖平台订单特有字段
	Platform                string `json:"platform"`                  // 平台名称（Grab、LINE MAN 等）
	OrderType               string `json:"order_type"`                // 订单类型（DELIVERY、PICKUP 等）
	DeliveryFee             string `json:"delivery_fee"`              // 配送费
	CustomerName            string `json:"customer_name"`             // 顾客姓名
	CustomerPhone           string `json:"customer_phone"`            // 顾客电话
	CustomerAddress         string `json:"customer_address"`          // 顾客地址
	WarningMessage          string `json:"warning_message"`           // 异常提示信息
	WarningMessageSeparator string `json:"warning_message_separator"` // 异常提示信息分隔符
	PaidAmount              string `json:"paid_amount"`               // 实付金额
}

// StatementBuffetData 自助餐数据结构体
type StatementBuffetData struct {
	Name     string `json:"name"`      // 自助餐名称
	PriceNum string `json:"price_num"` // 自助餐价格数量
	Price    string `json:"price"`     // 自助餐价格
	Num      uint   `json:"num"`       // 自助餐数量
	Subtotal string `json:"subtotal"`  // 自助餐小计
	Attrs    string `json:"attrs"`     // 自助餐属性
	Attr     string `json:"attr"`      // 自助餐属性
}

// StatementDelayData 加钟数据结构体
type StatementDelayData struct {
	Name     string `json:"name"`      // 加钟名称
	PriceNum string `json:"price_num"` // 加钟价格数量
	Price    string `json:"price"`     // 加钟价格
	Num      uint   `json:"num"`       // 加钟数量
	Attrs    string `json:"attrs"`     // 加钟属性
	Attr     string `json:"Attr"`      // 加钟属性
	Subtotal string `json:"subtotal"`  // 加钟小计
}

// StatementProductData 商品数据结构体
type StatementProductData struct {
	Name            string                          `json:"name"`              // 商品名称
	PriceNum        string                          `json:"price_num"`         // 商品价格数量
	Price           string                          `json:"price"`             // 商品价格
	Num             float64                         `json:"num"`               // 商品数量
	Subtotal        string                          `json:"subtotal"`          // 商品小计
	Remark          string                          `json:"remark"`            // 商品备注
	Attrs           string                          `json:"attrs"`             // 商品属性-包含小料-全部
	Attr            string                          `json:"attr"`              // 商品属性
	AttrList        []StatementProductDataAttrList  `json:"attr_list"`         // 商品属性列表
	FlavorName      string                          `json:"flavor_name"`       // 商品规格
	SauceList       []StatementProductDataSauceList `json:"sauce_list"`        // 商品小料列表
	SauceNames      string                          `json:"sauce_names"`       // 商品小料名称
	IsDelay         bool                            `json:"is_delay"`          // 是否加钟
	IsBuffet        bool                            `json:"is_buffet"`         // 是否自助餐
	IsBuffetProduct bool                            `json:"is_buffet_product"` // 是否自助餐商品
	IsGift          bool                            `json:"is_gift"`           // 是否赠品
	IsPackage       bool                            `json:"is_package"`        // 是否打包
	IsSubProduct    bool                            `json:"is_sub_product"`    // 是否套餐子商品
}

// StatementProductDataSauceList
type StatementProductDataAttrList struct {
	Name string `json:"name"` // 名称
	Text string `json:"text"` // 值
}

// StatementProductDataSauceList
type StatementProductDataSauceList struct {
	Name string `json:"name"` // 名称
	Text string `json:"text"` // 值
}

// StatementPaymentMethod 支付方式数据结构体
type StatementPaymentMethod struct {
	Name string `json:"name"` // 支付方式名称
	Text string `json:"text"` // 支付方式值
}

// StatementPercentageData 百分比数据结构体
type StatementPercentageData struct {
	TaxRate    string `json:"tax_rate"`    // 税率
	TaxFee     string `json:"tax_fee"`     // 税费
	TotalPrice string `json:"total_price"` // 总价
}

// StatementInvoiceData 发票数据结构体
type StatementInvoiceData struct {
	InvoiceNumber     string `json:"invoice_number"`      // 发票编号
	UnifiedCreditCode string `json:"unified_credit_code"` // 统一信用代码
	CompanyName       string `json:"company_name"`        // 公司名称
	CompanyAddress    string `json:"company_address"`     // 公司地址
	CompanyPhone      string `json:"company_phone"`       // 公司电话
	CompanyTaxNumber  string `json:"company_tax_number"`  // 公司税号
	PrintNum          int    `json:"print_num"`           // 打印次数
}

// GetStatementProductData 获取商品数据
func GetStatementProductData(saleOrder *model.SaleOrder) []StatementProductData {
	products := []StatementProductData{}
	for _, orderBuffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
		if orderBuffetCustomer.IsDelete() {
			continue
		}
	}
	for _, delay := range saleOrder.SaleOrderBuffetDelayProducts {
		if delay.IsDelete() {
			continue
		}
	}
	for _, product := range saleOrder.SaleOrderProducts {
		if product.IsDelete() {
			continue
		}
	}
	return products
}
