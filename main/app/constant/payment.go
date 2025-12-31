package constant

// 支付方式
const (
	PaymentMethodCodeFreePay             = -1    // 免单
	PaymentMethodCodeBalance             = 10    // 余额支付
	PaymentMethodCodeCash                = 40    // 现金支付
	PaymentMethodCodeWechat              = 20    // 微信支付
	PaymentMethodCodeAliPay              = 30    // 支付宝支付
	PaymentMethodCodeOWechat             = 50    // 自有微信
	PaymentMethodCodeOAliPay             = 60    // 自有支付宝
	PaymentMethodCodePOS                 = 70    // 自有POS刷卡
	PaymentMethodCodeQRPromptPay         = 80    // QR PromptPay
	PaymentMethodCodeQRCode              = 90    // QR code
	PaymentMethodCodeSCBEasy             = 100   // SCB easy
	PaymentMethodCodeKrungthaiNext       = 110   // Krungthai NEXT
	PaymentMethodCodeKrungsriMobile      = 120   // Krungsri Mobile
	PaymentMethodCodeCrossBorderQR       = 130   // Cross-Border QR
	PaymentMethodCodeTrueMoneyWallet     = 140   // TrueMoneyWallet
	PaymentMethodCodeLINEPay             = 150   // LINE Pay
	PaymentMethodCodeJACreditCard        = 160   // ja  credit card
	PaymentMethodCodeJAICTrafficCard     = 170   // ja  credit card
	PaymentMethodCodeJAQRCode            = 180   // JA QRCODE
	PaymentMethodCodeJAQCreditDebit      = 190   // JA QRCODE
	PaymentMethodCodeLianLianWechatPay   = 90111 // LianLianWechatPay
	PaymentMethodCodeLianLianAliPay      = 90222 // LianLianAliPay
	PaymentMethodCodeLianLianQRPromptPay = 90333 // LianLianQRPromptPay
	PaymentMethodCodeGrab                = 91100 // Grab
	PaymentMethodCodeLineMan             = 91200 // LINE MAN
	PaymentMethodCodeFreeMealForErp      = 92000 // Free Meal for ERP（用于ERP同步的免单支付方式）
	PaymentMethodCodeKbankAlipay         = 93000 // Alipay（Kbank）
	PaymentMethodCodeKbankWechat         = 93100 // WeChatPay（Kbank）
	PaymentMethodCodeKbankCreditQR       = 93200 // Credit QR（Kbank）
	PaymentMethodCodeKbankThaiQR         = 93300 // Thai QR（Kbank）
	PaymentMethodCodeKbankCreditCard     = 93400 // Credit Card（Kbank）
)

// 支付方式名称常量
const (
	PaymentMethodNameGrab    = "Grab"     // Grab 支付方式名称
	PaymentMethodNameLineMan = "LINE MAN" // LINE MAN 支付方式名称
)

const (
	PaymentMethodStatusDisable = 0 // 禁用
	PaymentMethodStatusEnable  = 1 // 启用
	PaymentMethodStatusDraft   = 2 // 草稿
)

const (
	PaymentMethodSourceSystem      = 0 // 系统默认
	PaymentMethodSourceDefault     = 1 // 自行添加
	PaymentMethodSourceLianLianPay = 2 // LianLianPay
	PaymentMethodSourceKbank       = 3 // Kbank
)

var PaymentMethodSourceTextMap = map[int]string{
	PaymentMethodSourceSystem:      "系统默认",
	PaymentMethodSourceDefault:     "自行添加",
	PaymentMethodSourceLianLianPay: "LianLianPay",
	PaymentMethodSourceKbank:       "Kbank",
}

const (
	PaymentOrderStatusUnPay  = 0 // 支付订单未支付
	PaymentOrderStatusPaid   = 1 // 支付订单已支付
	PaymentOrderStatusRefund = 2 // 支付订单已退款
	PaymentOrderStatusFailed = 3 // 支付订单支付失败
)

// 支付订单关联的订单类型
const (
	PaymentOrderRelatedTypeSaleOrder     = 0 // 销售订单
	PaymentOrderRelatedTypeRechargeOrder = 1 // 充值订单
)

const (
	PaymentMethodShowAll      = "all"
	PaymentMethodShowCheckout = "checkout"
	PaymentMethodShowRecharge = "recharge"
)
