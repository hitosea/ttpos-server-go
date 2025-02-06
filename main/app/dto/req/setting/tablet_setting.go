package setting

// Tablet 平板端设置
type Tablet struct {
	Carousel           []CarouselItem   `json:"carousel"`              // 上传后的轮播内容url（图片 + 视频）
	IsCallService      string           `json:"is_call_service"`       // 是否开启呼叫服务员 0-关闭 1-开启
	IsCustomerOrder    string           `json:"is_customer_order"`     // 是否开启顾客自助下单 0-关闭 1-开启
	IsVoiceRemind      string           `json:"is_voice_remind"`       // 是否开启声音提醒 0-关闭 1-开启
	IsShowSoldOut      int              `json:"is_show_sold_out"`      // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	IsBuffetOrderLimit string           `json:"is_buffet_order_limit"` // 是否开启自助餐下单限制 0-关闭 1-开启
	BuffetOrderLimit   BuffetOrderLimit `json:"buffet_order_limit"`    // 自助餐下单限制
	IsOrderLimit       string           `json:"is_order_limit"`        // 是否开启非自助餐下单限制 0-关闭 1-开启
	OrderLimit         OrderLimit       `json:"order_limit"`           // 非自助餐下单限制
	Server             Server           `json:"server"`                // 平板服务器连接
	AdvancedPassword   string           `json:"advanced_password"`     // 高级设置密码
	LanguageList       []LanguageItem   `json:"language_list"`         // 语言列表
	Language           []string         `json:"language"`              // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage    string           `json:"default_language"`      // 默认语言
	KitchenLanguage    string           `json:"kitchen_language"`
}
