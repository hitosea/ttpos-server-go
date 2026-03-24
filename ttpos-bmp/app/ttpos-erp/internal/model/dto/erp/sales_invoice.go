package erp

// SalesInvoice ERPNext Sales Invoice API 结构
type SalesInvoice struct {
	PosProfile                   string                      `json:"pos_profile,omitempty"`
	Company                      string                      `json:"company"`
	Customer                     string                      `json:"customer"`
	IsPos                        int                         `json:"is_pos"`
	UpdateStock                  int                         `json:"update_stock"` // 0=不扣库存
	PostingDate                  string                      `json:"posting_date"`
	PostingTime                  string                      `json:"posting_time,omitempty"`
	SetPostingTime               int                         `json:"set_posting_time,omitempty"`
	Currency                     string                      `json:"currency"`
	ConversionRate               float64                     `json:"conversion_rate,omitempty"`
	PriceListCurrency            string                      `json:"price_list_currency,omitempty"`
	PlcConversionRate            float64                     `json:"plc_conversion_rate,omitempty"`
	SellingPriceList             string                      `json:"selling_price_list,omitempty"`
	Branch                       string                      `json:"branch,omitempty"`
	Items                        []SalesInvoiceItem          `json:"items"`
	Taxes                        []SalesInvoiceTax           `json:"taxes,omitempty"`
	Payments                     []SalesInvoicePaymentDetail `json:"payments,omitempty"`
	AdditionalDiscountPercentage float64                     `json:"additional_discount_percentage,omitempty"`
	DiscountAmount               float64                     `json:"discount_amount,omitempty"`
	Docstatus                    int                         `json:"docstatus,omitempty"`
	// Customer's Purchase Order（在 ERPNext 中显示为订单号）
	PoNo string `json:"po_no,omitempty"`
	// 自定义字段
	TtposTakeoutOrderNo  string `json:"custom_takeout_order_no,omitempty"`
	TtposTakeoutProvider string `json:"custom_takeout_provider,omitempty"`
	// 反结账
	AmendedFrom string `json:"amended_from,omitempty"`
}

// SalesInvoiceItem Sales Invoice 明细行
type SalesInvoiceItem struct {
	ItemCode           string  `json:"item_code"`
	Qty                float64 `json:"qty"`
	Rate               float64 `json:"rate"`
	Amount             float64 `json:"amount"`
	Uom                string  `json:"uom,omitempty"`
	Description        string  `json:"description,omitempty"`
	IsFreeItem         int     `json:"is_free_item,omitempty"`
	DiscountPercentage float64 `json:"discount_percentage,omitempty"` // 行级折扣百分比，free item 需设为 100
	Warehouse          string  `json:"warehouse,omitempty"`           // 行级仓库
}

// SalesInvoiceTax Sales Invoice 税费
type SalesInvoiceTax struct {
	ChargeType          string  `json:"charge_type"`
	AccountHead         string  `json:"account_head,omitempty"`
	Description         string  `json:"description"`
	Rate                float64 `json:"rate,omitempty"`
	TaxAmount           float64 `json:"tax_amount"`
	IncludedInPrintRate int     `json:"included_in_print_rate,omitempty"`
}

// SalesInvoicePaymentDetail Sales Invoice 支付信息（is_pos=1 时用）
type SalesInvoicePaymentDetail struct {
	ModeOfPayment string  `json:"mode_of_payment"`
	Amount        float64 `json:"amount"`
}

// SalesInvoiceResp ERPNext Sales Invoice 创建响应
type SalesInvoiceResp struct {
	Data struct {
		Name      string `json:"name"`
		Docstatus int    `json:"docstatus"`
	} `json:"data"`
}
