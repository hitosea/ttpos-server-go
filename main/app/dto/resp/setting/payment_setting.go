package setting

// Payment 支付方式
type Payment struct {
	IsCash    string `json:"is_cash"`    // 是否开启现金支付 0-关闭 1-开启
	IsBalance string `json:"is_balance"` // 是否开启余额支付 0-关闭 1-开启
	IsOther   string `json:"is_other"`   // 是否开启其他方式支付 0-关闭 1-开启
}

// PaymentMethodListResp 支付方式列表响应
type PaymentMethodListResp struct {
	List []PaymentMethod `json:"list"`
}

// PaymentMethod 支付方式
type PaymentMethod struct {
	Uuid        uint64 `json:"uuid"`         // 支付方式UUID
	Name        string `json:"name"`         // 支付方式名称
	PaymentName string `json:"payment_name"` // 支付名称
}
