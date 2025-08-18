package buying

type CreatePurchaseFromMqReq struct {
	SourceName string `json:"source_name,omitempty"`
	Supplier   string `json:"supplier,omitempty"`
}
