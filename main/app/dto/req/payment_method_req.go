package req

// PaymentMethodManagementListReq 支付方式管理列表请求
type PaymentMethodManagementListReq struct {
	PageNo   int `form:"page_no" json:"page_no" binding:"required,min=1"`              // 页码
	PageSize int `form:"page_size" json:"page_size" binding:"required,min=1,max=1100"` // 每页大小
}

// PaymentMethodCreateReq 创建支付方式请求（批量创建）
type PaymentMethodCreateReq struct {
	Items []PaymentMethodCreateItem `json:"items" binding:"required,min=1,dive"` // 支付方式列表
}

// PaymentMethodCreateItem 创建支付方式单项
type PaymentMethodCreateItem struct {
	Name                 string  `json:"name" binding:"required"`             // 支付方式名称
	PaymentName          string  `json:"payment_name" binding:"required"`     // 支付名称
	Code                 int     `json:"code"`                                // 支付方式代号（可选，用于Kbank支付方式）
	Source               int     `json:"source"`                              // 来源（可选，默认1，Kbank支付方式为3）
	LogoFileUuid         uint64  `json:"logo_file_uuid"`                      // Logo图片UUID
	QrcodeFileUuid       uint64  `json:"qrcode_file_uuid"`                    // 二维码图片UUID
	DefaultImg           string  `json:"default_img"`                         // 默认图片
	FeePercent           float64 `json:"fee_percent" binding:"gte=0,lte=100"` // 手续费百分比，取值范围0-100
	IsShowCashier        int     `json:"is_show_cashier"`                     // 0-不显示 1-收银机结账显示
	IsShowAssistant      int     `json:"is_show_assistant"`                   // 0-不显示 1-点餐助手结账显示
	IsShowKiosk          int     `json:"is_show_kiosk"`                       // 0-不显示 1-自助点餐机结账显示
	IsShowMemberRecharge int     `json:"is_show_member_recharge"`             // 0-不显示 1-收银机会员充值显示
	Status               int     `json:"status"`                              // 状态 0-禁用 1-启用 2-草稿
}

// PaymentMethodUpdateReq 更新支付方式请求
type PaymentMethodUpdateReq struct {
	Uuid                 uint64  `json:"uuid" binding:"required"`                       // 支付方式UUID
	Name                 string  `json:"name"`                                          // 支付方式名称
	LogoFileUuid         uint64  `json:"logo_file_uuid"`                                // Logo图片UUID
	QrcodeFileUuid       uint64  `json:"qrcode_file_uuid"`                              // 二维码图片UUID
	FeePercent           float64 `json:"fee_percent" binding:"omitempty,gte=0,lte=100"` // 手续费百分比，取值范围0-100
	IsShowCashier        int     `json:"is_show_cashier"`                               // 0-不显示 1-收银机结账显示
	IsShowAssistant      int     `json:"is_show_assistant"`                             // 0-不显示 1-点餐助手结账显示
	IsShowKiosk          int     `json:"is_show_kiosk"`                                 // 0-不显示 1-自助点餐机结账显示
	IsShowMemberRecharge int     `json:"is_show_member_recharge"`                       // 0-不显示 1-收银机会员充值显示
	Status               int     `json:"status"`                                        // 状态 0-禁用 1-启用 2-草稿
}

// PaymentMethodGetReq 获取支付方式详情请求
type PaymentMethodGetReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 支付方式UUID
}

// PaymentMethodSortUpdateReq 排序更新请求（批量）
type PaymentMethodSortUpdateReq struct {
	Items []PaymentMethodSortItem `json:"items" binding:"required,min=1"` // 排序项列表
}

// PaymentMethodSortItem 支付方式排序项
type PaymentMethodSortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"`       // 支付方式UUID
	Sort int    `json:"sort" binding:"required,min=1"` // 排序值
}

// LianlianPayConfigGetReq LianlianPay 配置查询请求
// 无参数，按当前公司查询
type LianlianPayConfigGetReq struct {
}

// LianlianPayConfigUpdateReq LianlianPay 配置更新请求
type LianlianPayConfigUpdateReq struct {
	LlWhiteIp            string `json:"ll_white_ip" binding:"required"`             // 白名单IP
	LlMerchantId         string `json:"ll_merchant_id" binding:"required"`          // 商户号
	LlStoreId            string `json:"ll_store_id" binding:"required"`             // 站点ID
	LlPublicKey          string `json:"ll_public_key" binding:"required"`           // 支付方式公钥
	LlMerchantPrivateKey string `json:"ll_merchant_private_key" binding:"required"` // 商户私钥
	LlToken              string `json:"ll_token" binding:"required"`                // Token
}
