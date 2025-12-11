package resp

import "ttpos-server-go/app/dto"

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

// PaymentMethodListItemResp 支付方式列表项响应（列表用）
type PaymentMethodListItemResp struct {
	Uuid   uint64 `json:"uuid"`   // 支付方式UUID
	Name   string `json:"name"`   // 支付方式名称
	Source int    `json:"source"` // 来源 0-系统 1-手动 2-LianLianPay
	Status int    `json:"status"` // 状态 0-禁用 1-启用 2-草稿
	Sort   int    `json:"sort"`   // 排序
}

// PaymentMethodDetailResp 支付方式详情响应（详情用）
type PaymentMethodDetailResp struct {
	Uuid                 uint64  `json:"uuid"`                    // 支付方式UUID
	Name                 string  `json:"name"`                    // 支付方式名称
	PaymentName          string  `json:"payment_name"`            // 支付名称
	Source               int     `json:"source"`                  // 来源 0-系统 1-手动 2-LianLianPay
	LogoFileUuid         uint64  `json:"logo_file_uuid"`          // Logo图片UUID
	LogoFile             string  `json:"logo_file"`               // Logo文件URL
	QrcodeFileUuid       uint64  `json:"qrcode_file_uuid"`        // 二维码图片UUID
	QrcodeFile           string  `json:"qrcode_file"`             // 二维码文件URL
	FeePercent           float64 `json:"fee_percent"`             // 手续费百分比
	IsShowCashier        int     `json:"is_show_cashier"`         // 0-不显示 1-收银机结账显示
	IsShowAssistant      int     `json:"is_show_assistant"`       // 0-不显示 1-点餐助手结账显示
	IsShowMemberRecharge int     `json:"is_show_member_recharge"` // 0-不显示 1-收银机会员充值显示
	Status               int     `json:"status"`                  // 状态 0-禁用 1-启用 2-草稿
	Sort                 int     `json:"sort"`                    // 排序
}

// PaymentMethodListResp 支付方式列表响应
type PaymentMethodListResp struct {
	List []*PaymentMethodListItemResp `json:"list"` // 支付方式列表
	Meta *dto.PageResponse            `json:"meta"` // 分页信息
}

// LianlianPayConfigResp LianlianPay 配置响应
type LianlianPayConfigResp struct {
	LlWhiteIp            string `json:"ll_white_ip"`             // 白名单IP
	LlMerchantId         string `json:"ll_merchant_id"`          // 商户号
	LlStoreId            string `json:"ll_store_id"`             // 站点ID
	LlPublicKey          string `json:"ll_public_key"`           // 支付方式公钥
	LlMerchantPrivateKey string `json:"ll_merchant_private_key"` // 商户私钥（暂不支持加密存储）
	LlToken              string `json:"ll_token"`                // Token（暂不支持加密存储）
}
