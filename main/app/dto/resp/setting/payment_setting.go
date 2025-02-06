package setting

// Payment 支付方式
type Payment struct {
	IsCash    string `json:"is_cash"`    // 是否开启现金支付 0-关闭 1-开启
	IsBalance string `json:"is_balance"` // 是否开启余额支付 0-关闭 1-开启
	IsOther   string `json:"is_other"`   // 是否开启其他方式支付 0-关闭 1-开启
}
