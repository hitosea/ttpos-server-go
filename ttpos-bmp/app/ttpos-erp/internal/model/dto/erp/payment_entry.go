package erp

// PaymentEntry ERPNext Payment Entry API 结构
type PaymentEntry struct {
	PaymentType   string             `json:"payment_type"`   // "Receive" (收款) / "Pay" (退款)
	PartyType     string             `json:"party_type"`     // "Customer"
	Party         string             `json:"party"`          // 客户名称
	Company       string             `json:"company"`
	ModeOfPayment string             `json:"mode_of_payment"`
	PaidAmount    float64            `json:"paid_amount"`
	PaidFrom      string             `json:"paid_from,omitempty"`
	PaidTo        string             `json:"paid_to,omitempty"`
	PaidFromAccountCurrency string   `json:"paid_from_account_currency,omitempty"`
	PaidToAccountCurrency   string   `json:"paid_to_account_currency,omitempty"`
	ReceivedAmount         float64            `json:"received_amount,omitempty"`
	SourceExchangeRate     float64            `json:"source_exchange_rate,omitempty"`
	TargetExchangeRate     float64            `json:"target_exchange_rate,omitempty"`
	References             []PaymentReference `json:"references"`
	PostingDate   string             `json:"posting_date,omitempty"`
	Docstatus     int                `json:"docstatus,omitempty"`
	// 自定义字段
	TtposSaleOrderUuid string `json:"custom_ttpos_sale_order_uuid,omitempty"`
}

// PaymentReference Payment Entry 关联单据
type PaymentReference struct {
	ReferenceDoctype string  `json:"reference_doctype"` // "Sales Invoice"
	ReferenceName    string  `json:"reference_name"`    // SI/Credit Note 名称
	AllocatedAmount  float64 `json:"allocated_amount"`
}

// PaymentEntryResp ERPNext Payment Entry 创建响应
type PaymentEntryResp struct {
	Data struct {
		Name      string `json:"name"`
		Docstatus int    `json:"docstatus"`
	} `json:"data"`
}

// CreditNote ERPNext Credit Note (退款 SI) 结构
type CreditNote struct {
	IsReturn      int                 `json:"is_return"`       // 1
	ReturnAgainst string              `json:"return_against"`  // 原 SI 名称
	Customer      string              `json:"customer"`
	Company       string              `json:"company"`
	PostingDate   string              `json:"posting_date"`
	UpdateStock   int                 `json:"update_stock"`    // 0（不回增库存）
	Items         []SalesInvoiceItem  `json:"items"`
	Taxes         []SalesInvoiceTax   `json:"taxes,omitempty"`
	// 自定义字段
	TtposSaleOrderUuid string `json:"custom_ttpos_sale_order_uuid,omitempty"`
	TtposRefundType    string `json:"custom_ttpos_refund_type,omitempty"` // full_refund/partial_refund
}
