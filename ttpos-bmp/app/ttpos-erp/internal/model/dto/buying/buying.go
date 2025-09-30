package buying

type CreatePurchaseFromMqReq struct {
	SourceName      string `json:"source_name,omitempty"`
	Supplier        string `json:"supplier,omitempty"`
	RequiredBy      string `json:"required_by,omitempty"`
	TargetWarehouse string `json:"target_warehouse,omitempty"`
}

type CreateInnerSaleOrderFromPurchaseOrderReq struct {
	SourceName      string `json:"source_name,omitempty"`
	DeliveryDate    string `json:"delivery_date,omitempty"`
	SourceWarehouse string `json:"source_warehouse,omitempty"`
}

type AddSupplerTransactCompanyReq struct {
	Supplier        string `json:"supplier"`
	WithCompanyAbbr string `json:"companyAbbr"`
}

type CreateDeliveryNoteFromInnerSaleOrderReq struct {
	SourceName      string `json:"source_name,omitempty"`
	SourceWarehouse string `json:"source_warehouse,omitempty"`
	TargetWarehouse string `json:"target_warehouse,omitempty"`
}
