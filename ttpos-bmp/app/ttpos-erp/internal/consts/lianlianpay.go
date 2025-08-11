package consts

type LianlianPayType string

const (
	LianlianPayWechat LianlianPayType = "WeChat Pay"
	LianlianPayAlipay LianlianPayType = "Alipay"
	LianlianPayQr     LianlianPayType = "QR PromptPay"
)

type PaySource int

const (
	LianlianPaySource PaySource = 2
)
