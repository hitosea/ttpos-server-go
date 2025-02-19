package resp

type PaymentMethodItem struct {
	SourceText  string  `json:"source_text"`
	Uuid        uint64  `json:"uuid"`
	PaymentName string  `json:"payment_name"`
	FeePercent  float64 `json:"fee_percent"`
	Logo        string  `json:"logo"`
	Qrcode      string  `json:"qrcode"`
}

type PaymentMethodList struct {
	List []PaymentMethodItem
}
