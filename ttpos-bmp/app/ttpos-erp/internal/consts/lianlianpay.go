package consts

type LianlianPayType string

// LianlianPayPrefix 连连支付前缀
const LianlianPayPrefix = "LianlianPay-"

const (
	LianlianPayWechat LianlianPayType = "WeChat Pay"
	LianlianPayAlipay LianlianPayType = "Alipay"
	LianlianPayQr     LianlianPayType = "QR PromptPay"
)

type PaySource int

const (
	LianlianPaySource PaySource = 2
)

type ModeOfPayment string

const (
	ModeOfPaymentCash    ModeOfPayment = "Cash"
	ModeOfPaymentBalance ModeOfPayment = "Balance"
)

type ItemGroup string

const (
	// ItemGroupProducts 商品
	ItemGroupProducts ItemGroup = "Products"
	// ItemGroupRawMaterial 原材料
	ItemGroupRawMaterial ItemGroup = "Raw Material"
)
