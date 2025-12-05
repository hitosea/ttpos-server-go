package resp

type PaymentMethodItem struct {
	Source        int     `json:"source"`         // 来源
	SourceText    string  `json:"source_text"`    // 来源文案
	Uuid          uint64  `json:"uuid"`           // 支付方式uuid
	PaymentName   string  `json:"payment_name"`   // 支付名称
	PaymentMethod string  `json:"payment_method"` // 支付方式
	FeePercent    float64 `json:"fee_percent"`    // 手续费率
	Logo          string  `json:"logo"`           // logo
	Qrcode        string  `json:"qrcode"`         // 二维码
	IsAvailable   bool    `json:"is_available"`   // 是否可用
	Code          int     `json:"code"`           // 代号: -1 免单; 10 余额支付; 40 现金支付; 20 微信支付; 30 支付宝支付; 50 自有微信; 60 自有支付宝; 70 自有POS刷卡; 80 QR PromptPay; 90 QR code; 100 SCB easy; 110 Krungthai NEXT; 120 Krungsri Mobile; 130 Cross-Border QR; 140 TrueMoneyWallet; 150 LINE Pay; 160 ja  credit card; 170 ja  credit card; 180 JA QRCODE; 190 JA QRCODE; 90111 LianLianWechatPay; 90222 LianLianAliPay; 90333 LianLianQRPromptPay;
}

type PaymentMethodList struct {
	List []PaymentMethodItem `json:"list"`
}
