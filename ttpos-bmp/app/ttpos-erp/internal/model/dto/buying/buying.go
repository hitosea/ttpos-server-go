package buying

type CreatePurchaseFromMqReq struct {
	SourceName string `json:"source_name,omitempty"`
	Supplier   string `json:"supplier,omitempty"`
}

type CreateInnerSaleOrderFromPurchaseOrderReq struct {
	SourceName   string `json:"source_name,omitempty"`
	DeliveryDate string `json:"delivery_date,omitempty"`
}
