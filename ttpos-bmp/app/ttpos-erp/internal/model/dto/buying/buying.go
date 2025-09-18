package buying

type CreatePurchaseFromMqReq struct {
	SourceName string `json:"source_name,omitempty"`
	Supplier   string `json:"supplier,omitempty"`
	RequiredBy string `json:"required_by,omitempty"`
}

type CreateInnerSaleOrderFromPurchaseOrderReq struct {
	SourceName   string `json:"source_name,omitempty"`
	DeliveryDate string `json:"delivery_date,omitempty"`
}

type AddSupplerTransactCompanyReq struct {
	Supplier        string `json:"supplier"`
	WithCompanyAbbr string `json:"companyAbbr"`
}
